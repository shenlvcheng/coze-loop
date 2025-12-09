// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 字段映射预览组件

import { I18n } from '@cozeloop/i18n-adapter';

import { type CreateExperimentValues } from '@/types/evaluate-target';
import { ReadonlyMappingItem } from '@/components/mapping-item-field/readonly-mapping-item';

export function WorkflowFieldMappingPreview({
  createExperimentValues,
}: {
  /** 渲染数据 */
  createExperimentValues: CreateExperimentValues;
}) {
  const { evalTargetMapping } = createExperimentValues ?? {};

  if (!evalTargetMapping) {
    return null;
  }

  const mappingEntries = Object.entries(evalTargetMapping);

  if (mappingEntries.length === 0) {
    return null;
  }

  return (
    <div className="flex flex-col gap-3">
      {mappingEntries.map(([key, optionSchema]) => (
        <ReadonlyMappingItem
          key={key}
          keyTitle={I18n.t('evaluation_object')}
          keySchema={{ key, name: key, type: 'string' }}
          optionSchema={optionSchema}
        />
      ))}
    </div>
  );
}

// --------------end-----------------------
