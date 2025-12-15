// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"strconv"
	"time"

	"github.com/bytedance/gg/gptr"

	"github.com/coze-dev/coze-loop/backend/infra/middleware/session"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/consts"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/domain/component/rpc"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/domain/entity"
	"github.com/coze-dev/coze-loop/backend/modules/evaluation/pkg/errno"
	"github.com/coze-dev/coze-loop/backend/pkg/ctxcache"
	"github.com/coze-dev/coze-loop/backend/pkg/errorx"
	"github.com/coze-dev/coze-loop/backend/pkg/logs"
)

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - ISourceEvalTargetOperateService实现

// WorkflowSourceEvalTargetServiceImpl 智宇工作流评测对象服务实现
type WorkflowSourceEvalTargetServiceImpl struct {
	workflowRPCAdapter rpc.IWorkflowRPCAdapter
}

// NewWorkflowSourceEvalTargetServiceImpl 创建智宇工作流评测对象服务
func NewWorkflowSourceEvalTargetServiceImpl(workflowRPCAdapter rpc.IWorkflowRPCAdapter) ISourceEvalTargetOperateService {
	return &WorkflowSourceEvalTargetServiceImpl{
		workflowRPCAdapter: workflowRPCAdapter,
	}
}

// RuntimeParam 返回运行时参数
func (t *WorkflowSourceEvalTargetServiceImpl) RuntimeParam() entity.IRuntimeParam {
	return nil // 工作流暂不支持运行时参数
}

// EvalType 返回评测对象类型
func (t *WorkflowSourceEvalTargetServiceImpl) EvalType() entity.EvalTargetType {
	return entity.EvalTargetTypeCozeWorkflow
}

// ValidateInput 验证输入
func (t *WorkflowSourceEvalTargetServiceImpl) ValidateInput(ctx context.Context, spaceID int64, inputSchema []*entity.ArgsSchema, input *entity.EvalTargetInputData) error {
	return input.ValidateInputSchema(inputSchema)
}

// Execute 执行工作流
func (t *WorkflowSourceEvalTargetServiceImpl) Execute(ctx context.Context, spaceID int64, param *entity.ExecuteEvalTargetParam) (evaluatorOutputData *entity.EvalTargetOutputData, status entity.EvalTargetRunStatus, err error) {
	start := time.Now()

	evaluatorOutputData = &entity.EvalTargetOutputData{}
	defer func() {
		timeCostMS := time.Since(start).Milliseconds()
		evaluatorOutputData.TimeConsumingMS = gptr.Of(timeCostMS)
		if err != nil {
			evaluatorOutputData.EvalTargetRunError = &entity.EvalTargetRunError{}
			statusErr, ok := errorx.FromStatusError(err)
			if ok {
				evaluatorOutputData.EvalTargetRunError.Code = statusErr.Code()
				evaluatorOutputData.EvalTargetRunError.Message = statusErr.Error()
			} else {
				evaluatorOutputData.EvalTargetRunError.Code = errno.CommonInternalErrorCode
				evaluatorOutputData.EvalTargetRunError.Message = err.Error()
			}
		}
	}()

	// 构建执行参数
	execParam := &rpc.ExecuteWorkflowParam{
		SceneKey:  param.SourceTargetID, // sceneKey 作为 SourceTargetID
		InputData: make(map[string]string),
	}

	if param != nil && param.Input != nil && param.Input.Ext != nil {
		ctx = ctxcache.Init(ctx)
		if v := param.Input.Ext[consts.ZhiyuAuthorizationExtKey]; len(v) > 0 {
			ctxcache.Store(ctx, consts.ZhiyuAuthorizationCtxKey, v)
		}
		if v := param.Input.Ext[consts.ZhiyuAuthTokenExtKey]; len(v) > 0 {
			ctxcache.Store(ctx, consts.ZhiyuAuthTokenCtxKey, v)
		}
	}

	// 转换输入字段
	for key, content := range param.Input.InputFields {
		if content != nil && content.Text != nil {
			execParam.InputData[key] = *content.Text
		}
	}

	// 执行工作流
	result, err := t.workflowRPCAdapter.ExecuteWorkflow(ctx, execParam)
	if err != nil {
		return evaluatorOutputData, entity.EvalTargetRunStatusFail, err
	}

	// 构建输出
	outputStr := result.Content
	evaluatorOutputData.OutputFields = map[string]*entity.Content{
		consts.OutputSchemaKey: {
			ContentType: gptr.Of(entity.ContentTypeText),
			Format:      gptr.Of(entity.Markdown),
			Text:        &outputStr,
		},
	}

	return evaluatorOutputData, entity.EvalTargetRunStatusSuccess, nil
}

// BuildBySource 根据源构建评测对象
func (t *WorkflowSourceEvalTargetServiceImpl) BuildBySource(ctx context.Context, spaceID int64, sourceTargetID, sourceTargetVersion string, opts ...entity.Option) (*entity.EvalTarget, error) {
	// 获取工作流详情
	detail, err := t.workflowRPCAdapter.GetWorkflowDetail(ctx, sourceTargetID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, errorx.NewByCode(errno.ResourceNotFoundCode)
	}

	// 构建输入Schema
	inputSchema := make([]*entity.ArgsSchema, 0)
	for _, param := range detail.GlobalParams {
		jsonSchema := convertValueTypeToJsonSchema(param.ValueType)
		inputSchema = append(inputSchema, &entity.ArgsSchema{
			Key:                 gptr.Of(param.ParamName),
			SupportContentTypes: []entity.ContentType{entity.ContentTypeText},
			JsonSchema:          gptr.Of(jsonSchema),
		})
	}

	userIDInContext := session.UserIDInCtxOrEmpty(ctx)
	do := &entity.EvalTarget{
		SpaceID:        spaceID,
		SourceTargetID: sourceTargetID,
		EvalTargetType: entity.EvalTargetTypeCozeWorkflow,
		EvalTargetVersion: &entity.EvalTargetVersion{
			SpaceID:             spaceID,
			SourceTargetVersion: sourceTargetVersion, // 固定版本 "0.0.1"
			EvalTargetType:      entity.EvalTargetTypeCozeWorkflow,
			CozeWorkflow: &entity.CozeWorkflow{
				ID:      sourceTargetID,
				Version: sourceTargetVersion,
			},
			InputSchema: inputSchema,
			OutputSchema: []*entity.ArgsSchema{
				{
					Key:                 gptr.Of(consts.OutputSchemaKey),
					SupportContentTypes: []entity.ContentType{entity.ContentTypeText},
					JsonSchema:          gptr.Of(consts.StringJsonSchema),
				},
			},
			BaseInfo: &entity.BaseInfo{
				CreatedBy: &entity.UserInfo{
					UserID: gptr.Of(userIDInContext),
				},
				UpdatedBy: &entity.UserInfo{
					UserID: gptr.Of(userIDInContext),
				},
			},
		},
		BaseInfo: &entity.BaseInfo{
			CreatedBy: &entity.UserInfo{
				UserID: gptr.Of(userIDInContext),
			},
			UpdatedBy: &entity.UserInfo{
				UserID: gptr.Of(userIDInContext),
			},
		},
	}
	return do, nil
}

// ListSource 列出源
func (t *WorkflowSourceEvalTargetServiceImpl) ListSource(ctx context.Context, param *entity.ListSourceParam) (targets []*entity.EvalTarget, nextCursor string, hasMore bool, err error) {
	page, err := buildPageByCursor(param.Cursor)
	if err != nil {
		return nil, "", false, err
	}

	// 请求工作流列表
	workflows, _, err := t.workflowRPCAdapter.ListWorkflows(ctx, &rpc.ListWorkflowsParam{
		PageNum:  page,
		PageSize: gptr.Indirect(param.PageSize),
		KeyWord:  param.KeyWord,
	})
	if err != nil {
		return nil, "", false, err
	}

	// 构建结果
	targets = make([]*entity.EvalTarget, 0)
	for _, wf := range workflows {
		targets = append(targets, &entity.EvalTarget{
			SpaceID:        gptr.Indirect(param.SpaceID),
			SourceTargetID: wf.SceneKey,
			EvalTargetType: entity.EvalTargetTypeCozeWorkflow,
			EvalTargetVersion: &entity.EvalTargetVersion{
				SpaceID:        gptr.Indirect(param.SpaceID),
				EvalTargetType: entity.EvalTargetTypeCozeWorkflow,
				CozeWorkflow: &entity.CozeWorkflow{
					ID:          wf.SceneKey,
					Version:     "0.0.1", // 固定版本
					Name:        wf.SceneName,
					AvatarURL:   wf.ScenePath,
					Description: wf.Remark,
				},
			},
		})
	}

	return targets, strconv.FormatInt(int64(page+1), 10), len(workflows) == int(gptr.Indirect(param.PageSize)), nil
}

// ListSourceVersion 列出源版本
func (t *WorkflowSourceEvalTargetServiceImpl) ListSourceVersion(ctx context.Context, param *entity.ListSourceVersionParam) (versions []*entity.EvalTargetVersion, nextCursor string, hasMore bool, err error) {
	// 工作流没有版本概念，返回固定版本 0.0.1
	// 先获取工作流详情来填充名称等信息
	workflows, _, err := t.workflowRPCAdapter.ListWorkflows(ctx, &rpc.ListWorkflowsParam{
		PageNum:  1,
		PageSize: 9999,
	})
	if err != nil {
		return nil, "", false, err
	}

	// 查找对应的工作流
	var targetWorkflow *rpc.ZhiyuWorkflow
	for _, wf := range workflows {
		if wf.SceneKey == param.SourceTargetID {
			targetWorkflow = wf
			break
		}
	}

	if targetWorkflow == nil {
		return nil, "", false, errorx.NewByCode(errno.ResourceNotFoundCode)
	}

	// 获取工作流详情（包含参数信息）
	detail, err := t.workflowRPCAdapter.GetWorkflowDetail(ctx, param.SourceTargetID)
	if err != nil {
		return nil, "", false, err
	}

	// 获取默认参数配置
	defaultParams := t.workflowRPCAdapter.GetDefaultParams(ctx)

	// 构建 InputSchema
	inputSchemas := make([]*entity.ArgsSchema, 0)

	// 添加4个默认字段（都是必填，都有默认值，显示为Input输入框）
	inputSchemas = append(inputSchemas,
		&entity.ArgsSchema{
			Key:                 gptr.Of("sceneKey"),
			SupportContentTypes: []entity.ContentType{entity.ContentTypeText},
			JsonSchema:          gptr.Of(`{"type":"string"}`),
			IsRequired:          gptr.Of(true),
			DefaultValue:        gptr.Of(param.SourceTargetID), // sceneKey 就是工作流ID
		},
		&entity.ArgsSchema{
			Key:                 gptr.Of("processCode"),
			SupportContentTypes: []entity.ContentType{entity.ContentTypeText},
			JsonSchema:          gptr.Of(`{"type":"string"}`),
			IsRequired:          gptr.Of(true),
			DefaultValue:        gptr.Of(defaultParams.ProcessCode),
		},
		&entity.ArgsSchema{
			Key:                 gptr.Of("appId"),
			SupportContentTypes: []entity.ContentType{entity.ContentTypeText},
			JsonSchema:          gptr.Of(`{"type":"string"}`),
			IsRequired:          gptr.Of(true),
			DefaultValue:        gptr.Of(defaultParams.AppID),
		},
		&entity.ArgsSchema{
			Key:                 gptr.Of("accessToken"),
			SupportContentTypes: []entity.ContentType{entity.ContentTypeText},
			JsonSchema:          gptr.Of(`{"type":"string"}`),
			IsRequired:          gptr.Of(true),
			DefaultValue:        gptr.Of(defaultParams.AccessToken),
		},
	)

	// 添加工作流详情中的 globalParams
	if detail != nil && detail.GlobalParams != nil {
		for _, p := range detail.GlobalParams {
			// 跳过与默认字段重复的参数
			if p.ParamName == "sceneKey" || p.ParamName == "processCode" || p.ParamName == "appId" || p.ParamName == "accessToken" {
				continue
			}
			// isRequired: 1=必填, 0=非必填
			isRequired := p.IsRequired == 1
			inputSchemas = append(inputSchemas, &entity.ArgsSchema{
				Key:                 gptr.Of(p.ParamName),
				SupportContentTypes: []entity.ContentType{entity.ContentTypeText},
				JsonSchema:          gptr.Of(valueTypeToJsonSchema(p.ValueType)),
				IsRequired:          gptr.Of(isRequired),
				// globalParams 中的字段没有默认值，显示为下拉选择框，让用户选择评测集字段
			})
		}
	}

	versions = []*entity.EvalTargetVersion{
		{
			SpaceID:             gptr.Indirect(param.SpaceID),
			SourceTargetVersion: "0.0.1", // 固定版本
			EvalTargetType:      entity.EvalTargetTypeCozeWorkflow,
			CozeWorkflow: &entity.CozeWorkflow{
				ID:          targetWorkflow.SceneKey,
				Version:     "0.0.1",
				Name:        targetWorkflow.SceneName,
				AvatarURL:   targetWorkflow.ScenePath,
				Description: targetWorkflow.Remark,
			},
			InputSchema: inputSchemas,
		},
	}

	return versions, "", false, nil
}

// valueTypeToJsonSchema 将智宇工作流的 valueType 转换为 JSON Schema
func valueTypeToJsonSchema(valueType string) string {
	switch valueType {
	case "string":
		return `{"type":"string"}`
	case "integer", "int":
		return `{"type":"integer"}`
	case "number", "float", "double":
		return `{"type":"number"}`
	case "boolean", "bool":
		return `{"type":"boolean"}`
	case "array<string>":
		return `{"type":"array","items":{"type":"string"}}`
	case "array<integer>", "array<int>":
		return `{"type":"array","items":{"type":"integer"}}`
	case "array<number>":
		return `{"type":"array","items":{"type":"number"}}`
	case "object":
		return `{"type":"object"}`
	default:
		return `{"type":"string"}`
	}
}

// PackSourceInfo 填充源信息
func (t *WorkflowSourceEvalTargetServiceImpl) PackSourceInfo(ctx context.Context, spaceID int64, dos []*entity.EvalTarget) (err error) {
	// 获取所有工作流列表
	// 注意：这里可能没有 authorization（如实验列表页面），此时应跳过而不是报错
	workflows, _, err := t.workflowRPCAdapter.ListWorkflows(ctx, &rpc.ListWorkflowsParam{
		PageNum:  1,
		PageSize: 9999,
	})
	if err != nil {
		// 如果是因为没有 authorization 导致的错误，跳过不报错
		logs.CtxWarn(ctx, "PackSourceInfo: ListWorkflows failed, skip filling workflow info, err=%v", err)
		return nil
	}

	// 构建映射
	workflowMap := make(map[string]*rpc.ZhiyuWorkflow)
	for _, wf := range workflows {
		workflowMap[wf.SceneKey] = wf
	}

	// 填充信息
	for _, do := range dos {
		if do.EvalTargetType != entity.EvalTargetTypeCozeWorkflow {
			continue
		}
		if wf, ok := workflowMap[do.SourceTargetID]; ok {
			do.EvalTargetVersion = &entity.EvalTargetVersion{
				CozeWorkflow: &entity.CozeWorkflow{
					Name: wf.SceneName,
				},
			}
		}
	}
	return nil
}

// PackSourceVersionInfo 填充源版本信息
func (t *WorkflowSourceEvalTargetServiceImpl) PackSourceVersionInfo(ctx context.Context, spaceID int64, dos []*entity.EvalTarget) (err error) {
	// 获取所有工作流列表
	// 注意：这里可能没有 authorization（如实验列表页面），此时应跳过而不是报错
	workflows, _, err := t.workflowRPCAdapter.ListWorkflows(ctx, &rpc.ListWorkflowsParam{
		PageNum:  1,
		PageSize: 9999,
	})
	if err != nil {
		// 如果是因为没有 authorization 导致的错误，跳过不报错
		logs.CtxWarn(ctx, "PackSourceVersionInfo: ListWorkflows failed, skip filling workflow info, err=%v", err)
		return nil
	}

	// 构建映射
	workflowMap := make(map[string]*rpc.ZhiyuWorkflow)
	for _, wf := range workflows {
		workflowMap[wf.SceneKey] = wf
	}

	// 填充信息
	for _, do := range dos {
		if do.EvalTargetType != entity.EvalTargetTypeCozeWorkflow {
			continue
		}
		if do.EvalTargetVersion == nil || do.EvalTargetVersion.CozeWorkflow == nil {
			continue
		}
		if wf, ok := workflowMap[do.SourceTargetID]; ok {
			do.EvalTargetVersion.CozeWorkflow.Name = wf.SceneName
			do.EvalTargetVersion.CozeWorkflow.Description = wf.Remark
		} else {
			do.BaseInfo.DeletedAt = gptr.Of(int64(1)) // 说明源数据已删除
		}
	}
	return nil
}

// BatchGetSource 批量获取源
func (t *WorkflowSourceEvalTargetServiceImpl) BatchGetSource(ctx context.Context, spaceID int64, ids []string) (targets []*entity.EvalTarget, err error) {
	// 获取所有工作流列表
	workflows, _, err := t.workflowRPCAdapter.ListWorkflows(ctx, &rpc.ListWorkflowsParam{
		PageNum:  1,
		PageSize: 9999,
	})
	if err != nil {
		return nil, err
	}

	// 构建映射
	workflowMap := make(map[string]*rpc.ZhiyuWorkflow)
	for _, wf := range workflows {
		workflowMap[wf.SceneKey] = wf
	}

	// 构建结果
	targets = make([]*entity.EvalTarget, 0)
	for _, id := range ids {
		if wf, ok := workflowMap[id]; ok {
			targets = append(targets, &entity.EvalTarget{
				SpaceID:        spaceID,
				SourceTargetID: wf.SceneKey,
				EvalTargetType: entity.EvalTargetTypeCozeWorkflow,
				EvalTargetVersion: &entity.EvalTargetVersion{
					SpaceID:        spaceID,
					EvalTargetType: entity.EvalTargetTypeCozeWorkflow,
					CozeWorkflow: &entity.CozeWorkflow{
						ID:          wf.SceneKey,
						Name:        wf.SceneName,
						AvatarURL:   wf.ScenePath,
						Description: wf.Remark,
					},
				},
			})
		}
	}
	return targets, nil
}

// convertValueTypeToJsonSchema 将工作流参数类型转换为JSON Schema
func convertValueTypeToJsonSchema(valueType string) string {
	switch valueType {
	case "string":
		return consts.StringJsonSchema
	case "integer":
		return consts.IntegerJsonSchema
	case "number":
		return consts.NumberJsonSchema
	case "boolean":
		return consts.BooleanJsonSchema
	case "object":
		return consts.ObjectJsonSchema
	case "array<string>":
		return consts.ArrayStringJsonSchema
	case "array<integer>":
		return consts.ArrayIntegerJsonSchema
	case "array<number>":
		return consts.ArrayNumberJsonSchema
	case "array<boolean>":
		return consts.ArrayBooleanJsonSchema
	case "array<object>":
		return consts.ArrayObjectJsonSchema
	default:
		return consts.StringJsonSchema // 默认是string
	}
}

// --------------end-----------------------
