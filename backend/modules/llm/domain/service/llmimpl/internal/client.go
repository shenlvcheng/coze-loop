// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

//--------------start----------------------
//新增代码人  Claude (AI Assistant)
//新增代码原因：实现内部协议HTTP客户端，支持：
//             1. 自定义认证headers（AI-API-CODE, AI-APP-KEY, CALLER-TOKEN, description）
//             2. 额外请求body字段（processCode, appId, accessToken）
//             3. Function Calling工具调用
//             4. 流式和非流式响应
//             5. OpenAI兼容的响应格式解析

package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	einoModel "github.com/cloudwego/eino/components/model"
	"github.com/pkg/errors"
)

// ChatModelConfig 内部协议聊天模型配置
type ChatModelConfig struct {
	BaseURL     string        // API基础URL
	Model       string        // 模型名称
	Timeout     time.Duration // 超时时间
	MaxTokens   *int          // 最大token数
	Temperature *float64      // 温度参数
	TopP        *float64      // TopP参数
	Stop        []string      // 停止词

	// 内部协议特有的认证和配置
	AIApiCode   string // AI-API-CODE header
	AIAppKey    string // AI-APP-KEY header
	CallerToken string // CALLER-TOKEN header
	Description string // description header
	ProcessCode string // processCode in body
	AppID       string // appId in body
	AccessToken string // accessToken in body
}

// ChatModel 内部协议聊天模型实现
type ChatModel struct {
	config     *ChatModelConfig
	httpClient *http.Client
	tools      []*einoModel.ToolInfo // 绑定的工具
}

// NewChatModel 创建新的内部协议聊天模型实例
func NewChatModel(ctx context.Context, cfg *ChatModelConfig) (einoModel.ToolCallingChatModel, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if cfg.BaseURL == "" {
		return nil, errors.New("base_url is required")
	}
	if cfg.Model == "" {
		return nil, errors.New("model is required")
	}

	return &ChatModel{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		tools: nil,
	}, nil
}

// Generate 非流式生成
func (c *ChatModel) Generate(ctx context.Context, input []*einoModel.Message, opts ...einoModel.Option) (*einoModel.Message, error) {
	// 构建请求
	reqBody := c.buildRequestBody(input, false)
	req, err := c.buildHTTPRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return c.convertToEinoMessage(&result), nil
}

// Stream 流式生成
func (c *ChatModel) Stream(ctx context.Context, input []*einoModel.Message, opts ...einoModel.Option) (einoModel.StreamReader[*einoModel.Message], error) {
	// 构建请求
	reqBody := c.buildRequestBody(input, true)
	req, err := c.buildHTTPRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, errors.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return &streamReader{
		scanner: bufio.NewScanner(resp.Body),
		closer:  resp.Body,
	}, nil
}

// buildRequestBody 构建请求body
func (c *ChatModel) buildRequestBody(input []*einoModel.Message, stream bool) map[string]interface{} {
	messages := make([]map[string]interface{}, 0, len(input))
	for _, msg := range input {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	body := map[string]interface{}{
		"model":       c.config.Model,
		"messages":    messages,
		"stream":      stream,
		"processCode": c.config.ProcessCode,
		"appId":       c.config.AppID,
		"accessToken": c.config.AccessToken,
	}

	if c.config.MaxTokens != nil {
		body["max_tokens"] = *c.config.MaxTokens
	}
	if c.config.Temperature != nil {
		body["temperature"] = *c.config.Temperature
	}
	if c.config.TopP != nil {
		body["top_p"] = *c.config.TopP
	}
	if len(c.config.Stop) > 0 {
		body["stop"] = c.config.Stop
	}

	// 添加工具调用支持
	if len(c.tools) > 0 {
		tools := make([]map[string]interface{}, 0, len(c.tools))
		for _, tool := range c.tools {
			tools = append(tools, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        tool.Name,
					"description": tool.Desc,
					"parameters":  tool.ParamsOneOf.Schemas[0], // 使用第一个schema
				},
			})
		}
		body["tools"] = tools
	}

	return body
}

// buildHTTPRequest 构建HTTP请求
func (c *ChatModel) buildHTTPRequest(ctx context.Context, body map[string]interface{}) (*http.Request, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	url := c.config.BaseURL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// 设置标准headers
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	// 设置内部协议特有的认证headers
	req.Header.Set("AI-API-CODE", c.config.AIApiCode)
	req.Header.Set("AI-APP-KEY", c.config.AIAppKey)
	req.Header.Set("CALLER-TOKEN", c.config.CallerToken)
	req.Header.Set("description", c.config.Description)

	return req, nil
}

// convertToEinoMessage 将OpenAI格式的响应转换为Eino Message
func (c *ChatModel) convertToEinoMessage(resp *openAIResponse) *einoModel.Message {
	if len(resp.Choices) == 0 {
		return &einoModel.Message{
			Role:    einoModel.Assistant,
			Content: "",
		}
	}

	choice := resp.Choices[0]
	msg := &einoModel.Message{
		Role:    einoModel.Assistant,
		Content: choice.Message.Content,
	}

	// 处理工具调用
	if len(choice.Message.ToolCalls) > 0 {
		toolCalls := make([]*einoModel.ToolCall, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			toolCalls = append(toolCalls, &einoModel.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: &einoModel.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
		msg.ToolCalls = toolCalls
	}

	// 设置 ResponseMeta
	if resp.Usage != nil {
		msg.ResponseMeta = &einoModel.ResponseMeta{
			FinishReason: choice.FinishReason,
			Usage: &einoModel.TokenUsage{
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				TotalTokens:      resp.Usage.TotalTokens,
			},
		}
	}

	return msg
}

// BindTools 绑定工具（用于实现ChatModel接口）
func (c *ChatModel) BindTools(tools []*einoModel.ToolInfo) (einoModel.ChatModel, error) {
	c.tools = tools
	return c, nil
}

// WithTools 绑定工具（用于实现ToolCallingChatModel接口）
func (c *ChatModel) WithTools(tools []*einoModel.ToolInfo) (einoModel.ToolCallingChatModel, error) {
	c.tools = tools
	return c, nil
}

// openAIResponse OpenAI格式的响应
type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// streamReader 流式读取器
type streamReader struct {
	scanner *bufio.Scanner
	closer  io.Closer
}

func (s *streamReader) Recv() (*einoModel.Message, error) {
	for s.scanner.Scan() {
		line := s.scanner.Text()
		line = strings.TrimSpace(line)

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查是否是结束标记
		if strings.HasPrefix(line, "data: [DONE]") {
			return nil, io.EOF
		}

		// 解析 SSE 格式
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		jsonStr := strings.TrimPrefix(line, "data: ")
		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		msg := &einoModel.Message{
			Role:    einoModel.Assistant,
			Content: choice.Delta.Content,
		}

		// 处理工具调用
		if len(choice.Delta.ToolCalls) > 0 {
			toolCalls := make([]*einoModel.ToolCall, 0, len(choice.Delta.ToolCalls))
			for _, tc := range choice.Delta.ToolCalls {
				toolCalls = append(toolCalls, &einoModel.ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: &einoModel.ToolCallFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
			msg.ToolCalls = toolCalls
		}

		// 如果有 finish_reason，设置 ResponseMeta
		if choice.FinishReason != "" {
			msg.ResponseMeta = &einoModel.ResponseMeta{
				FinishReason: choice.FinishReason,
			}
		}

		// 如果有 usage 信息
		if chunk.Usage != nil {
			if msg.ResponseMeta == nil {
				msg.ResponseMeta = &einoModel.ResponseMeta{}
			}
			msg.ResponseMeta.Usage = &einoModel.TokenUsage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}

		return msg, nil
	}

	if err := s.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

func (s *streamReader) Close() error {
	return s.closer.Close()
}

// openAIStreamChunk OpenAI流式响应块
type openAIStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string `json:"role,omitempty"`
			Content   string `json:"content,omitempty"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}
//--------------end-----------------------
