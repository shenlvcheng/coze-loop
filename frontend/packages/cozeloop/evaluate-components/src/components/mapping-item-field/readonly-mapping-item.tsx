// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0
import { type FieldSchema } from '@cozeloop/api-schema/evaluation';

import {
  EqualItem,
  getSchemaTypeText,
  getTypeText,
  ReadonlyItem,
} from '../column-item-map';
import { schemaSourceTypeMap, type OptionSchema } from './types';

export function ReadonlyMappingItem({
  keyTitle,
  keySchema,
  optionSchema,
}: {
  keyTitle?: string;
  keySchema?: FieldSchema;
  optionSchema?: OptionSchema;
}) {
  // 如果是默认值字段，显示 constValue 而不是 name
  const displayValue =
    (optionSchema as any)?.schemaSourceType === '__default_value__'
      ? (optionSchema as any)?.constValue || optionSchema?.name
      : optionSchema?.name;

  return (
    <div className="flex flex-row items-center gap-2">
      <ReadonlyItem
        className="flex-1 basis-80 overflow-hidden"
        title={keyTitle}
        typeText={getSchemaTypeText(keySchema)}
        value={keySchema?.name}
      />
      <EqualItem />
      <ReadonlyItem
        className="flex-1 basis-80 overflow-hidden"
        title={
          optionSchema?.schemaSourceType &&
          schemaSourceTypeMap[optionSchema.schemaSourceType as keyof typeof schemaSourceTypeMap]
        }
        typeText={getTypeText(optionSchema)}
        value={displayValue}
      />
    </div>
  );
}
