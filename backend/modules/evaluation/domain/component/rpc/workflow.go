// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"
)

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - RPC接口定义

// IWorkflowRPCAdapter 智宇工作流RPC适配器接口
//
//go:generate mockgen -destination=mocks/workflow.go -package=mocks . IWorkflowRPCAdapter
type IWorkflowRPCAdapter interface {
	// ListWorkflows 获取工作流列表
	ListWorkflows(ctx context.Context, param *ListWorkflowsParam) (workflows []*ZhiyuWorkflow, total int64, err error)
	// GetWorkflowDetail 获取工作流详情（包含参数信息）
	GetWorkflowDetail(ctx context.Context, sceneKey string) (detail *WorkflowDetail, err error)
	// ExecuteWorkflow 执行工作流
	ExecuteWorkflow(ctx context.Context, param *ExecuteWorkflowParam) (result *ExecuteWorkflowResult, err error)
	// GetDefaultParams 获取默认参数配置（processCode、appId、accessToken）
	GetDefaultParams(ctx context.Context) *WorkflowDefaultParams
}

// WorkflowDefaultParams 工作流默认参数
type WorkflowDefaultParams struct {
	ProcessCode string // 流程编码
	AppID       string // 应用ID
	AccessToken string // 访问令牌
}

// ListWorkflowsParam 获取工作流列表参数
type ListWorkflowsParam struct {
	PageNum  int32
	PageSize int32
	KeyWord  *string
}

// ZhiyuWorkflow 智宇工作流
type ZhiyuWorkflow struct {
	SceneID      int64  // 工作流ID
	SceneName    string // 工作流名称
	SceneKey     string // 工作流唯一标识
	ScenePath    string // 图标(base64)
	Remark       string // 描述
	ReleaseState string // 发布状态
}

// WorkflowDetail 工作流详情
type WorkflowDetail struct {
	SceneKey     string             // 工作流唯一标识
	GlobalParams []*WorkflowParam   // 全局参数列表
	APIHttpHost  string             // HTTP API Host
	APIHttpPath  string             // HTTP API Path
}

// WorkflowParam 工作流参数
type WorkflowParam struct {
	ParamName   string // 参数名
	ParamNameCn string // 参数中文名
	ValueType   string // 参数类型
	IsRequired  int32  // 是否必填 1=必填 0=非必填
}

// ExecuteWorkflowParam 执行工作流参数
type ExecuteWorkflowParam struct {
	SceneKey   string            // 工作流标识
	InputData  map[string]string // 动态输入参数
}

// ExecuteWorkflowResult 执行工作流结果
type ExecuteWorkflowResult struct {
	Content string // 返回内容
	Code    int32  // 状态码
	Message string // 消息
}

// --------------end-----------------------
