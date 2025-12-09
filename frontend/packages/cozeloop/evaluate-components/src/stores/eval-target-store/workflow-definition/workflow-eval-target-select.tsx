// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 工作流选择器组件

import { useDebounceFn, useRequest } from 'ahooks';
import { I18n } from '@cozeloop/i18n-adapter';
import { BaseSearchSelect } from '@cozeloop/components';
import { useSpace } from '@cozeloop/biz-hooks-adapter';
import { EvalTargetType } from '@cozeloop/api-schema/evaluation';
import { StoneEvaluationApi } from '@cozeloop/api-schema';
import { type SelectProps, Typography } from '@coze-arch/coze-design';

const ellipsis = {
  showTooltip: true,
};

/**
 * 智宇工作流选择器
 */
const WorkflowEvalTargetSelect = ({
  onlyShowOptionName = false,
  ...props
}: SelectProps & { onlyShowOptionName?: boolean }) => {
  const { spaceID } = useSpace();

  const service = useRequest(async (text?: string) => {
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
        label: onlyShowOptionName ? (
          <Typography.Text ellipsis={ellipsis}>{subTitle}</Typography.Text>
        ) : (
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

  const handleSearch = useDebounceFn(service.run, {
    wait: 500,
  });

  return (
    <BaseSearchSelect
      className={props.className}
      emptyContent={I18n.t('no_data')}
      loading={service.loading}
      onSearch={handleSearch.run}
      showRefreshBtn={true}
      onClickRefresh={() => service.run()}
      optionList={service.data}
      {...props}
    />
  );
};

export default WorkflowEvalTargetSelect;

// --------------end-----------------------
