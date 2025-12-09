// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 工作流预览组件

import { type EvalTarget } from '@cozeloop/api-schema/evaluation';
import { Typography, Tag } from '@coze-arch/coze-design';

import { BaseTargetPreview } from '../base-target-preview';

interface WorkflowTargetPreviewProps {
  evalTarget?: EvalTarget;
}

const WorkflowTargetPreview = ({ evalTarget }: WorkflowTargetPreviewProps) => {
  const etc = evalTarget?.eval_target_version?.eval_target_content;
  const workflow = etc?.coze_workflow;

  return (
    <BaseTargetPreview
      name={
        <div className="flex items-center gap-2">
          <Typography.Text className="font-medium">
            {workflow?.name || workflow?.id || '-'}
          </Typography.Text>
          <Tag color="cyan" size="small">
            智宇工作流
          </Tag>
        </div>
      }
      version={evalTarget?.eval_target_version?.source_target_version || '0.0.1'}
    />
  );
};

export default WorkflowTargetPreview;

// --------------end-----------------------
