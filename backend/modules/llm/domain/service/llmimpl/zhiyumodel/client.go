// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package zhiyumodel

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/pkg/errors"

	"github.com/coze-dev/coze-loop/backend/modules/llm/domain/entity"
)

// Config 智谕模型配置
type Config struct {
	// 基础配置
	BaseURL string
	Model   string
	Timeout time.Duration

	// Header 认证参数
	AIAPICode   string
	AIAppKey    string
	CallerToken string
	Description string

	// Body 额外参数
	ProcessCode string
	AppID       string
	AccessToken string

	// 模型参数
	MaxTokens        *int
	Temperature      *float32
	TopP             *float32
	TopK             *int32
	Stop             []string
	FrequencyPenalty *float32
	PresencePenalty  *float32
}

// Client 智谕模型客户端
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient 创建智谕模型客户端
func NewClient(cfg *Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	ProcessCode string     `json:"processCode"`
	AppID       string     `json:"appId"`
	AccessToken string     `json:"accessToken"`
	Stream      bool       `json:"stream"`
	Model       string     `json:"model"`
	MaxTokens   int        `json:"max_tokens,omitempty"`
	Temperature *float32   `json:"temperature,omitempty"`
	TopP        *float32   `json:"top_p,omitempty"`
	Messages    []*Message `json:"messages"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []*Choice `json:"choices"`
	Usage   *Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int              `json:"index"`
	Message      *ResponseMessage `json:"message,omitempty"`
	Delta        *ResponseMessage `json:"delta,omitempty"`
	FinishReason *string          `json:"finish_reason"`
}

// ResponseMessage 响应消息
type ResponseMessage struct {
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	Index    int           `json:"index"`
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Function *FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Usage token 使用量
type Usage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

// Chat 非流式聊天
func (c *Client) Chat(ctx context.Context, messages []*entity.Message, opts ...entity.Option) (*entity.Message, error) {
	options := entity.ApplyOptions(nil, opts...)

	req := c.buildRequest(messages, false, options)
	respBody, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp ChatResponse
	if err := sonic.Unmarshal(respBody, &resp); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal response: %s", string(respBody))
	}

	return c.parseResponse(&resp)
}

// ChatStream 流式聊天
func (c *Client) ChatStream(ctx context.Context, messages []*entity.Message, opts ...entity.Option) (entity.IStreamReader, error) {
	options := entity.ApplyOptions(nil, opts...)

	req := c.buildRequest(messages, true, options)
	httpReq, err := c.createHTTPRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		return nil, errors.Errorf("request failed with status %d: %s", httpResp.StatusCode, string(body))
	}

	return NewStreamReader(httpResp.Body), nil
}

func (c *Client) buildRequest(messages []*entity.Message, stream bool, options *entity.Options) *ChatRequest {
	req := &ChatRequest{
		ProcessCode: c.config.ProcessCode,
		AppID:       c.config.AppID,
		AccessToken: c.config.AccessToken,
		Stream:      stream,
		Model:       c.config.Model,
	}

	// 设置模型参数
	if options.MaxTokens != nil {
		req.MaxTokens = *options.MaxTokens
	} else if c.config.MaxTokens != nil {
		req.MaxTokens = *c.config.MaxTokens
	}

	if options.Temperature != nil {
		req.Temperature = options.Temperature
	} else if c.config.Temperature != nil {
		req.Temperature = c.config.Temperature
	}

	if options.TopP != nil {
		req.TopP = options.TopP
	} else if c.config.TopP != nil {
		req.TopP = c.config.TopP
	}

	// 转换消息
	for _, msg := range messages {
		req.Messages = append(req.Messages, &Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return req
}

func (c *Client) createHTTPRequest(ctx context.Context, req *ChatRequest) (*http.Request, error) {
	body, err := sonic.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	url := fmt.Sprintf("%s/v1/chat/completions", strings.TrimSuffix(c.config.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// 设置 Headers
	httpReq.Header.Set("Content-Type", "application/json;charset=UTF-8")
	httpReq.Header.Set("AI-API-CODE", c.config.AIAPICode)
	httpReq.Header.Set("AI-APP-KEY", c.config.AIAppKey)
	httpReq.Header.Set("CALLER-TOKEN", c.config.CallerToken)
	if c.config.Description != "" {
		httpReq.Header.Set("description", c.config.Description)
	}

	return httpReq, nil
}

func (c *Client) doRequest(ctx context.Context, req *ChatRequest) ([]byte, error) {
	httpReq, err := c.createHTTPRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) parseResponse(resp *ChatResponse) (*entity.Message, error) {
	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}

	choice := resp.Choices[0]
	msg := choice.Message
	if msg == nil {
		return nil, errors.New("no message in choice")
	}

	result := &entity.Message{
		Role:    entity.Role(msg.Role),
		Content: msg.Content,
	}

	// 解析 tool calls
	if len(msg.ToolCalls) > 0 {
		for _, tc := range msg.ToolCalls {
			index := int64(tc.Index)
			result.ToolCalls = append(result.ToolCalls, &entity.ToolCall{
				Index: &index,
				ID:    tc.ID,
				Type:  tc.Type,
				Function: &entity.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
	}

	// 设置 token 使用量和完成原因
	if resp.Usage != nil {
		result.ResponseMeta = &entity.ResponseMeta{
			Usage: &entity.TokenUsage{
				PromptTokens:     int(resp.Usage.PromptTokens),
				CompletionTokens: int(resp.Usage.CompletionTokens),
				TotalTokens:      int(resp.Usage.TotalTokens),
			},
		}
	}
	if choice.FinishReason != nil {
		if result.ResponseMeta == nil {
			result.ResponseMeta = &entity.ResponseMeta{}
		}
		result.ResponseMeta.FinishReason = *choice.FinishReason
	}

	return result, nil
}

// StreamReader 流式读取器
type StreamReader struct {
	reader  *bufio.Reader
	body    io.ReadCloser
	done    bool
	lastMsg *entity.Message
}

// NewStreamReader 创建流式读取器
func NewStreamReader(body io.ReadCloser) *StreamReader {
	return &StreamReader{
		reader: bufio.NewReader(body),
		body:   body,
	}
}

// Recv 接收下一个消息
func (s *StreamReader) Recv() (*entity.Message, error) {
	if s.done {
		return nil, io.EOF
	}

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.done = true
				return nil, io.EOF
			}
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			s.done = true
			return nil, io.EOF
		}

		var resp ChatResponse
		if err := sonic.UnmarshalString(data, &resp); err != nil {
			continue
		}

		msg := s.parseStreamResponse(&resp)
		if msg != nil {
			return msg, nil
		}
	}
}

func (s *StreamReader) parseStreamResponse(resp *ChatResponse) *entity.Message {
	if len(resp.Choices) == 0 {
		return nil
	}

	choice := resp.Choices[0]
	delta := choice.Delta
	if delta == nil {
		return nil
	}

	msg := &entity.Message{
		Role:    entity.Role(delta.Role),
		Content: delta.Content,
	}

	// 解析 tool calls
	if len(delta.ToolCalls) > 0 {
		for _, tc := range delta.ToolCalls {
			index := int64(tc.Index)
			msg.ToolCalls = append(msg.ToolCalls, &entity.ToolCall{
				Index: &index,
				ID:    tc.ID,
				Type:  tc.Type,
				Function: &entity.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
	}

	// 设置 token 使用量和完成原因
	if resp.Usage != nil {
		msg.ResponseMeta = &entity.ResponseMeta{
			Usage: &entity.TokenUsage{
				PromptTokens:     int(resp.Usage.PromptTokens),
				CompletionTokens: int(resp.Usage.CompletionTokens),
				TotalTokens:      int(resp.Usage.TotalTokens),
			},
		}
	}
	if choice.FinishReason != nil {
		if msg.ResponseMeta == nil {
			msg.ResponseMeta = &entity.ResponseMeta{}
		}
		msg.ResponseMeta.FinishReason = *choice.FinishReason
	}

	return msg
}

// Close 关闭流
func (s *StreamReader) Close() {
	if s.body != nil {
		s.body.Close()
	}
}
