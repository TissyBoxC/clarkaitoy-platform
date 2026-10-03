# System Feature

Owns feature switches, dictionaries, notification settings, and third-party
integration status.

## AI 服务默认值

页面加载时会从 `/api/v1/admin/ai-account-defaults` 读取由中转站统一维护的
初始余额和默认并发数，避免管理端与 sub2api 出现两套默认值。默认模型从
`/api/v1/admin/ai-models` 获取，并自动选择延迟最低的可用模型；运营人员仍可
在列表中调整。

两个接口不可用时，页面保留上一次保存的设置并给出提示，不会自行推算余额或
伪造模型延迟。
