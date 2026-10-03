# OTA Feature

Owns firmware releases, rollout groups, progress, rollback, and failure
statistics.

## 版本登记

版本号默认来自仓库根目录的 `VERSION`，构建时由 Vite 注入。选择版本、更新
类型和适用平台后，页面通过 `/api/v1/admin/release-artifacts/{version}` 查询
已经上传到下载服务的文件，并自动填入下载地址与 SHA-256 校验值。

自动查询失败时保留人工填写入口，避免下载服务暂时不可用时无法登记版本。
