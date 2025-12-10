// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0
import { useMemo, type FC } from 'react';

import { I18n } from '@cozeloop/i18n-adapter';
import { type FieldSchema } from '@cozeloop/api-schema/evaluation';
import { IconCozEmpty } from '@coze-arch/coze-design/icons';
import {
  EmptyState,
  Input,
  Loading,
  Tag,
  type SelectProps,
  withField,
  type CommonFieldProps,
} from '@coze-arch/coze-design';

import { type OptionGroup } from '../../../components/mapping-item-field/types';
import { MappingItemField } from '../../../components/mapping-item-field';
import {
  EqualItem,
  ReadonlyItem,
  getSchemaTypeText,
  getTypeText,
} from '../../../components/column-item-map';

import emptyStyles from './empty-state.module.less';

export interface EvaluateTargetMappingProps {
  loading?: boolean;
  keySchemas?: (FieldSchema & { type?: string; default_value?: string })[];
  prefixField: string;
  evaluationSetSchemas?: FieldSchema[];
  selectProps?: SelectProps;
}

const EvaluateTargetMappingField: FC<
  CommonFieldProps & EvaluateTargetMappingProps
> = withField((props: EvaluateTargetMappingProps) => {
  const {
    loading,
    keySchemas,
    prefixField,
    evaluationSetSchemas,
    selectProps,
  } = props;

  const optionGroups = useMemo(
    () =>
      evaluationSetSchemas
        ? [
            {
              schemaSourceType: 'set',
              children: evaluationSetSchemas?.map(s => ({
                ...s,
                schemaSourceType: 'set',
              })),
            } satisfies OptionGroup,
          ]
        : [],
    [evaluationSetSchemas],
  );

  if (!keySchemas) {
    return (
      <div className="h-[84px] w-full flex items-center justify-center">
        <EmptyState
          size="default"
          icon={<IconCozEmpty className="coz-fg-dim text-32px" />}
          title={I18n.t('no_data')}
          className={emptyStyles['empty-state']}
        />
      </div>
    );
  }
  return (
    <>
      <div className={loading ? 'hidden' : ''}>
        {keySchemas?.map(k => {
          // 有默认值的字段显示为只读输入框
          const hasDefaultValue = !!(k as any).default_value;
          // 根据 isRequired 决定是否必填校验
          const isRequired = k.isRequired !== false; // 默认必填，除非明确设置为 false

          // 有默认值的字段：显示为只读输入框
          if (hasDefaultValue) {
            return (
              <div key={k.name} className="flex flex-row items-center gap-2 mb-3">
                <ReadonlyItem
                  className="flex-1"
                  title={I18n.t('evaluation_object')}
                  typeText={getSchemaTypeText(k)}
                  value={k.name}
                />
                <EqualItem />
                <Input
                  className="flex-1"
                  value={(k as any).default_value}
                  disabled
                  suffix={
                    <Tag size="small" color="cyan">
                      默认值
                    </Tag>
                  }
                />
              </div>
            );
          }

          // 没有默认值的字段：显示为下拉选择框
          return (
            <MappingItemField
              key={k.name}
              noLabel
              field={`${prefixField}.${k.name}`}
              fieldClassName="!pt-0"
              keyTitle={I18n.t('evaluation_object')}
              keySchema={k}
              optionGroups={optionGroups}
              selectProps={selectProps}
              rules={[
                {
                  validator: (_rule, v) => {
                    // 只有必填字段才校验空值
                    if (isRequired && !v) {
                      return new Error(I18n.t('please_select'));
                    }
                    if (v && getTypeText(v) !== getSchemaTypeText(k)) {
                      return new Error(I18n.t('selected_fields_inconsistent'));
                    }
                    return true;
                  },
                },
              ]}
            />
          );
        })}
      </div>
      {/* loading态不能直接返回咯爱的loading dom，不渲染MappingItemField会导致表单数据丢失 */}
      {loading ? (
        <div className="h-[84px] w-full flex items-center justify-center">
          <Loading
            className="!w-full"
            size="large"
            label={I18n.t('loading_field_mapping')}
            loading={true}
          />
        </div>
      ) : null}
    </>
  );
});
export default EvaluateTargetMappingField;
