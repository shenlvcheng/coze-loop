// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 工作流表单组件

import { useEffect, useMemo } from 'react';

import { isEmpty } from 'lodash-es';
import { useDebounceFn, useRequest } from 'ahooks';
import { I18n } from '@cozeloop/i18n-adapter';
import { useSpace } from '@cozeloop/biz-hooks-adapter';
import { EvalTargetType, type FieldSchema } from '@cozeloop/api-schema/evaluation';
import { StoneEvaluationApi } from '@cozeloop/api-schema';
import { IconCozInfoCircle } from '@coze-arch/coze-design/icons';
import { Form, FormSelect, Tag, Tooltip, Typography } from '@coze-arch/coze-design';

import { type PluginEvalTargetFormProps, type OptionSchema } from '@/types/evaluate-target';
import { EvaluateTargetMappingField } from '@/components/selectors/evaluate-target';

const ellipsis = {
  showTooltip: true,
};

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

  // 评测集字段
  const evaluationSetSchemas =
    createExperimentValues?.evaluationSetVersionDetail?.evaluation_set_schema
      ?.field_schemas;

  // 获取工作流列表
  const workflowListService = useRequest(async (text?: string) => {
    const res = await StoneEvaluationApi.ListSourceEvalTargets({
      target_type: EvalTargetType.CozeWorkflow,
      name: text || undefined,
      workspace_id: spaceID,
      page_size: 100,
    });
    return res.eval_targets?.map(item => {
      const etc = item.eval_target_version?.eval_target_content;
      const title = etc?.coze_workflow?.id || '';
      const subTitle = etc?.coze_workflow?.name || '';

      return {
        value: item.source_target_id,
        label: (
          <div className="flex flex-row items-center w-full overflow-hidden">
            <Typography.Text
              className={'flex-shrink !max-w-[600px] text-[13px]'}
              ellipsis={ellipsis}
            >
              {subTitle}
            </Typography.Text>
            <Typography.Text
              className={'flex-1 w-0 ml-3 text-xs font-medium coz-fg-secondary'}
              ellipsis={ellipsis}
            >
              {title}
            </Typography.Text>
          </div>
        ),
        ...item,
      };
    });
  });

  const handleWorkflowSearch = useDebounceFn(workflowListService.run, {
    wait: 500,
  });

  // 获取工作流详情（包含参数）
  const workflowDetailService = useRequest(
    async () => {
      if (!workflowId) return null;
      const res = await StoneEvaluationApi.ListSourceEvalTargetVersions({
        workspace_id: spaceID,
        source_target_id: workflowId,
        target_type: EvalTargetType.CozeWorkflow,
        page_size: 1,
      });
      return res.versions?.[0];
    },
    {
      refreshDeps: [workflowId],
      ready: !!workflowId,
      onError: () => {
        // 报错时清空字段映射，避免显示旧数据
        onChange('evalTargetMapping', undefined);
      },
    },
  );

  // 从工作流详情中获取输入参数schema
  // 所有字段都展示，有 default_value 的字段显示为只读输入框
  const inputSchemas = useMemo(() => {
    const schemas = workflowDetailService.data?.eval_target_content?.input_schemas;
    if (!schemas) return [];
    return schemas.map(schema => ({
      key: schema.key || '',
      name: schema.key || '',
      text_schema: schema.json_schema,
      isRequired: schema.is_required,
      default_value: schema.default_value, // 有默认值的字段显示为只读输入框
    })) as FieldSchema[];
  }, [workflowDetailService.data]);

  const handleEvalTargetChange = () => {
    onChange('evalTargetVersion', '0.0.1'); // 固定版本
    onChange('evalTargetMapping', undefined);
  };

  // 当输入参数变化时，初始化字段映射
  useEffect(() => {
    // 只有在成功获取到数据且没有错误时才初始化字段映射
    if (inputSchemas?.length > 0 && !workflowDetailService.error) {
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
          {/* 工作流选择 */}
          <FormSelect
            className="w-full"
            field="evalTarget"
            label="工作流名称"
            placeholder={I18n.t('please_select')}
            rules={[
              { required: true, message: I18n.t('please_select') },
            ]}
            onChange={handleEvalTargetChange}
            filter={true}
            loading={workflowListService.loading}
            optionList={workflowListService.data}
            onSearch={handleWorkflowSearch.run}
            showClear={true}
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
