// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package workflow

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - API响应结构体定义

// ListWorkflowsResponse 工作流列表响应
type ListWorkflowsResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List     []WorkflowItem `json:"list"`
		Total    int64          `json:"total"`
		PageNum  int32          `json:"pageNum"`
		PageSize int32          `json:"pageSize"`
	} `json:"data"`
}

// WorkflowItem 工作流列表项
type WorkflowItem struct {
	SceneID      int64  `json:"sceneId"`
	SceneName    string `json:"sceneName"`
	SceneKey     string `json:"sceneKey"`
	ScenePath    string `json:"scenePath"`
	Remark       string `json:"remark"`
	ReleaseState string `json:"releaseState"`
	SceneType    string `json:"sceneType"`
}

// GetWorkflowDetailResponse 工作流详情响应
type GetWorkflowDetailResponse struct {
	Status  string `json:"status"`
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Details struct {
			APIHttpHost string `json:"apiHttpHost"`
			APIHttpPath string `json:"apiHttpPath"`
			APIWsHost   string `json:"apiWsHost"`
			APIWsPath   string `json:"apiWsPath"`
			APISseHost  string `json:"apiSseHost"`
			APISsePath  string `json:"apiSsePath"`
		} `json:"details"`
		SceneBack struct {
			GlobalParams []GlobalParam `json:"globalParams"`
		} `json:"sceneBack"`
	} `json:"data"`
}

// GlobalParam 全局参数
type GlobalParam struct {
	ID          int64  `json:"id"`
	ParamName   string `json:"paramName"`
	ParamNameCn string `json:"paramNameCn"`
	ValueType   string `json:"valueType"`
	IsRequired  int32  `json:"isRequired"`
	ParamType   string `json:"paramType"`
	ParamDesc   string `json:"paramDesc"`
}

// ExecuteWorkflowResponse 执行工作流响应
type ExecuteWorkflowResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"msg"`
	Status  string `json:"status"`
	Data    string `json:"data"`
}

// --------------end-----------------------
