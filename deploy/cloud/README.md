# 如此萌屋云端部署（1Panel）

这是一套与本地开发栈分离的云端部署包。云端只拉取 GitHub Release 发布的
镜像，不在服务器上编译 Go、Flutter 或 Vue。

## 1. 目录结构

```text
deploy/cloud/
├── docker-compose.yml
├── .env.example
├── ../public-endpoints.env   # 公网域名唯一配置源
├── mosquitto/
│   ├── mosquitto.conf
│   └── certs/                 # 由证书脚本生成，不提交
├── postgres/init/
│   └── 01-create-service-databases.sql
└── scripts/
    ├── generate-mqtt-certs.sh
    ├── export-local-data.ps1
    ├── import-cloud-data.sh
    ├── diagnose-sub2api.sh
    └── check-stack.sh
```

## 2. 迁移步骤

### 2.1 在本地导出数据

先确保本地 Compose 正在运行，然后在平台仓库根目录执行：

```powershell
.\deploy\cloud\scripts\export-local-data.ps1
```

默认输出到 `deploy/cloud/migration-data/`，包含：

- `sprout_device_platform.dump`
- `sprout_sub2api.dump`
- `sprout_sub2api_data.tar.gz`（仅模型价格、页面和插件运行资源）
- `SHA256SUMS`

这个目录包含儿童相关数据，只能通过加密通道传输，导入完成后必须删除本地和
服务器上的临时副本。

Sub2API 的 `config.yaml` 和 `.installed` 不会导出。它们保存的是数据库密码、
Redis 密码和 JWT 密钥，只属于导出的那台服务器。云端会由 `.env` 和
`AUTO_SETUP` 重新生成，避免旧密码覆盖新环境。

### 2.2 上传部署包

在 1Panel 的“文件”中创建 `/opt/1panel/apps/sprout`，上传除
`migration-data` 外的整个 `deploy/cloud` 内容。数据包单独传到
`/opt/1panel/apps/sprout/migration-data`。

也可以使用 `scp`：

```bash
scp -r deploy/cloud/* root@<server>:/opt/1panel/apps/sprout/
scp -r deploy/cloud/migration-data root@<server>:/opt/1panel/apps/sprout/
```

### 2.3 生成生产密钥

```bash
cp .env.example .env
openssl rand -hex 32   # 分别生成数据库、Redis、令牌和加密密钥
```

每个密钥生成一次后立即写入 `.env`。不要使用同一个值填充多个字段，
`SPROUT_MFA_CREDENTIAL_KEY` 必须与 `SPROUT_AI_CREDENTIAL_KEY` 不同。

`SPROUT_SUB2API_API_KEY` 不能凭空生成。它必须与 `sprout_sub2api.api_keys`
中一条 active 记录一致。导出数据库后可用以下命令查询，并把输出填入
`.env`：

```bash
docker exec -i <postgres-container> psql -U sprout -d sprout_sub2api \
  -Atc "SELECT key FROM api_keys WHERE status = 'active' ORDER BY id LIMIT 1;"
```

不要直接复用本地 `voice_gateway` 容器环境变量中的值；当前本地环境已出现
数据库记录与容器变量不一致的情况。迁移时必须以数据库查询结果为准。

### 2.4 生成 MQTT 证书

在云端执行：

```bash
chmod +x scripts/*.sh
./scripts/generate-mqtt-certs.sh mqtt.<待确认域名>
```

参数必须是设备实际连接的主机名或公网 IP。不要把 `127.0.0.1` 用于生产，
否则设备会因证书域名不匹配而拒绝连接。

当前脚本生成一张共享设备证书，只用于单批次早期迁移。正式量产前必须改为
每台设备一张唯一证书，并为设备身份建立签发、吊销和轮换流程，这是项目 P0
安全约束要求。

### 2.5 在 1Panel 启动

在“容器 → 编排 → 创建编排”中选择
`/opt/1panel/apps/sprout/docker-compose.yml`，项目名使用 `sprout`。

第一次启动只启动 PostgreSQL、Redis 和 MQTT：

```bash
docker compose -f docker-compose.yml up -d --no-deps postgres redis mqtt
```

确认 PostgreSQL 已健康后导入迁移数据：

```bash
./scripts/import-cloud-data.sh migration-data
```

导入脚本会先校验 `SHA256SUMS`，再恢复两个数据库，并把 Sub2API 的非敏感
运行数据直接写入对应命名卷；缺少命名卷时脚本会先创建它。最后启动全部业务
服务：

```bash
docker compose -f docker-compose.yml up -d
./scripts/check-stack.sh
```

如果 Sub2API 反复重启，执行诊断脚本并把输出发回：

```bash
./scripts/diagnose-sub2api.sh
```

脚本只读取状态、日志和数据卷目录，不会停止容器或修改数据。

Sub2API 的 `TOTP_ENCRYPTION_KEY` 必须是 64 位十六进制字符串，不能保留
`.env.example` 中的 `replace-with...` 文本。启动前可以先执行：

```bash
./scripts/validate-cloud-env.sh
```

校验失败时脚本会明确指出缺少或格式错误的变量，不会启动容器。

## 3. 反向代理

在 1Panel“网站”中创建以下反向代理：

| 域名 | 上游 | 用途 |
| --- | --- | --- |
| `admin.clarkhub.cn` | `http://127.0.0.1:8083` | 管理端 |
| `api.clarkhub.cn` | `http://127.0.0.1:8081` | 家长端和设备 API |
| `voice.clarkhub.cn` | `http://127.0.0.1:8082` | 语音网关，需开启 WebSocket |
| `sub.clarkhub.cn` | `http://127.0.0.1:8084` | Sub2API 管理界面和网关 API |

Sub2API 仅绑定宿主机回环地址，公网访问必须经过 1Panel 反向代理并使用
HTTPS。管理端和其他服务仍通过 Compose 内网域名调用 Sub2API，不经过公网。

Sub2API 的管理界面和 API 共用同一站点：访问域名根路径进入管理界面，
OpenAI 兼容接口继续使用 `/api/v1`。在 1Panel 中创建网站并设置反向代理后，
为 Sub2API 站点开启 WebSocket 和长连接支持，并关闭响应缓冲，避免流式输出
被整段缓存：

```nginx
proxy_buffering off;
proxy_cache off;
proxy_request_buffering off;
proxy_read_timeout 3600s;
proxy_send_timeout 3600s;
send_timeout 3600s;
```

公网域名会同时暴露 Sub2API 的登录页和网关 API。必须使用强管理员密码、
启用 MFA、限制接口密钥权限和调用额度，并定期检查登录与调用审计。若后续
不再需要公网访问，删除该域名的反向代理即可，不影响 Compose 内部调用。

管理端镜像已内置 `/api/` 同源代理，不需要为管理端再单独配置 API 路径。

设备 MQTT 使用独立域名（待确认），在 1Panel 或云防火墙中放行
`8883/tcp`。不要对公网放行 `1883/tcp`。

如果 MQTT 容器反复重启，先检查证书所有权和日志：

```bash
find mosquitto/certs -maxdepth 2 -type f -printf '%M %u:%g %p\n'
docker logs --tail=100 sprout-mqtt-1
```

证书按使用方分开存放：

- `mosquitto/certs/broker/server.key` 归 MQTT 用户所有，默认 UID/GID `1883`
- `mosquitto/certs/broker/healthcheck.key` 是本地健康检查证书，同样归 MQTT 用户
- `mosquitto/certs/device/device.key` 归平台服务用户所有，默认 UID/GID `65532`

两组私钥权限都是 `640`，证书是 `644`。不要把两组私钥放在同一个目录后
统一 `chown`，否则 broker 和平台服务之间必有一方无法读取。

如果是旧版本生成的证书目录，执行下面命令迁移到新结构并重启：

```bash
./scripts/generate-mqtt-certs.sh mqtt.<待确认域名>
docker compose --env-file .env up -d --force-recreate mqtt device_platform
```

脚本默认复用已有的 CA，只重新签发服务端和设备证书，因此不会改变设备
信任根。只有明确需要轮换整个信任链时才执行：

```bash
./scripts/generate-mqtt-certs.sh mqtt.<待确认域名> --rotate-ca
```

轮换后，已经安装旧 CA 的设备必须同步更新 `ca.crt`，否则设备会拒绝新证书。
证书目录中的私钥不应进入 Git。

## 4. 环境变量逐行说明

### PostgreSQL

| 变量 | 作用 | 生产建议 |
| --- | --- | --- |
| `SPROUT_POSTGRES_USER` | PostgreSQL 用户名 | 可保留 `sprout` |
| `SPROUT_POSTGRES_PASSWORD` | 数据库密码，被 device_platform 和 sub2api 共用 | 32 位以上随机值，不能与 Redis 相同 |
| `SPROUT_POSTGRES_DB` | 容器初始化时创建的默认库 | `sprout` |
| `SPROUT_DEVICE_PLATFORM_DATABASE_NAME` | 家长、儿童和设备业务库名 | `sprout_device_platform` |
| `SPROUT_SUB2API_DATABASE_NAME` | AI 账户、密钥和用量库名 | `sprout_sub2api` |

### Redis

| 变量 | 作用 | 生产建议 |
| --- | --- | --- |
| `SPROUT_REDIS_PASSWORD` | Redis 密码，被平台、语音网关和 sub2api 共用 | 32 位以上随机值 |

### 服务端口与监听地址

| 变量 | 作用 | 生产建议 |
| --- | --- | --- |
| `SPROUT_SERVICE_BIND_ADDRESS` | HTTP 服务的宿主监听地址 | 保持 `127.0.0.1`，由 1Panel 反代 |
| `SPROUT_MQTT_BIND_ADDRESS` | 明文 MQTT 的宿主监听地址 | 保持 `127.0.0.1`，禁止暴露公网 |
| `SPROUT_MQTTS_BIND_ADDRESS` | 加密 MQTT 的宿主监听地址 | `0.0.0.0`，公网只放行 8883 |
| `SPROUT_MQTT_UID` | 证书私钥归属的宿主 UID | 官方 Mosquitto 镜像默认 `1883` |
| `SPROUT_MQTT_GID` | 证书私钥归属的宿主 GID | 官方 Mosquitto 镜像默认 `1883` |
| `SPROUT_DEVICE_PLATFORM_UID` | 设备证书私钥归属的宿主 UID | 平台镜像默认 `65532` |
| `SPROUT_DEVICE_PLATFORM_GID` | 设备证书私钥归属的宿主 GID | 平台镜像默认 `65532` |
| `SPROUT_DEVICE_PLATFORM_PORT` | 设备平台宿主端口 | `8081` |
| `SPROUT_VOICE_GATEWAY_PORT` | 语音网关宿主端口 | `8082` |
| `SPROUT_ADMIN_WEB_PORT` | 管理端宿主端口 | `8083` |
| `SPROUT_SUB2API_PORT` | Sub2API 宿主端口，仅用于 1Panel 反向代理 | `8084` |
| `SPROUT_MQTT_PORT` | 明文 MQTT 端口 | 仅内网或其他服务使用 |
| `SPROUT_MQTTS_PORT` | 双向 TLS MQTT 端口 | 设备连接的端口，放行 8883 |

### 镜像版本

| 变量 | 作用 | 生产建议 |
| --- | --- | --- |
| `SPROUT_PLATFORM_VERSION` | 平台三个镜像的版本 | 固定为 GitHub Release 版本，不用 `latest` |
| `SPROUT_SUB2API_VERSION` | AI 网关镜像版本 | 固定为 GitHub Release 版本 |

### 服务间共享密钥

| 变量 | 作用 | 生产建议 |
| --- | --- | --- |
| `SPROUT_INTERNAL_SERVICE_TOKEN` | device_platform 调用 sub2api 内部接口的令牌 | 32 位以上随机值，绝不能下发到客户端 |
| `SPROUT_AUTH_ACCESS_TOKEN_SECRET` | 家长登录访问令牌的签名密钥 | 32 位以上随机值 |
| `SPROUT_AI_CREDENTIAL_KEY` | AI 账号凭据的加密密钥 | 32 位以上随机值 |
| `SPROUT_MFA_CREDENTIAL_KEY` | 管理端 MFA 凭据的加密密钥 | 与 AI 密钥不同，32 位以上 |
| `SPROUT_AUTH_ACCESS_TOKEN_TTL` | 访问令牌有效期 | 生产可保持 `15m` |
| `SPROUT_AUTH_REFRESH_TOKEN_TTL` | 刷新令牌有效期 | `720h` 表示 30 天 |
| `SPROUT_MFA_CHALLENGE_TTL` | MFA 挑战有效期 | 生产可保持 `5m` |

### 家长手机号验证

| 变量 | 作用 | 生产建议 |
| --- | --- | --- |
| `SPROUT_PHONE_VERIFICATION_MODE` | 手机号验证模式 | 生产必须是 `disabled`，接入真实短信后再扩展 |
| `SPROUT_ALLOW_LOCAL_SMS_BYPASS` | 是否允许本地验证码直通 | 生产必须为 `false` |

`local` 模式会接受空验证码或 `000000`，只允许在开发环境使用。

### AI 账号默认值

| 变量 | 作用 | 生产建议 |
| --- | --- | --- |
| `SPROUT_AI_DEFAULT_BALANCE_USD` | 新 AI 账号初始余额 | 按运营策略设置 |
| `SPROUT_AI_DEFAULT_CONCURRENCY` | 新 AI 账号默认并发 | 建议从 `1` 开始 |
| `SPROUT_AI_DEFAULT_MODELS` | 新账号默认允许的模型，逗号分隔 | 留空表示不额外限制 |

### Sub2API 出站域名白名单

`SECURITY_URL_ALLOWLIST_UPSTREAM_HOSTS` 是 Sub2API 允许连接的上游域名列表。
该变量是完整覆盖而非追加，多个域名使用英文逗号分隔，不能包含协议、端口或
路径。新增中转站时必须保留当前正在使用的全部域名，否则其他账号会立即失效。

```env
SECURITY_URL_ALLOWLIST_UPSTREAM_HOSTS=sub.unsee.you,api.openai.com,api.anthropic.com,api.kimi.com,api.moonshot.ai,api.moonshot.cn,open.bigmodel.cn,api.minimaxi.com,api.minimax.io,opencode.ai,generativelanguage.googleapis.com,cloudcode-pa.googleapis.com,*.openai.azure.com
```

修改后必须重建 Sub2API 容器，环境变量不会在运行中的容器内自动刷新。

### Sub2API

| 变量 | 作用 | 生产建议 |
| --- | --- | --- |
| `SPROUT_SUB2API_API_KEY` | voice_gateway 调用 AI 网关的密钥 | 必须与数据库中的 active key 一致 |
| `SPROUT_SUB2API_ADMIN_EMAIL` | Sub2API 初始化管理员邮箱 | 使用真实可用邮箱 |
| `SPROUT_SUB2API_ADMIN_PASSWORD` | Sub2API 管理员密码 | 强随机密码 |
| `SPROUT_SUB2API_JWT_SECRET` | Sub2API 会话签名密钥 | 32 位以上随机值 |
| `SPROUT_SUB2API_TOTP_ENCRYPTION_KEY` | Sub2API 的 TOTP 加密密钥 | 必须为 64 位十六进制字符串，且必须固定 |
| `SPROUT_TIMEZONE` | 容器时区 | `Asia/Shanghai` |

## 5. 升级方式

1. 在 GitHub 确认新 Release 已发布，并记录版本号。
2. 修改 `.env` 中的 `SPROUT_PLATFORM_VERSION` 和 `SPROUT_SUB2API_VERSION`。
3. 在 1Panel 重新拉取镜像并重建：

```bash
docker compose -f docker-compose.yml pull
docker compose -f docker-compose.yml up -d --remove-orphans
./scripts/check-stack.sh
```

PostgreSQL、Redis、MQTT 和 Sub2API 数据都在命名卷中，重建容器不会删除。
不要执行 `docker compose down -v`。

## 6. 安全清单

- `.env` 权限设为 `600`，不提交 Git。
- Sub2API 只绑定 `127.0.0.1`，公网入口由 1Panel 终止 HTTPS 后转发；禁止
  直接映射到 `0.0.0.0`。数据库端口始终不暴露公网。
- 生产环境禁用手机号验证直通。
- 生产环境替换 MQTT 证书，并启用域名匹配校验。
- 迁移完成后删除服务器上的 `migration-data` 和本地导出包。
- 在 1Panel 中启用 HTTPS，并确认管理端、API 和语音域名均使用证书。
