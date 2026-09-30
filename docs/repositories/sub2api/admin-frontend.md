# sub2api 管理端代码文档

## 1. 定位

`services/sub2api_fork/frontend` 是 sub2api 自带的 Vue 3 管理界面，负责管理
模型账号、渠道、分组、用户、配额、支付、审计、风控和系统设置。

Clarkaitoy 的运营人员主要使用平台仓库的 `apps/admin_web`。sub2api 原生管理端
保留用于 AI 基础设施管理，不作为儿童和家庭运营入口。

## 2. 技术栈

| 层级 | 当前实现 |
| --- | --- |
| 框架 | Vue 3 |
| 路由 | Vue Router 4 |
| 状态 | Pinia |
| 构建 | Vite 5 |
| 语言 | TypeScript 5.6 |
| 样式 | Tailwind CSS 3 |
| 图表 | Chart.js + vue-chartjs |
| 国际化 | vue-i18n 9 |
| HTTP | Axios |
| 组件 | 自有组件 + VueUse + Vue Draggable Plus |
| 测试 | Vitest + Vue Test Utils |
| 包管理 | pnpm |

## 3. 目录结构

```text
frontend/
  src/
    api/                  API 客户端和类型
    assets/               图片和样式资源
    components/           通用组件
    composables/          组合式函数
    constants/            常量
    features/             功能模块
    i18n/                 国际化
    router/               路由和守卫
    stores/               Pinia store
    styles/               全局样式
    types/                TypeScript 类型
    utils/                工具函数
    views/                页面
  public/
  vite.config.ts
  package.json
  pnpm-lock.yaml
```

## 4. 主要页面

| 页面 | 功能 |
| --- | --- |
| Setup | 首次安装向导、数据库、Redis、管理员 |
| Login/Register | 登录、注册、OAuth、Passkey、2FA |
| Dashboard | 用户、用量、余额和系统概览 |
| API Keys | 创建、查看、禁用、额度和管理 |
| Usage | Token、图片、视频、模型和调用记录 |
| Subscriptions/Orders | 套餐、订单、支付和发票 |
| Admin Accounts | AI 账号、凭据、平台、分组和健康状态 |
| Admin Groups | 账号组、用户组、模型权限和配额 |
| Admin Channels | 渠道、定价、监控和上游配置 |
| Admin Users | 用户、余额、角色、配额和状态 |
| Admin Risk Control | 风控、Prompt 审计和安全策略 |
| Admin Audit | 管理操作、登录和敏感访问记录 |
| Admin Settings | 系统、安全、日志、支付和品牌设置 |

## 5. Clarkaitoy 定制建议

- 不把儿童、家庭、设备和 OTA 页面写进 sub2api 管理端。
- 增加 Clarkaitoy 租户、设备策略、调用来源和模型白名单视图。
- 增加按设备、用途、模型和策略版本的用量统计。
- 增加 AI 安全策略、Prompt 审计和异常调用告警。
- 通过内部 API 与 `device_platform` 同步最小化标签。
- 所有供应商密钥只以脱敏状态展示。

## 6. 路由和权限

路由守卫需要检查：

```text
requiresAuth
requiresAdmin
requiresPayment
requiresRiskControl
```

扩展权限时应使用统一导航和菜单配置，不在页面中散落角色判断。危险操作必须
二次确认，并写审计日志。

## 7. API 客户端

- Axios 实例统一处理 token、刷新、超时和错误。
- 所有响应由 TypeScript 类型约束。
- 禁止把敏感凭据写入 localStorage、URL 或日志。
- 页面不直接拼接请求 URL，统一通过 `src/api/`。
- 请求失败转换为可展示的错误状态。

## 8. 测试和质量

```powershell
Set-Location frontend
pnpm install --frozen-lockfile
pnpm run lint:check
pnpm run typecheck
pnpm run test:run
pnpm run build
```

重点关注：

- 登录、OAuth、Passkey 和 2FA。
- 账号、渠道、分组和模型路由。
- 支付、订单、订阅和余额。
- 用量、风控、审计和系统设置。
- 国际化 key 完整性和构建分包。

## 9. 提交前检查

- `pnpm-lock.yaml` 与 `package.json` 同步。
- 不提交 `node_modules/`、`dist/`、日志和本地配置。
- 检查 API 路径、权限和审计字段。
- 检查移动端和桌面端关键页面可用。
- 按 `<type>(<scope>): <summary>` 提交。
