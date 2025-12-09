// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"context"

	"github.com/coze-dev/coze-loop/backend/pkg/conf"
)

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 配置接口和实现

// IWorkflowConfiger 智宇工作流配置接口
//
//go:generate mockgen -destination=mocks/workflow_configer.go -package=mocks . IWorkflowConfiger
type IWorkflowConfiger interface {
	GetWorkflowConfig(ctx context.Context) *WorkflowConfig
}

// WorkflowConfig 智宇工作流配置
type WorkflowConfig struct {
	// 列表/详情 API 配置
	BaseURL       string `json:"base_url" yaml:"base_url" mapstructure:"base_url"`
	Authorization string `json:"authorization" yaml:"authorization" mapstructure:"authorization"`

	// 执行 API 配置
	ExecuteURL string `json:"execute_url" yaml:"execute_url" mapstructure:"execute_url"`
	AuthToken  string `json:"auth_token" yaml:"auth_token" mapstructure:"auth_token"`

	// 执行时的固定参数
	ProcessCode string `json:"process_code" yaml:"process_code" mapstructure:"process_code"`
	AppID       string `json:"app_id" yaml:"app_id" mapstructure:"app_id"`
	AccessToken string `json:"access_token" yaml:"access_token" mapstructure:"access_token"`
}

// workflowConfiger 智宇工作流配置实现
type workflowConfiger struct {
	loader conf.IConfigLoader
}

// NewWorkflowConfiger 创建智宇工作流配置
func NewWorkflowConfiger(configFactory conf.IConfigLoaderFactory) IWorkflowConfiger {
	loader, err := configFactory.NewConfigLoader("evaluation.yaml")
	if err != nil {
		return &workflowConfiger{}
	}
	return &workflowConfiger{
		loader: loader,
	}
}

// GetWorkflowConfig 获取工作流配置
func (c *workflowConfiger) GetWorkflowConfig(ctx context.Context) *WorkflowConfig {
	if c.loader == nil {
		return DefaultWorkflowConfig()
	}
	const key = "zhiyu_workflow"
	cfg := &WorkflowConfig{}
	if err := c.loader.UnmarshalKey(ctx, key, cfg); err != nil {
		return DefaultWorkflowConfig()
	}
	return cfg
}

// DefaultWorkflowConfig 默认工作流配置
func DefaultWorkflowConfig() *WorkflowConfig {
	return &WorkflowConfig{
		ProcessCode: "aigc_workflow_http_ability",
	}
}

// --------------end-----------------------
