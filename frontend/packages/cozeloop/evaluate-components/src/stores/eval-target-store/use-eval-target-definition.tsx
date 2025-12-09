// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

import { type EvalTargetDefinition } from '../../types/evaluate-target';
import { promptEvalTargetDefinitionPayload } from './prompt-definition/const';
// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 导入工作流定义
import { workflowEvalTargetDefinitionPayload } from './workflow-definition/const';
// --------------end-----------------------
// import { evalSetDefinitionPayload } from './eval-set-definition/const';

// 根据 type 注册
const evalTargetDefinitionMap = new Map<
  string | number,
  EvalTargetDefinition
>();

/**
 * 注册评测对象选择器
 * @returns 注销评测对象选择器 方法
 */
const registerEvalTargetDefinition = (item: EvalTargetDefinition) => {
  evalTargetDefinitionMap.set(item.type, item);
  return () => {
    evalTargetDefinitionMap.delete(item.type);
  };
};

// --------------start----------------------
// 新增代码人: Cascade
// 新增代码原因: 智宇工作流评估功能 - 在模块加载时立即注册定义，确保在组件渲染前就可用
registerEvalTargetDefinition(promptEvalTargetDefinitionPayload);
registerEvalTargetDefinition(workflowEvalTargetDefinitionPayload);
// registerEvalTargetDefinition(evalSetDefinitionPayload);
// --------------end-----------------------

/**
 * 获取 type 评测对象选择器
 * @param type 评测对象的value
 * @returns 评测对象选择器
 */
const getEvalTargetDefinition = (type: string | number) =>
  evalTargetDefinitionMap.get(type);

const getEvalTargetDefinitionList = () =>
  Array.from(evalTargetDefinitionMap.values());

/**
 * 评测对象选择器
 */
export const useEvalTargetDefinition = () => {
  return {
    getEvalTargetDefinition,
    registerEvalTargetDefinition,
    getEvalTargetDefinitionList,
  };
};
