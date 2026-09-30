# GitHub 仓库与上传指南

## 1. 推荐仓库划分

Clarkaitoy 使用混合式仓库结构：

| 仓库 | 本地路径 | 内容 |
| --- | --- | --- |
| `clarkaitoy-platform` | `D:\service\clarkaitoy` | 家长端、管理端、两个 Go 服务、共享契约、文档和部署配置 |
| `clarkaitoy-firmware` | `D:\service\clarkaitoy\firmware` | ESP32-S3 N16R8 PlatformIO 工程 |
| `clarkaitoy-sub2api-fork` | `D:\service\clarkaitoy\services\sub2api_fork` | `sub2api` fork 和 Clarkaitoy AI 网关适配 |

平台仓库不跟踪另外两个仓库的内容，只通过 `workspace.lock.yaml` 记录它们的产品版本。

## 2. 首次创建 GitHub 仓库

在 GitHub 网页上创建三个空仓库，不要勾选初始化 README、`.gitignore` 或 License：

```text
clarkaitoy-platform
clarkaitoy-firmware
clarkaitoy-sub2api-fork
```

如果账号属于组织，把 `TissyBoxC` 替换为组织名。三个仓库都建议设为 Private。

## 3. 上传平台仓库

在 `D:\service\clarkaitoy` 执行：

```powershell
git status --short --branch
git remote add origin https://github.com/TissyBoxC/clarkaitoy-platform.git
git push -u origin master
```

如果希望默认分支名与 GitHub 常见的 `main` 一致，先改分支再推送：

```powershell
git branch -M main
git push -u origin main
```

根仓库当前没有远端；上面的 `remote add` 只执行一次。

## 4. 上传固件仓库

固件仓库已经存在 `origin`，但当前有一个未推送提交：

```powershell
Set-Location D:\service\clarkaitoy\firmware
git status --short --branch
git push origin main
```

如果希望把它推送到新的 `clarkaitoy-firmware` 仓库：

```powershell
git remote set-url origin https://github.com/TissyBoxC/clarkaitoy-firmware.git
git push -u origin main
```

该仓库当前远端是 `https://github.com/TissyBoxC/Clarkaitoy.git`。在覆盖前确认它确实是新的固件仓库地址，避免把固件提交推到错误项目。

## 5. 上传 sub2api fork

`services/sub2api_fork` 保留了两个远端：

| 远端 | 用途 |
| --- | --- |
| `origin` | Clarkaitoy fork，允许推送 |
| `upstream` | 原项目，只用于同步，禁止推送 |

先确认远端：

```powershell
Set-Location D:\service\clarkaitoy\services\sub2api_fork
git remote -v
```

推送当前 fork：

```powershell
git push origin main
```

如果 fork 名称或所有者发生变化：

```powershell
git remote set-url origin https://github.com/TissyBoxC/clarkaitoy-sub2api-fork.git
git push -u origin main
```

开始魔改前创建保护分支：

```powershell
git switch -c clarkaitoy/voice-gateway-adapter
```

同步上游时始终先 fetch，再 rebase：

```powershell
git fetch upstream
git switch main
git rebase upstream/main
git push origin main
```

不要向 `upstream` 执行 push，也不要把儿童、家庭、设备、OTA 或业务内容推进 `sub2api` 上游核心。

## 6. 认证方式

HTTPS 推送需要 GitHub Personal Access Token 或 GitHub Desktop 凭据。推荐安装 GitHub CLI 后执行：

```powershell
gh auth login
gh auth status
```

使用 SSH 时可以改为：

```powershell
git remote set-url origin git@github.com:TissyBoxC/clarkaitoy-platform.git
ssh -T git@github.com
```

## 7. 推送前检查

每个仓库分别执行：

```powershell
git status --short --branch
git diff --check
git log --oneline -5
```

确认没有提交以下内容：

```text
AGENTS.md
node_modules/
bgen/
dist/
.env
设备证书和私钥
本地数据库或日志
```

当前 `AGENTS.md` 只作为本机开发约束文件，已经被三个仓库的忽略规则排除，不作为产品源码上传。

## 8. 分支和提交规范

提交格式保持：

```text
<type>(<scope>): <summary>
```

推荐类型：

```text
feat
fix
refactor
perf
test
docs
build
chore
```

示例：

```text
feat(voice_gateway): add websocket audio session
fix(device_platform): validate device activation code
docs(repo): document external repository revisions
```
