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
	"github.com/cloudwego/kitex/pkg/kerrors"

	"github.com/coze-dev/coze-loop/backend/modules/evaluation/consts"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/domain/component/rpc"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/pkg/conf"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/pkg/errno"
	"github.com/coze-dev/coze-loop/backend/pkg/ctxcache"
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

	// authorization 必须由前端传入，不再从配置文件读取
	authorization, ok := ctxcache.Get[string](ctx, consts.ZhiyuAuthorizationCtxKey)
	if !ok || len(authorization) == 0 {
		return nil, 0, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("请在页面上填写 authorization"))
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/aistar_server/scene/api/v1/list?releaseState=&partition=1&teamId=&pageSize=%d&pageNum=%d",
		cfg.BaseURL, param.PageSize, param.PageNum)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, errorx.Wrapf(err, "create list workflows request failed")
	}

	// 设置Header
	req.Header.Set("Authorization", authorization)
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
			SceneType:    item.SceneType,
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

	// authorization 必须由前端传入，不再从配置文件读取
	authorization, ok := ctxcache.Get[string](ctx, consts.ZhiyuAuthorizationCtxKey)
	if !ok || len(authorization) == 0 {
		return nil, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("请在页面上填写 authorization"))
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
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json")

	// 记录请求日志
	logs.CtxInfo(ctx, "GetWorkflowDetail request: url=%s, sceneKey=%s, headers={Authorization: [REDACTED], Content-Type: application/json}", url, sceneKey)

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

	// 记录响应日志
	logs.CtxInfo(ctx, "GetWorkflowDetail response: status=%d, body=%s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, errorx.NewByCode(errno.CommonRPCErrorCode, errorx.WithExtraMsg(fmt.Sprintf("get workflow detail failed, status: %d, body: %s", resp.StatusCode, string(body))))
	}

	// 解析响应
	var detailResp GetWorkflowDetailResponse
	if err := sonic.Unmarshal(body, &detailResp); err != nil {
		logs.CtxError(ctx, "GetWorkflowDetail unmarshal failed: err=%v, body=%s", err, string(body))
		return nil, errorx.Wrapf(err, "unmarshal get workflow detail response failed")
	}

	logs.CtxInfo(ctx, "GetWorkflowDetail parsed response: status=%s, code=%d, message=%s", detailResp.Status, detailResp.Code, detailResp.Message)

	if detailResp.Code != 1000 {
		// 返回 BizStatusError，使用 WorkflowAPIErrorCode（message 为空）避免被国际化翻译覆盖
		errMsg := fmt.Sprintf("智宇工作流错误: %s (code: %d)", detailResp.Message, detailResp.Code)
		return nil, kerrors.NewBizStatusError(errno.WorkflowAPIErrorCode, errMsg)
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

	logs.CtxInfo(ctx, "GetWorkflowDetail success: sceneKey=%s, apiHttpHost=%s, apiHttpPath=%s, globalParamsCount=%d",
		detail.SceneKey, detail.APIHttpHost, detail.APIHttpPath, len(detail.GlobalParams))

	return detail, nil
}

// ExecuteWorkflow 执行工作流
func (w *WorkflowRPCAdapter) ExecuteWorkflow(ctx context.Context, param *rpc.ExecuteWorkflowParam) (result *rpc.ExecuteWorkflowResult, err error) {
	cfg := w.configer.GetWorkflowConfig(ctx)
	if cfg == nil {
		return nil, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("workflow config not found"))
	}

	// auth_token 必须由前端传入，不再从配置文件读取
	authToken, ok := ctxcache.Get[string](ctx, consts.ZhiyuAuthTokenCtxKey)
	if !ok || len(authToken) == 0 {
		return nil, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("请在页面上填写 auth_token"))
	}

	// 从 context 获取 sceneType（由前端传入）
	sceneType, _ := ctxcache.Get[string](ctx, consts.ZhiyuSceneTypeCtxKey)

	// 根据 sceneType 选择 URL
	var executeURL string
	if sceneType == "2" { // 流式
		if cfg.StreamExecuteURL == "" {
			return nil, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("stream execute url not configured"))
		}
		executeURL = fmt.Sprintf("%s/%s", cfg.StreamExecuteURL, param.SceneKey)
	} else { // 非流式（sceneType == "0" 或其他）
		if cfg.ExecuteURL == "" {
			return nil, errorx.NewByCode(errno.CommonInvalidParamCode, errorx.WithExtraMsg("execute url not configured"))
		}
		executeURL = fmt.Sprintf("%s/%s", cfg.ExecuteURL, param.SceneKey)
	}

	// 4. 构建请求体
	reqBody := make(map[string]interface{})

	// 先添加字段映射的数据
	for k, v := range param.InputData {
		reqBody[k] = v
	}

	// 然后设置固定参数（固定参数优先级更高，会覆盖字段映射中的同名字段）
	reqBody["sceneKey"] = param.SceneKey
	// 根据 sceneType 选择 processCode（固定参数，必须覆盖字段映射中的值）
	if sceneType == "2" { // 流式
		if cfg.StreamProcessCode != "" {
			reqBody["processCode"] = cfg.StreamProcessCode
		} else {
			reqBody["processCode"] = cfg.ProcessCode // 如果未配置流式 processCode，则使用默认的
		}
	} else { // 非流式
		reqBody["processCode"] = cfg.ProcessCode
	}
	reqBody["appId"] = cfg.AppID
	reqBody["accessToken"] = cfg.AccessToken

	reqBodyBytes, err := sonic.Marshal(reqBody)
	if err != nil {
		return nil, errorx.Wrapf(err, "marshal execute workflow request failed")
	}

	// 5. 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, executeURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, errorx.Wrapf(err, "create execute workflow request failed")
	}

	// 6. 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AuthToken", authToken)

	// 7. 记录请求日志（包含请求头）
	logs.CtxInfo(ctx, "ExecuteWorkflow request: url=%s, headers={Content-Type: application/json, AuthToken: [REDACTED]}, body=%s",
		executeURL, string(reqBodyBytes))

	// 8. 发送请求
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, errorx.Wrapf(err, "execute workflow request failed")
	}
	defer resp.Body.Close()

	// 9. 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errorx.Wrapf(err, "read execute workflow response failed")
	}

	// 10. 记录响应日志
	logs.CtxInfo(ctx, "ExecuteWorkflow response: status=%d, body=%s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, errorx.NewByCode(errno.CommonRPCErrorCode,
			errorx.WithExtraMsg(fmt.Sprintf("execute workflow failed, status: %d, body: %s", resp.StatusCode, string(body))))
	}

	// 11. 根据 sceneType 处理响应
	if sceneType == "2" {
		// 流式响应处理
		return w.handleStreamingResponse(ctx, body)
	}
	// 非流式响应处理
	return w.handleNormalResponse(ctx, body)
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

// handleStreamingResponse 处理流式响应（SSE格式）
func (w *WorkflowRPCAdapter) handleStreamingResponse(ctx context.Context, body []byte) (*rpc.ExecuteWorkflowResult, error) {
	var contentBuilder bytes.Buffer

	// 按行处理SSE格式
	lines := bytes.Split(body, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("data:")) {
			content := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
			contentBuilder.Write(content)
		}
	}

	return &rpc.ExecuteWorkflowResult{
		Content: contentBuilder.String(),
		Code:    200,
		Message: "success",
	}, nil
}

// handleNormalResponse 处理普通响应（JSON格式）
func (w *WorkflowRPCAdapter) handleNormalResponse(ctx context.Context, body []byte) (*rpc.ExecuteWorkflowResult, error) {
	var execResp ExecuteWorkflowResponse
	if err := sonic.Unmarshal(body, &execResp); err != nil {
		return nil, errorx.Wrapf(err, "unmarshal execute workflow response failed")
	}

	if execResp.ErrCode != 0 {
		return nil, errorx.NewByCode(errno.CommonRPCErrorCode,
			errorx.WithExtraMsg(fmt.Sprintf("workflow execute failed, code: %d, message: %s", execResp.ErrCode, execResp.ErrMsg)))
	}

	return &rpc.ExecuteWorkflowResult{
		Content: execResp.Data.Data,
		Code:    execResp.Data.Code,
		Message: execResp.Data.Msg,
	}, nil
}

// --------------end-----------------------
