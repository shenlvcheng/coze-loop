// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 工作流评测对象视图组件

import { I18n } from '@cozeloop/i18n-adapter';
import { Tag } from '@coze-arch/coze-design';

import { type CreateExperimentValues } from '../../../types/evaluate-target';

/**
 * 智宇工作流评测对象视图
 */
export const WorkflowEvalTargetView = (props: {
  formValues: CreateExperimentValues;
}) => {
  const { formValues } = props;

  // 工作流 sceneKey
  const workflowId = formValues.evalTarget || '';

  // 工作流版本 (固定 0.0.1)
  const workflowVersion = formValues.evalTargetVersion || '0.0.1';

  // 工作流名称
  const workflowName = workflowId || '-';

  return (
    <>
      <div className="text-[16px] leading-[22px] font-medium coz-fg-primary mb-5">
        {I18n.t('evaluation_object')}
      </div>
      <div className="flex flex-row gap-5">
        <div className="flex-1 w-0">
          <div className="text-sm font-medium coz-fg-primary mb-2">
            {I18n.t('type')}
          </div>
          <div className="text-sm font-normal coz-fg-primary">
            {I18n.t('zhiyu_workflow', '智宇工作流')}
          </div>
        </div>
        <div className="flex-1 w-0 mb-4">
          <div className="text-sm font-medium coz-fg-primary mb-2">
            {I18n.t('name_and_version')}
          </div>
          <div className="flex flex-row items-center gap-1">
            <div className={'text-sm font-normal coz-fg-primary'}>
              {workflowName}
            </div>
            <Tag color="cyan" className="!h-5 !px-2 !py-[2px] rounded-[3px]">
              {workflowVersion}
            </Tag>
          </div>
        </div>
      </div>
    </>
  );
};

// --------------end-----------------------
