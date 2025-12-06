// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package zhiyumodel

import (
	"context"
	"time"

	"github.com/coze-dev/coze-loop/backend/modules/llm/domain/entity"
	"github.com/coze-dev/coze-loop/backend/modules/llm/domain/service/llminterface"
)

// LLM 智谕模型 LLM 实现
type LLM struct {
	client *Client
}

var _ llminterface.ILLM = (*LLM)(nil)

// NewLLM 创建智谕模型 LLM
func NewLLM(ctx context.Context, model *entity.Model, opts ...entity.Option) (*LLM, error) {
	p := model.ProtocolConfig
	options := entity.ApplyOptions(nil, opts...)

	cfg := &Config{
		BaseURL: p.BaseURL,
		Model:   p.Model,
	}

	// 设置超时
	if p.TimeoutMs != nil {
		cfg.Timeout = time.Duration(*p.TimeoutMs) * time.Millisecond
	}

	// 设置智谕模型特有配置
	if zhiyuCfg := p.ProtocolConfigZhiyuModel; zhiyuCfg != nil {
		cfg.AIAPICode = zhiyuCfg.AIAPICode
		cfg.AIAppKey = zhiyuCfg.AIAppKey
		cfg.CallerToken = zhiyuCfg.CallerToken
		cfg.Description = zhiyuCfg.Description
		cfg.ProcessCode = zhiyuCfg.ProcessCode
		cfg.AppID = zhiyuCfg.AppID
		cfg.AccessToken = zhiyuCfg.AccessToken
	}

	// 设置模型参数
	cfg.MaxTokens = options.MaxTokens
	cfg.Temperature = options.Temperature
	cfg.TopP = options.TopP
	cfg.TopK = options.TopK
	cfg.Stop = options.Stop
	cfg.FrequencyPenalty = options.FrequencyPenalty
	cfg.PresencePenalty = options.PresencePenalty

	return &LLM{
		client: NewClient(cfg),
	}, nil
}

// Generate 非流式生成
func (l *LLM) Generate(ctx context.Context, input []*entity.Message, opts ...entity.Option) (*entity.Message, error) {
	return l.client.Chat(ctx, input, opts...)
}

// Stream 流式生成
func (l *LLM) Stream(ctx context.Context, input []*entity.Message, opts ...entity.Option) (entity.IStreamReader, error) {
	return l.client.ChatStream(ctx, input, opts...)
}
