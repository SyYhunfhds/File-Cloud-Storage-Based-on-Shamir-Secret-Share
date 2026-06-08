# 更新 API 文档计划

## 概要

更新 `2026-06-09.md`，反映审计操作与份额刷新分离后的真实 API 结构。

## 当前状态

### 审计操作 API (`audit/operate/`) [L169-199]
文档仍展示旧的 DTO（包含 `ItemId`、`RecoveryCode`、`DeviceShare`），但后端实际代码已将这些字段注释掉，只保留 `Operations`。响应也改为只返回 `Affected`。

### 份额刷新 API — 缺失
文档完全不包含 `share/refresh/` 路由说明。后端实现已完成，使用 SSE 推送进度。

### 份额刷新 API 行为
- Path: `POST v1/protected/share/refresh/`
- 传输: SSE (Server-Sent Events)，Content-Type: `text/event-stream`
- SSE 消息体: JSON，结构为 `Msg { progress, message, data? }`
- 只需展示完整的进度阶段列表，前端的职责：监听 SSE 流、等待进度100的 data 即为最终响应

### SSE 进度阶段
| progress | message | 说明 |
|----------|---------|------|
| - | 解析用户Claims | 纯文本消息 |
| 10 | 用户Claims解析成功 | |
| 20 | 开始校验参数有效性 | |
| 40 | 开始提取份额并进行密钥重建 | |
| 45 | Auth Share可能已失效 | **仅错误时推送** |
| 50 | 查找其他用户的坐标, 用于计算份额 | |
| 55 | 无法获取Member的信息, 将不会为其他用户计算份额 | **仅无其他用户时推送** |
| 60 | 密钥重建结束 | |
| 70 | 重新计算份额 | |
| 75 | 份额加密失败 | **仅错误时推送** |
| 90 | 份额入库完毕 | |
| 100 | 请求处理完毕 | 含最终 data |

注：任何未到进度100的传输都是后端处理异常，与前端实现无关。

### DTO

```go
// Req
type ShareRefreshReq struct {
    ItemId       int    `json:"item_id"`
    RecoveryCode string `json:"recovery_code"` // 可选
    DeviceShare  string `json:"share"`          // 可选（但RecoveryCode和DeviceShare至少提供一个）
}

// Res (仅进度100时在 data 字段中返回)
type ShareRefreshRes struct {
    DeviceShare              string `json:"device_share"`
    IsRecoveryCodeReGenerated bool   `json:"is_recovery_code_re_generated"`
    RecoveryCode              string `json:"recovery_code"`
    RecoveryShare             string `json:"recovery_share"`
}
```

## 改动清单

### 文件: `backend/api/2026-06-09.md`

#### 1. 更新审计操作API [L169-199]
- 移除 "当前API未完成" 说明
- 更新 `AuditOperationReq` DTO：移除 `ItemId`、`RecoveryCode`、`DeviceShare`
- 更新 `AuditOperationRes` DTO：用 `Affected` 替换旧字段
- 保留 `Operations` 的定义不变

#### 2. 在审计操作API之后新增份额刷新API
- 添加一级标题 `#### 份额链路 share/`
- 添加二级标题 `##### 份额刷新API refresh/`
- 添加 SSE 说明和进度表
- 添加 Req/Res DTO
- 添加 SSE 响应示例（正常完成和错误中断两种）
