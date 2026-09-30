# 后台管理端代码文档

## 1. 定位

`apps/admin_web` 是 Clarkaitoy 内部运营和管理人员使用的 Vue 3 控制台，负责账号
权限、家庭儿童设备运营、内容、AI 网关、OTA、审计、监控和系统配置。

管理端不直接连接设备，不保存模型供应商密钥，不把设备长期密钥返回浏览器。

## 2. 目标目录

```text
apps/admin_web/
  src/
    app/
      adminApp.ts
      router.ts
      routes/
      featureRegistry.ts
    api/
      httpClient.ts
      contracts/
      errorMapper.ts
    components/
      data-table/
      filters/
      forms/
      charts/
    layouts/
    features/
      auth/
      rbac/
      dashboard/
      family/
      child/
      device/
      content/
      ai_gateway/
      ota/
      parent_policy/
      audit/
      monitoring/
      system/
      notification/
    stores/
    types/
    utils/
    i18n/
  tests/
    unit/
    component/
    e2e/
```

每个 feature 统一使用：

```text
features/<feature_name>/
  presentation/
    pages/
    components/
  application/
    use-cases/
    stores/
  domain/
    models/
    errors/
  data/
    api/
    mappers/
  routes.ts
  permissions.ts
  README.md
```

## 3. 技术选型

完整版本和理由见
[第三方库选型](../technology-selection.md#3-vue-3--vue-router)。

| 场景 | 默认选择 | 说明 |
| --- | --- | --- |
| 框架 | Vue 3 + TypeScript | Composition API |
| 路由 | `vue-router` | 动态路由、守卫、懒加载 |
| 状态 | `pinia` | 认证、权限、主题和全局配置 |
| 服务端状态 | `@tanstack/vue-query` | 查询缓存、失效、分页 |
| HTTP | `axios` | 拦截器、错误映射、上传下载 |
| UI | `element-plus` | 中后台默认组件库 |
| 工具 | `@vueuse/core` | 浏览器 API 和组合式工具 |
| 表单 | `vee-validate` + `zod` | 表单状态、校验和契约校验 |
| 图表 | `echarts` | 监控和运营图表 |
| 日期 | `dayjs` | 时间显示和筛选 |

## 4. 功能模块

### 4.1 P0

| 模块 | 页面/能力 | 依赖 |
| --- | --- | --- |
| `auth` | 管理员登录、MFA、会话、退出 | `device_platform/auth` |
| `rbac` | 角色、菜单、数据权限、操作权限 | `device_platform/auth` |
| `dashboard` | 设备、在线率、AI 调用、错误、OTA 概览 | `device_platform/telemetry` |
| `family` | 家庭、成员、设备关联和状态 | `device_platform/family` |
| `child` | 儿童档案、内容等级、策略和数据删除 | `device_platform/child` |
| `device` | 注册、型号、固件、在线、配置、解绑 | `device_platform/device` |
| `content` | 音频、文本、图片、主题包、分龄、审核、发布 | `device_platform/content` |
| `ai_gateway` | 模型、账号池、路由、限流、配额、调用日志 | `voice_gateway` + `sub2api_fork` |
| `ota` | 固件包、灰度、设备组、发布、回滚、失败统计 | `device_platform/ota` |
| `audit` | 管理操作、登录、敏感数据访问 | `device_platform/audit` |

### 4.2 P1

| 模块 | 页面/能力 | 依赖 |
| --- | --- | --- |
| `parent_policy` | 查看和协助配置家长策略、异常审核 | `device_platform/parent_policy` |
| `monitoring` | 设备、服务、模型、网络和告警规则 | `device_platform/telemetry` |
| `notification` | 推送、短信、邮件、站内消息和模板 | `device_platform/notification` |
| `companionship` | 成长记录、长期记忆授权和内容推荐 | `device_platform/companion` |
| `english_learning` | 学习内容、课程、进度和反馈统计 | `device_platform/learning` |
| `vision` | 摄像头能力、图片隐私和视觉模型配置 | `device_platform/vision` |
| `diagnostics` | 设备日志、崩溃、网络质量和温度趋势 | `device_platform/telemetry` |

### 4.3 P2

| 模块 | 页面/能力 | 依赖 |
| --- | --- | --- |
| `display` | 屏幕、亮度、护眼、动画和视频配置 | `device_platform/display` |
| `touch` | 触摸、手势和触觉反馈配置 | `device_platform/input` |
| `cellular` | eSIM、流量、漫游、远程唤醒和套餐 | `device_platform/connectivity` |
| `battery` | 电池、充电、低功耗和休眠策略 | `device_platform/power` |
| `motion` | 动作、舵机、电机和动作编排 | `device_platform/motion` |
| `form_factor` | 机器狗、机器猫、毛绒等形态能力 | `device_platform/device` |
| `video_companion` | 视频通话、远程陪伴和会话统计 | `voice_gateway/video` |
| `wearable` | 定位、电话、电子围栏、SOS | `device_platform/wearable` |
| `multi_device` | 家庭联动、设备组网和内容同步 | `device_platform/device_mesh` |

## 5. 权限模型

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

权限同时作用于菜单、路由、按钮、接口和数据范围。删除、发布、回滚、额度和权限
调整必须写审计日志，并进行二次确认。

## 6. 模块注册

`featureRegistry.ts` 负责集中注册 feature：

```ts
export const featureRegistry: AdminFeature[] = [
  authFeature,
  dashboardFeature,
  deviceFeature,
  contentFeature,
  aiGatewayFeature,
  otaFeature,
  auditFeature,
  systemFeature,
]
```

每个 feature 提供路由、权限、导航项和依赖声明。禁用 feature 时，菜单、路由、
按钮和 API 一起隐藏，不允许页面保留不可用入口。

## 7. 与 sub2api 的边界

- `sub2api` 负责 AI 账号、渠道、模型、配额、计费和调用统计。
- Clarkaitoy 管理台负责儿童产品、设备、内容和运营业务。
- AI 网关视图通过 `device_platform/ai_gateway` 适配层调用 `sub2api`。
- 不要求运营人员直接登录 `sub2api` 原生后台。
- 供应商密钥只在服务端保存，浏览器只接收脱敏状态和统计数据。

## 8. 数据表格和表单

- 列表统一支持分页、排序、筛选、列显隐、批量操作和导出。
- 服务器分页优先，不在浏览器加载全量数据。
- 删除、发布、回滚、配额调整必须二次确认和审计。
- 所有接口由 TypeScript 类型约束。
- Zod 负责运行时契约校验，不能让非法响应进入领域模型。
- API 错误统一映射为页面可展示的错误状态。

## 9. 测试

| 层级 | 内容 |
| --- | --- |
| 单元测试 | 权限、状态、校验、数据映射 |
| 组件测试 | 表格、筛选、表单、危险操作确认 |
| E2E | 登录、RBAC、设备、内容、OTA、审计 |
| 契约测试 | OpenAPI 响应和前端类型一致 |
| 视觉测试 | 关键运营页面和移动宽度 |

## 10. 删除方法

1. 删除 `features/<feature_name>/`。
2. 从 `featureRegistry.ts` 移除 feature。
3. 移除路由、权限项、导航项和 API 客户端。
4. 移除 feature 专属依赖。
5. 运行类型检查、构建、单元测试和 E2E。
