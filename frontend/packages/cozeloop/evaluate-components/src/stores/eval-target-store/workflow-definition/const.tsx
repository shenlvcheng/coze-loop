// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 前端定义

import { EvalTargetType } from '@cozeloop/api-schema/evaluation';

import {
  type CreateExperimentValues,
  ExtCreateStep,
  type EvalTargetDefinition,
} from '../../../types/evaluate-target';
import WorkflowTargetPreview from './workflow-target-preview';
import { WorkflowFieldMappingPreview } from './workflow-field-mapping-preview';
import { WorkflowEvalTargetView } from './workflow-eval-target-view';
import WorkflowPluginEvalTargetForm from './workflow-plugin-eval-target-form';

const getEvalTargetValidFields = (values: CreateExperimentValues) => {
  const { evalTargetMapping = {} } = values;
  const result = ['evalTarget', 'evalTargetVersion', 'evalTargetMapping'];

  Object.keys(evalTargetMapping).forEach(key => {
    result.push(`evalTargetMapping.${key}`);
  });
  return result;
};

export const workflowEvalTargetDefinitionPayload: EvalTargetDefinition = {
  type: EvalTargetType.CozeWorkflow,
  name: '智宇工作流',
  selector: undefined,
  preview: WorkflowTargetPreview,
  extraValidFields: {
    [ExtCreateStep.EVAL_TARGET]: getEvalTargetValidFields,
  },
  evalTargetFormSlotContent: WorkflowPluginEvalTargetForm,
  evalTargetView: WorkflowEvalTargetView,
  viewSubmitFieldMappingPreview: WorkflowFieldMappingPreview,
  targetInfo: {
    color: 'cyan',
    tagColor: 'cyan',
  },
};

// --------------end-----------------------
