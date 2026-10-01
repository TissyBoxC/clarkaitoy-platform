<div align="center">

<img src="../../assets/brand/sprout/brand_banner.png" alt="如此萌屋" width="560" />

</div>

# 如此萌屋后台管理端

面向平台运营和管理人员的 Vue 3 管理端，负责芽系列·初芽平台的日常管理和安全审计。

## 职责

- 账号、角色、MFA 和最小权限管理。
- 家庭、儿童和设备状态管理。
- 内容、界面文案、AI 网关和 OTA 发布管理。
- 审计日志、异常告警和运营数据查看。

浏览器只访问平台管理服务，不保存供应商密钥，也不直接暴露儿童敏感数据。

## 技术栈

| 场景 | 选型 |
| --- | --- |
| 框架 | Vue 3、TypeScript、Vite |
| 路由 | Vue Router |
| 状态管理 | Pinia |
| 网络 | Axios |
| 测试 | Playwright、Vue Test Utils 按功能补充 |
| 格式化 | Prettier |

管理端 UI 组件库将在中后台功能开发时优先从 Element Plus、Naive UI 或
Ant Design Vue 中按现有设计约束选定，避免同时引入多套组件体系。

## 本地运行

```powershell
npm install
npm run dev
```

## 目录

```text
src/api/          HTTP、错误映射和会话刷新边界
src/app/          应用装配、路由和全局状态
src/components/   无业务通用组件
src/features/     按业务能力拆分的功能模块
src/layouts/      管理端页面骨架和品牌区域
public/brand/     浏览器可直接访问的品牌资源
```

## 质量门禁

```powershell
npm run type-check
npm run build
npm run test:e2e
```
