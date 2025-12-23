// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 工作流表单组件

import { useEffect, useMemo } from 'react';

import { isEmpty } from 'lodash-es';
import { useRequest } from 'ahooks';
import { I18n } from '@cozeloop/i18n-adapter';
import { useSpace } from '@cozeloop/biz-hooks-adapter';
import { EvalTargetType, type FieldSchema } from '@cozeloop/api-schema/evaluation';
import { StoneEvaluationApi } from '@cozeloop/api-schema';
import { IconCozInfoCircle } from '@coze-arch/coze-design/icons';
import { Form, Tag, Tooltip, Typography } from '@coze-arch/coze-design';

import { type PluginEvalTargetFormProps, type OptionSchema } from '@/types/evaluate-target';
import { EvaluateTargetMappingField } from '@/components/selectors/evaluate-target';

const EvaluateTargetMappingFieldLabel = (
  <div className="inline-flex flex-row items-center">
    {I18n.t('field_mapping')}
    <Tooltip
      theme="dark"
      content={I18n.t(
        'evaluation_set_field_to_evaluation_object_field_mapping',
      )}
    >
      <IconCozInfoCircle className="ml-1 w-4 h-4 coz-fg-secondary" />
    </Tooltip>
  </div>
);

const ZHIYU_AUTHORIZATION_EXT_KEY = 'zhiyu_authorization';
const ZHIYU_AUTH_TOKEN_EXT_KEY = 'zhiyu_auth_token';
const ZHIYU_SCENE_TYPE_EXT_KEY = 'zhiyu_scene_type';
const ZHIYU_AUTHORIZATION_HEADER = 'X-Zhiyu-Authorization';
const ZHIYU_AUTH_TOKEN_HEADER = 'X-Zhiyu-Auth-Token';
const ZHIYU_SCENE_TYPE_HEADER = 'X-Zhiyu-Scene-Type';

// sceneType 选项：0-非流式，2-流式
const SCENE_TYPE_OPTIONS = [
  { value: '0', label: '非流式处理(http)' },
  { value: '2', label: '流式处理(sse)' },
];

/**
 * 智宇工作流评测对象表单
 * 包含: 工作流选择, 版本显示, 字段映射
 */
const WorkflowPluginEvalTargetForm = (props: PluginEvalTargetFormProps) => {
  const {
    formValues,
    createExperimentValues,
    onChange,
  } = props;

  const { spaceID } = useSpace();
  const targetType = formValues.evalTargetType;
  const workflowId = formValues.evalTarget || '';
  const sourceTargetVersion = formValues.evalTargetVersion || '0.0.1';

  const zhiyuAuthorization =
    (formValues.ext as Record<string, string> | undefined)?.[
      ZHIYU_AUTHORIZATION_EXT_KEY
    ] || '';
  const zhiyuAuthToken =
    (formValues.ext as Record<string, string> | undefined)?.[
      ZHIYU_AUTH_TOKEN_EXT_KEY
    ] || '';
  const zhiyuSceneType =
    (formValues.ext as Record<string, string> | undefined)?.[
      ZHIYU_SCENE_TYPE_EXT_KEY
    ] || ''; // 默认为空，需要用户选择

  // 评测集字段
  const evaluationSetSchemas =
    createExperimentValues?.evaluationSetVersionDetail?.evaluation_set_schema
      ?.field_schemas;

  // 是否填写了必要的认证信息、sceneKey 和工作流模式
  const hasRequiredAuth = !!zhiyuAuthorization && !!zhiyuAuthToken;
  const canFetchDetail = hasRequiredAuth && !!workflowId && !!zhiyuSceneType;

  // 获取工作流详情（包含参数）
  const workflowDetailService = useRequest(
    async () => {
      if (!workflowId) return null;
      // 每次请求时重新构建 headers，确保使用最新值
      const currentHeaders: Record<string, string> = {};
      const currentAuth = (formValues.ext as Record<string, string> | undefined)?.[ZHIYU_AUTHORIZATION_EXT_KEY] || '';
      const currentToken = (formValues.ext as Record<string, string> | undefined)?.[ZHIYU_AUTH_TOKEN_EXT_KEY] || '';
      const currentSceneType = (formValues.ext as Record<string, string> | undefined)?.[ZHIYU_SCENE_TYPE_EXT_KEY];
      if (currentAuth) {
        currentHeaders[ZHIYU_AUTHORIZATION_HEADER] = currentAuth;
      }
      if (currentToken) {
        currentHeaders[ZHIYU_AUTH_TOKEN_HEADER] = currentToken;
      }
      // 传递 sceneType 到 header，让后端能够根据 sceneType 返回对应的 processCode 默认值
      // 只有在用户选择了 sceneType 时才传递
      if (currentSceneType) {
        currentHeaders[ZHIYU_SCENE_TYPE_HEADER] = currentSceneType;
      }

      const res = await StoneEvaluationApi.ListSourceEvalTargetVersions(
        {
          workspace_id: spaceID,
          source_target_id: workflowId,
          target_type: EvalTargetType.CozeWorkflow,
          page_size: 1,
        },
        {
          headers: currentHeaders,
        },
      );
      return res.versions?.[0];
    },
    {
      refreshDeps: [workflowId, zhiyuAuthorization, zhiyuAuthToken, zhiyuSceneType],
      ready: canFetchDetail,
      onError: () => {
        // 报错时清空字段映射，避免显示旧数据
        onChange('evalTargetMapping', undefined);
      },
    },
  );

  // 从工作流详情中获取输入参数schema
  // 所有字段都展示，有 default_value 的字段显示为只读输入框
  const inputSchemas = useMemo(() => {
    // 如果有错误，返回空数组，避免使用旧数据
    if (workflowDetailService.error) return [];

    const schemas = workflowDetailService.data?.eval_target_content?.input_schemas;
    if (!schemas) return [];
    return schemas.map(schema => ({
      key: schema.key || '',
      name: schema.key || '',
      text_schema: schema.json_schema,
      isRequired: schema.is_required,
      default_value: schema.default_value, // 有默认值的字段显示为只读输入框
    })) as FieldSchema[];
  }, [workflowDetailService.data, workflowDetailService.error]);

  const handleEvalTargetChange = (value: string) => {
    onChange('evalTarget', value);
    onChange('evalTargetVersion', '0.0.1'); // 固定版本
    onChange('evalTargetMapping', undefined);
  };

  // 当输入参数变化时，初始化字段映射
  useEffect(() => {
    // 如果有错误，清空字段映射
    if (workflowDetailService.error) {
      onChange('evalTargetMapping', undefined);
      return;
    }

    // 只有在成功获取到数据且没有错误时才初始化字段映射
    if (inputSchemas?.length > 0) {
      const payload: Record<string, OptionSchema | undefined> = {};
      const currentMapping = formValues?.evalTargetMapping || {};
      inputSchemas.forEach(v => {
        const fieldName = v?.name || '';
        // 有默认值的字段，使用特殊标记来表示使用默认值
        if ((v as any).default_value) {
          // 使用 unknown 中间转换绕过类型检查
          payload[fieldName] = {
            key: fieldName,
            name: fieldName,
            // 使用 __default_value__ 作为特殊标记，实际值存储在 constValue 中
            schemaSourceType: '__default_value__',
            constValue: (v as any).default_value,
          } as unknown as OptionSchema;
        } else {
          payload[fieldName] = currentMapping?.[fieldName] || undefined;
        }
      });
      onChange('evalTargetMapping', payload);
    }
  }, [inputSchemas, workflowDetailService.error]);

  // 设置默认版本
  useEffect(() => {
    if (workflowId && !formValues.evalTargetVersion) {
      onChange('evalTargetVersion', '0.0.1');
    }
  }, [workflowId]);


  return (
    <>
      {targetType === EvalTargetType.CozeWorkflow ? (
        <>
          <Form.Input
            field={`ext.${ZHIYU_AUTHORIZATION_EXT_KEY}`}
            label="authorization"
            className="w-full"
            autoComplete="off"
            initValue={zhiyuAuthorization}
            rules={[
              { required: true, message: '请输入 authorization' },
            ]}
            onChange={value => {
              onChange('ext', {
                ...(formValues.ext || {}),
                [ZHIYU_AUTHORIZATION_EXT_KEY]: value as string,
              });
              // 清空字段映射，因为认证信息变了
              onChange('evalTargetMapping', undefined);
            }}
          />

          <Form.Input
            field={`ext.${ZHIYU_AUTH_TOKEN_EXT_KEY}`}
            label="auth_token"
            className="w-full"
            autoComplete="off"
            initValue={zhiyuAuthToken}
            rules={[
              { required: true, message: '请输入 auth_token' },
            ]}
            onChange={value => {
              onChange('ext', {
                ...(formValues.ext || {}),
                [ZHIYU_AUTH_TOKEN_EXT_KEY]: value as string,
              });
              // 清空字段映射，因为认证信息变了
              onChange('evalTargetMapping', undefined);
            }}
          />

          {/* sceneKey 输入框 */}
          <Form.Input
            field="evalTarget"
            label="sceneKey"
            className="w-full"
            autoComplete="off"
            placeholder="请输入工作流 sceneKey"
            initValue={workflowId}
            rules={[
              { required: true, message: '请输入 sceneKey' },
            ]}
            onChange={handleEvalTargetChange}
          />

          {/* 流式/非流式选择 */}
          <Form.Select
            field={`ext.${ZHIYU_SCENE_TYPE_EXT_KEY}`}
            label="工作流模式"
            className="w-full"
            placeholder="请选择工作流模式"
            optionList={SCENE_TYPE_OPTIONS}
            initValue={zhiyuSceneType || undefined}
            showClear={false}
            rules={[
              { required: true, message: '请选择工作流模式' },
            ]}
            onChange={value => {
              // 确保选择后不能清空
              if (!value) {
                return;
              }
              onChange('ext', {
                ...(formValues.ext || {}),
                [ZHIYU_SCENE_TYPE_EXT_KEY]: value as string,
              });
              // 清空字段映射，因为工作流模式变了，需要重新获取
              onChange('evalTargetMapping', undefined);
            }}
          />

          {/* 版本显示（固定0.0.1） */}
          {workflowId ? (
            <Form.Slot label={I18n.t('version')}>
              <div className="flex items-center gap-2 h-8">
                <Tag color="cyan" size="small">
                  v{sourceTargetVersion}
                </Tag>
                <Typography.Text className="text-xs coz-fg-secondary">
                  发布版
                </Typography.Text>
              </div>
            </Form.Slot>
          ) : null}

          {/* 工作流描述 */}
          {workflowDetailService.data?.eval_target_content?.coze_workflow?.description ? (
            <Form.Slot label={I18n.t('description')}>
              <Typography.Text className="text-sm coz-fg-secondary">
                {workflowDetailService.data.eval_target_content.coze_workflow.description}
              </Typography.Text>
            </Form.Slot>
          ) : null}

          {/* 字段映射 */}
          <div className="evaluate-target-mapping-field-wrapper">
            <EvaluateTargetMappingField
              field="evalTargetMapping"
              prefixField="evalTargetMapping"
              label={EvaluateTargetMappingFieldLabel}
              evaluationSetSchemas={evaluationSetSchemas}
              rules={[
                {
                  required: true,
                  validator: (_, value) => {
                    if (inputSchemas?.length > 0 && isEmpty(value)) {
                      return new Error(
                        I18n.t('evaluate_please_configure_field_mapping'),
                      );
                    }
                    return true;
                  },
                  message: I18n.t('evaluate_please_configure_field_mapping'),
                },
              ]}
              loading={workflowDetailService.loading}
              keySchemas={inputSchemas}
              selectProps={{
                prefix: I18n.t('evaluation_set'),
              }}
            />
          </div>
        </>
      ) : null}
    </>
  );
};

export default WorkflowPluginEvalTargetForm;

// --------------end-----------------------
