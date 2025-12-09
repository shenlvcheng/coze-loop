// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 字段映射预览组件

import { I18n } from '@cozeloop/i18n-adapter';
import { type FieldSchema } from '@cozeloop/api-schema/evaluation';

import { ReadonlyItem, EqualItem } from '../../../components/column-item-map';

interface WorkflowFieldMappingPreviewProps {
  evalTargetMapping?: Record<string, FieldSchema>;
}

export const WorkflowFieldMappingPreview = ({
  evalTargetMapping,
}: WorkflowFieldMappingPreviewProps) => {
  if (!evalTargetMapping) {
    return null;
  }

  const mappingEntries = Object.entries(evalTargetMapping);

  if (mappingEntries.length === 0) {
    return <div className="text-gray-400">{I18n.t('no_field_mapping')}</div>;
  }

  return (
    <div className="flex flex-col gap-2">
      {mappingEntries.map(([key, value]) => (
        <div key={key} className="flex items-center gap-2">
          <ReadonlyItem title={I18n.t('evaluation_object')} value={key} typeText="String" />
          <EqualItem />
          <ReadonlyItem
            title={I18n.t('evaluation_set')}
            value={value?.name || '-'}
            typeText={value?.text_schema ? JSON.parse(value.text_schema)?.type : 'String'}
          />
        </div>
      ))}
    </div>
  );
};

// --------------end-----------------------
