# 后台管理端开发文档

## 1. 项目定位

后台管理端是 Clarkaitoy 内部运营和管理人员使用的 Web 控制台。

它负责以下核心事情：

- 账号和权限管理
- 设备和游戏机生命周期管理
- 儿童档案和家庭信息查看
- 内容包、主题包和素材管理
- AI 模型、通道、配额和计费配置
- 固件版本和 OTA 发布
- 设备日志、审计和质量监控
- 隐私、安全和合规操作

后台管理端不直接与 ESP32 通信，也不直接在浏览器中保存模型供应商密钥。

## 2. 在项目中的组成

```text
apps/
  admin_web/
    src/
      app/                 应用入口、路由、权限初始化
      api/                 HTTP 客户端和接口封装
      components/          通用组件
      layouts/             管理台布局
      features/
        auth/
        dashboard/
        family/
        child/
        device/
        content/
        ai_gateway/
        ota/
        audit/
        system/
      stores/              全局状态
      types/               共享类型
      utils/
      i18n/
    tests/
```

## 3. 技术栈

| 层级 | 推荐技术 |
| --- | --- |
| 前端框架 | Vue 3 或 React |
| 语言 | TypeScript |
| 构建 | Vite |
| UI 组件 | Element Plus 或 Ant Design |
| 状态管理 | Pinia 或 Zustand |
| 路由 | Vue Router 或 React Router |
| 网络 | Axios 或 TanStack Query |
| 表格 | TanStack Table |
| 表单 | Vue Formulate 或 React Hook Form |
| 图表 | ECharts |
| 国际化 | vue-i18n 或 i18next |
| 测试 | Vitest + Playwright |

如果与 `sub2api` 的 Vue 管理界面共存，建议统一使用 Vue 3 + TypeScript + Vite + Element Plus，减少多个前端技术栈的维护成本。

## 4. 推荐实用工具包

### Vue 路线

| 包名 | 用途 |
| --- | --- |
| `element-plus` | 管理后台组件库 |
| `@element-plus/icons-vue` | 图标 |
| `pinia` | 状态管理 |
| `vue-router` | 路由 |
| `axios` | HTTP 请求 |
| `@tanstack/vue-query` | 请求缓存和异步状态 |
| `@tanstack/vue-table` | 高级表格 |
| `echarts` | 图表 |
| `vue-i18n` | 国际化 |
| `dayjs` | 时间处理 |
| `zod` | 数据校验 |
| `unocss` 或 `tailwindcss` | 样式工具 |
| `vueuse` | 常用组合式工具 |

### React 路线

| 包名 | 用途 |
| --- | --- |
| `antd` | 管理后台组件库 |
| `@ant-design/icons` | 图标 |
| `zustand` | 状态管理 |
| `react-router-dom` | 路由 |
| `axios` | HTTP 请求 |
| `@tanstack/react-query` | 请求缓存 |
| `@tanstack/react-table` | 表格 |
| `echarts` | 图表 |
| `react-hook-form` | 表单 |
| `zod` | 数据校验 |
| `dayjs` | 时间处理 |
| `i18next` | 国际化 |

### 质量工具

| 包名 | 用途 |
| --- | --- |
| `eslint` | 代码检查 |
| `prettier` | 格式化 |
| `vitest` | 单元测试 |
| `@playwright/test` | 端到端测试 |
| `msw` | API mock |
| `commitlint` | Commit 规则 |

## 5. 功能模块

| 模块 | 功能 | 优先级 |
| --- | --- | --- |
| 登录 | 账号密码、MFA、会话管理、SSO 预留 | P0 |
| 权限 | 角色、菜单权限、数据权限、操作权限 | P0 |
| 仪表盘 | 设备总数、在线率、AI 调用、错误趋势、OTA 状态 | P0 |
| 家庭管理 | 家庭列表、成员、设备关联、状态 | P0 |
| 儿童档案 | 年龄、内容等级、使用策略、数据删除 | P0 |
| 设备管理 | 设备注册、型号、固件、在线状态、远程配置、解绑 | P0 |
| 内容管理 | 音频、文本、图片、主题包、分龄、审核、发布 | P0 |
| AI 网关 | 模型、账号池、路由、限流、配额、调用日志 | P0 |
| OTA | 固件包、灰度、设备组、发布、回滚、失败统计 | P0 |
| 家长策略 | 查看和协助配置策略、异常策略审核 | P1 |
| 审计 | 管理操作记录、登录记录、敏感操作追踪 | P0 |
| 监控 | 设备、服务、模型、网络、告警和告警规则 | P1 |
| 系统设置 | 参数、字典、通知、第三方服务、功能开关 | P1 |

## 6. 与 sub2api 的关系

后台管理端不应把 `sub2api` 的全部管理能力重新实现一遍。

建议：

- `sub2api` 继续负责 AI 账号、渠道、模型、配额、计费和调用统计。
- Clarkaitoy 的 Web 管理台负责儿童产品业务、设备和内容运营。
- 通过后台服务的 `ai_gateway` 适配层调用 `sub2api` 管理 API。
- 不建议让运营人员直接登录 `sub2api` 原生后台完成日常业务操作。

## 7. 权限模型

建议角色：

```text
super_admin          超级管理员
platform_admin       平台管理员
content_editor       内容运营
content_reviewer     内容审核
device_operator      设备运维
support_agent        客服支持
auditor              审计员
readonly_analyst     数据只读
```

所有删除、发布、回滚、额度和权限调整必须写审计日志。

## 8. 开发规范

- 页面、路由和 API 使用一致的模块名。
- 所有接口由 TypeScript 类型约束。
- 业务状态使用 feature store，不把所有状态塞进全局 store。
- 表格、筛选、分页、导出使用统一组件。
- 危险操作必须二次确认。
- 不把模型密钥、设备密钥和服务凭据返回浏览器。
- 功能开关关闭后，菜单、路由、按钮和 API 一起隐藏。
- 每个 feature 可以独立删除，不修改无关模块。
