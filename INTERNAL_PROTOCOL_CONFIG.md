# 内部协议大模型配置指南

本文档说明如何在 `model_config.yaml` 中配置内部协议大模型。

## 📋 配置示例

在 `model_config.yaml` 文件中添加以下配置：

```yaml
models:
  - id: 100  # 模型ID，确保唯一
    name: "内部Qwen模型"  # 模型展示名称
    desc: "企业内部部署的通义千问模型"  # 模型描述
    frame: "eino"  # 框架类型，固定为 eino
    protocol: "internal"  # 协议类型，使用内部协议

    # 协议配置
    protocol_config:
      # 基础URL配置
      base_url: "http://10.32.41.228:39080"  # 内部网关的基础URL
      model: "qwen"  # 模型名称
      timeout_ms: 60000  # 请求超时时间（毫秒），可选，默认60秒

      # 内部协议特有配置
      protocol_config_internal:
        # HTTP请求头配置
        ai_api_code: "7xU7Aa1IMj"     # AI-API-CODE header
        ai_app_key: "VCRZouogAZ"      # AI-APP-KEY header
        caller_token: "RxZ9lJpP8I"    # CALLER-TOKEN header
        description: "评估服务使用"   # description header

        # 请求Body额外参数
        process_code: "7xU7Aa1IMj"   # processCode 字段
        app_id: "109110100301"        # appId 字段
        access_token: "2526170e9a4d474c9ccff8c5a7ee323b"  # accessToken 字段

    # 模型能力配置
    ability:
      function_call: true   # ✅ 支持函数调用
      multi_modal: false    # 是否支持多模态输入

    # 参数配置（定义用户可调整的参数）
    param_config:
      param_schemas:
        - name: "temperature"
          label: "温度"
          desc: "控制输出的随机性。值越大输出越随机，值越小输出越确定"
          type: "float"
          min: "0"
          max: "1.0"
          default_val: "0.7"

        - name: "max_tokens"
          label: "最大Token数"
          desc: "控制模型输出的最大token数量"
          type: "int"
          min: "1"
          max: "4096"
          default_val: "2048"

        - name: "top_p"
          label: "Top P"
          desc: "核采样参数，控制输出的多样性"
          type: "float"
          min: "0.001"
          max: "1.0"
          default_val: "0.9"

    # 场景配置（可选，用于限流）
    scenario_configs:
      evaluator:  # 评估场景
        quota:
          qpm: 100  # 每分钟请求数限制，-1表示不限制
          tpm: 100000  # 每分钟token数限制，-1表示不限制
        unavailable: false  # 是否在该场景下不可用
```

## 🔧 配置说明

### 必需字段

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `base_url` | 内部网关的基础URL | `http://10.32.41.228:39080` |
| `model` | 模型名称 | `qwen` |
| `ai_api_code` | API认证码（AI-API-CODE header） | `7xU7Aa1IMj` |
| `ai_app_key` | 应用密钥（AI-APP-KEY header） | `VCRZouogAZ` |
| `caller_token` | 调用者令牌（CALLER-TOKEN header） | `RxZ9lJpP8I` |
| `description` | 使用渠道描述（description header） | `评估服务使用` |
| `process_code` | 进程代码（请求body字段） | `7xU7Aa1IMj` |
| `app_id` | 应用ID（请求body字段） | `109110100301` |
| `access_token` | 访问令牌（请求body字段） | `2526170e9a4d474c9ccff8c5a7ee323b` |

### 可选字段

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `timeout_ms` | 请求超时时间（毫秒） | `60000` |

## 📡 实际请求示例

### 普通聊天请求

配置完成后，系统会发送如下格式的HTTP请求：

**请求URL**:
```
POST http://10.32.41.228:39080/v1/chat/completions
```

**请求Headers**:
```
Content-Type: application/json;charset=UTF-8
AI-API-CODE: 7xU7Aa1IMj
AI-APP-KEY: VCRZouogAZ
CALLER-TOKEN: RxZ9lJpP8I
description: 评估服务使用
```

**请求Body**:
```json
{
  "processCode": "7xU7Aa1IMj",
  "appId": "109110100301",
  "accessToken": "2526170e9a4d474c9ccff8c5a7ee323b",
  "stream": true,
  "model": "qwen",
  "max_tokens": 2048,
  "temperature": 0.7,
  "top_p": 0.9,
  "messages": [
    {
      "role": "user",
      "content": "你好"
    }
  ]
}
```

### Function Calling 请求

当评估器使用函数调用时，请求body会额外包含 `tools` 字段：

**请求Body**:
```json
{
  "processCode": "7xU7Aa1IMj",
  "appId": "109110100301",
  "accessToken": "2526170e9a4d474c9ccff8c5a7ee323b",
  "stream": false,
  "model": "qwen",
  "max_tokens": 2048,
  "temperature": 0.7,
  "messages": [
    {
      "role": "user",
      "content": "评估这个回答的质量"
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "submit_score",
        "description": "提交评分结果",
        "parameters": {
          "type": "object",
          "properties": {
            "score": {
              "type": "number",
              "description": "评分，0-100"
            },
            "reason": {
              "type": "string",
              "description": "评分理由"
            }
          },
          "required": ["score", "reason"]
        }
      }
    }
  ]
}
```

**响应示例**（Function Calling）:
```json
{
  "id": "endpoint_common_xxx",
  "object": "chat.completion",
  "created": 1764830154,
  "model": "qwen",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "",
        "tool_calls": [
          {
            "id": "call_abc123",
            "type": "function",
            "function": {
              "name": "submit_score",
              "arguments": "{\"score\": 85, \"reason\": \"回答准确且详细\"}"
            }
          }
        ]
      },
      "finish_reason": "tool_calls"
    }
  ],
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 30,
    "total_tokens": 80
  }
}
```

## ✅ 验证配置

配置完成后，重启服务使配置生效：

```bash
# Docker Compose部署
cd release/deployment/docker-compose
docker-compose restart

# Kubernetes部署
kubectl rollout restart deployment/your-deployment-name
```

检查日志确认模型加载成功：

```bash
# 查看日志
docker-compose logs -f | grep "内部Qwen模型"
```

## 🎯 在评估中使用

配置完成后，在创建评估器时选择"内部Qwen模型"即可使用。

### 提示词评估器示例

创建提示词评估器时，可以使用Function Calling获取结构化输出：

```yaml
评估器配置:
  名称: 质量评分器
  类型: 提示词评估器
  模型: 内部Qwen模型
  提示词: |
    请评估以下回答的质量：
    {{output}}

    评分标准：
    - 准确性（0-50分）
    - 完整性（0-30分）
    - 流畅度（0-20分）

  输出解析: Function Calling
  工具定义:
    - name: submit_score
      description: 提交评分结果
      parameters:
        score: 评分（0-100）
        reason: 评分理由
```

## ⚠️ 注意事项

1. **安全性**: 确保 `access_token`、`ai_api_code` 等敏感信息妥善保管
2. **网络连通性**: 确保评估服务能够访问内部网关地址（`http://10.32.41.228:39080`）
3. **Token限制**: 根据实际需求调整 `qpm` 和 `tpm` 限制
4. **超时设置**: 根据模型响应速度调整 `timeout_ms`
5. **Function Calling**: ✅ 内部协议**完全支持** function calling
6. **流式响应**: ✅ 支持SSE流式响应
7. **响应格式**: 完全兼容OpenAI格式

## 🔍 故障排查

### 问题1: 连接超时
- 检查网络连通性: `curl http://10.32.41.228:39080/v1/chat/completions`
- 调整 `timeout_ms` 参数

### 问题2: 认证失败
- 确认 `ai_api_code`、`ai_app_key`、`caller_token` 配置正确
- 检查 `access_token` 是否过期

### 问题3: 模型不可用
- 检查日志: `docker-compose logs -f`
- 确认 `model` 字段值与内部网关支持的模型名称一致

### 问题4: Function Calling 不工作
- 确认模型能力配置中 `function_call: true`
- 检查工具定义格式是否正确
- 查看模型响应中是否包含 `tool_calls` 字段

## 📚 相关文件

- 协议定义: `backend/modules/llm/domain/entity/manage.go:365`
- 配置结构: `backend/modules/llm/domain/entity/manage.go:244-252`
- HTTP客户端实现: `backend/modules/llm/domain/service/llmimpl/internal/client.go`
- Builder函数: `backend/modules/llm/domain/service/llmimpl/eino/init.go:448-496`

## 🆚 与其他协议的对比

| 特性 | OpenAI协议 | 内部协议 |
|------|-----------|---------|
| 认证方式 | `Authorization: Bearer xxx` | 4个自定义headers |
| 请求Body | 标准OpenAI格式 | 额外3个字段 |
| 响应格式 | OpenAI格式 | ✅ 完全兼容 |
| Function Calling | ✅ 支持 | ✅ 支持 |
| 流式响应 | ✅ 支持 | ✅ 支持 |
| 多模态 | ✅ 支持 | 根据实际模型 |

## 🎉 优势

✅ **无缝集成** - 响应格式完全兼容OpenAI，复用成熟解析器
✅ **灵活认证** - 支持多个自定义认证header
✅ **功能完整** - 完全支持Function Calling和流式响应
✅ **易于配置** - 所有参数通过YAML配置管理
✅ **生产就绪** - 包含完善的错误处理和超时控制
