// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package workflow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bytedance/sonic"

	"github.com/coze-dev/coze-loop/backend/modules/evaluation/domain/component/rpc"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/pkg/conf"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/pkg/errno"
	"github.com/coze-dev/coze-loop/backend/pkg/errorx"
	"github.com/coze-dev/coze-loop/backend/pkg/logs"
)

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - HTTP调用实现

// WorkflowRPCAdapter 智宇工作流RPC适配器实现
type WorkflowRPCAdapter struct {
	configer   conf.IWorkflowConfiger
	httpClient *http.Client
}

// NewWorkflowRPCAdapter 创建智宇工作流RPC适配器
func NewWorkflowRPCAdapter(configer conf.IWorkflowConfiger) rpc.IWorkflowRPCAdapter {
	return &WorkflowRPCAdapter{
		configer: configer,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ListWorkflows 获取工作流列表
func (w *WorkflowRPCAdapter) ListWorkflows(ctx context.Context, param *rpc.ListWorkflowsParam) (workflows []*rpc.ZhiyuWorkflow, total int64, err error) {
	cfg := w.configer.GetWorkflowConfig(ctx)
	if cfg == nil || cfg.BaseURL == "" {
		return nil, 0, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("workflow config not found"))
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/aistar_server/scene/api/v1/list?releaseState=&partition=1&teamId=&pageSize=%d&pageNum=%d",
		cfg.BaseURL, param.PageSize, param.PageNum)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, errorx.Wrapf(err, "create list workflows request failed")
	}

	// 设置Header
	req.Header.Set("Authorization", cfg.Authorization)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, 0, errorx.Wrapf(err, "list workflows request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, errorx.Wrapf(err, "read list workflows response failed")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, 0, errorx.NewByCode(errno.CommonRPCErrorCode, errorx.WithExtraMsg(fmt.Sprintf("list workflows failed, status: %d, body: %s", resp.StatusCode, string(body))))
	}

	// 解析响应
	var listResp ListWorkflowsResponse
	if err := sonic.Unmarshal(body, &listResp); err != nil {
		return nil, 0, errorx.Wrapf(err, "unmarshal list workflows response failed")
	}

	if listResp.Code != 1000 {
		return nil, 0, errorx.NewByCode(errno.CommonRPCErrorCode, errorx.WithExtraMsg(fmt.Sprintf("list workflows failed, code: %d, message: %s", listResp.Code, listResp.Message)))
	}

	// 转换结果
	workflows = make([]*rpc.ZhiyuWorkflow, 0, len(listResp.Data.List))
	for _, item := range listResp.Data.List {
		// 如果有关键词过滤
		if param.KeyWord != nil && *param.KeyWord != "" {
			// 简单的名称匹配
			if !containsKeyword(item.SceneName, *param.KeyWord) {
				continue
			}
		}
		workflows = append(workflows, &rpc.ZhiyuWorkflow{
			SceneID:      item.SceneID,
			SceneName:    item.SceneName,
			SceneKey:     item.SceneKey,
			ScenePath:    item.ScenePath,
			Remark:       item.Remark,
			ReleaseState: item.ReleaseState,
		})
	}

	return workflows, listResp.Data.Total, nil
}

// GetWorkflowDetail 获取工作流详情
func (w *WorkflowRPCAdapter) GetWorkflowDetail(ctx context.Context, sceneKey string) (detail *rpc.WorkflowDetail, err error) {
	cfg := w.configer.GetWorkflowConfig(ctx)
	if cfg == nil || cfg.BaseURL == "" {
		return nil, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("workflow config not found"))
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/aistar_server/scene/release/getApiReleaseByKey", cfg.BaseURL)

	// 构建请求体
	reqBody := map[string]string{
		"sceneKey": sceneKey,
	}
	reqBodyBytes, err := sonic.Marshal(reqBody)
	if err != nil {
		return nil, errorx.Wrapf(err, "marshal get workflow detail request failed")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, errorx.Wrapf(err, "create get workflow detail request failed")
	}

	// 设置Header
	req.Header.Set("Authorization", cfg.Authorization)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, errorx.Wrapf(err, "get workflow detail request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errorx.Wrapf(err, "read get workflow detail response failed")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errorx.NewByCode(errno.CommonRPCErrorCode, errorx.WithExtraMsg(fmt.Sprintf("get workflow detail failed, status: %d, body: %s", resp.StatusCode, string(body))))
	}

	// 解析响应
	var detailResp GetWorkflowDetailResponse
	if err := sonic.Unmarshal(body, &detailResp); err != nil {
		return nil, errorx.Wrapf(err, "unmarshal get workflow detail response failed")
	}

	if detailResp.Code != 1000 {
		return nil, errorx.NewByCode(errno.CommonRPCErrorCode, errorx.WithExtraMsg(fmt.Sprintf("get workflow detail failed, code: %d, message: %s", detailResp.Code, detailResp.Message)))
	}

	// 转换结果
	detail = &rpc.WorkflowDetail{
		SceneKey:    sceneKey,
		APIHttpHost: detailResp.Data.Details.APIHttpHost,
		APIHttpPath: detailResp.Data.Details.APIHttpPath,
	}

	// 转换参数列表
	detail.GlobalParams = make([]*rpc.WorkflowParam, 0, len(detailResp.Data.SceneBack.GlobalParams))
	for _, param := range detailResp.Data.SceneBack.GlobalParams {
		detail.GlobalParams = append(detail.GlobalParams, &rpc.WorkflowParam{
			ParamName:   param.ParamName,
			ParamNameCn: param.ParamNameCn,
			ValueType:   param.ValueType,
			IsRequired:  param.IsRequired,
		})
	}

	return detail, nil
}

// ExecuteWorkflow 执行工作流
func (w *WorkflowRPCAdapter) ExecuteWorkflow(ctx context.Context, param *rpc.ExecuteWorkflowParam) (result *rpc.ExecuteWorkflowResult, err error) {
	cfg := w.configer.GetWorkflowConfig(ctx)
	if cfg == nil || cfg.ExecuteURL == "" {
		return nil, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("workflow execute config not found"))
	}

	// 构建请求体 - 动态参数 + 固定参数
	reqBody := make(map[string]interface{})
	for k, v := range param.InputData {
		reqBody[k] = v
	}
	// 添加固定参数
	reqBody["sceneKey"] = param.SceneKey
	reqBody["processCode"] = cfg.ProcessCode
	reqBody["appId"] = cfg.AppID
	reqBody["accessToken"] = cfg.AccessToken

	reqBodyBytes, err := sonic.Marshal(reqBody)
	if err != nil {
		return nil, errorx.Wrapf(err, "marshal execute workflow request failed")
	}

	logs.CtxInfo(ctx, "ExecuteWorkflow request: url=%s, body=%s", cfg.ExecuteURL, string(reqBodyBytes))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.ExecuteURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, errorx.Wrapf(err, "create execute workflow request failed")
	}

	// 设置Header
	req.Header.Set("AuthToken", cfg.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, errorx.Wrapf(err, "execute workflow request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errorx.Wrapf(err, "read execute workflow response failed")
	}

	logs.CtxInfo(ctx, "ExecuteWorkflow response: status=%d, body=%s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, errorx.NewByCode(errno.CommonRPCErrorCode, errorx.WithExtraMsg(fmt.Sprintf("execute workflow failed, status: %d, body: %s", resp.StatusCode, string(body))))
	}

	// 解析响应
	var execResp ExecuteWorkflowResponse
	if err := sonic.Unmarshal(body, &execResp); err != nil {
		// 如果解析失败，直接返回原始内容
		return &rpc.ExecuteWorkflowResult{
			Content: string(body),
			Code:    0,
			Message: "success",
		}, nil
	}

	return &rpc.ExecuteWorkflowResult{
		Content: execResp.Data,
		Code:    execResp.Code,
		Message: execResp.Message,
	}, nil
}

// GetDefaultParams 获取默认参数配置
func (w *WorkflowRPCAdapter) GetDefaultParams(ctx context.Context) *rpc.WorkflowDefaultParams {
	cfg := w.configer.GetWorkflowConfig(ctx)
	if cfg == nil {
		return &rpc.WorkflowDefaultParams{}
	}
	return &rpc.WorkflowDefaultParams{
		ProcessCode: cfg.ProcessCode,
		AppID:       cfg.AppID,
		AccessToken: cfg.AccessToken,
	}
}

// containsKeyword 检查字符串是否包含关键词
func containsKeyword(s, keyword string) bool {
	return len(s) > 0 && len(keyword) > 0 && (s == keyword || len(s) >= len(keyword) && (s[:len(keyword)] == keyword || s[len(s)-len(keyword):] == keyword || findSubstring(s, keyword)))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// --------------end-----------------------
