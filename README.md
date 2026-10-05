# 新蜂资产管理平台 — CMDB RPC 服务

承载资产模型与配置项（CI）的数据管理、校验、权限检查、变更记录及生命周期相关逻辑，为 CMDB API 和其他平台模块提供 RPC 接口。

仓库：[coder-lulu/newbee-cmdb-rpc](https://github.com/coder-lulu/newbee-cmdb-rpc) · [平台工作区](https://github.com/coder-lulu/newbee)

## 获取代码

推荐通过完整工作区开发，保留兄弟模块目录及本地 `replace` 依赖。以下命令使用 Bash；Go 工作区要求 Go 1.25.1 或更高版本。

```bash
git clone --recurse-submodules https://github.com/coder-lulu/newbee.git
cd newbee/cmdb/rpc
```

已有工作区执行 `git submodule update --init --recursive`。单独克隆模块时，需要自行补齐 `go.mod` 中的本地依赖路径。

## 目录导航

| 路径 | 用途 |
| --- | --- |
| `cmdb.go` | 服务入口 |
| `ent/` | 数据模型及生成代码 |
| `internal/logic/`、`internal/svc/` | 业务逻辑与依赖装配 |
| `internal/server/`、`types/` | RPC 实现和协议类型 |
| `etc/cmdb.yaml.example` | 公开配置模板 |

## 配置与本地运行

首次配置时执行下列复制命令；已有配置不要覆盖：

```bash
cp etc/cmdb.yaml.example etc/cmdb.yaml
```

按环境修改 `DatabaseConf、RedisConf、CoreRpc、CasbinConf 和 InputAdapterConf` 等配置项，以及监听地址。示例 RPC 端口为 `9200`，以实际配置为准。入口使用 `conf.UseEnv()`，可为示例中的 `${...}` 占位符设置对应环境变量，也可在本地配置中填写值。模板中的内网地址不是可直接使用的公共服务。

先准备数据库、Redis 和配置引用的 RPC 服务，再启动本服务；API 应在对应 RPC 就绪后启动。真实配置和凭据不要提交。

```bash
go run . -f etc/cmdb.yaml
```

## 首次数据库初始化

先初始化 Core，并配置 `CoreRpc`（支持 `Endpoints`、`Target` 或 Etcd 服务发现）。在平台根目录运行 `bash init-databases.sh -s cmdb`。初始化按 Ent 模型创建表并保留已有数据，补齐基础模型数据后，通过 Core RPC 为默认租户登记当前 API 目录和 CMDB 菜单。已有菜单按路径或组件复用，新增菜单使用数据库分配 ID；仅合并默认 `superadmin` 的菜单授权，保留原授权，普通角色需单独授权。

Core 不可用、默认角色缺失或目录登记失败会返回错误；修复后可重复初始化。请在启动 API 前完成初始化。无需导入本机数据库备份或手动执行历史固定菜单 ID 脚本。

## 构建与验证

在当前模块目录执行：

```bash
go build .
go test ./...
go vet ./...
```

测试中的集成用例需要对应基础服务。修改协议或 Ent Schema 后，应使用当前 Makefile 中对应生成目标并审查生成差异；不要直接编辑生成文件。上述命令是验证入口，不表示所有测试已通过。

## 文档

[CI 数据架构](doc/ci_data_architecture.md) · [输入适配器指南](doc/input_adapter_usage_guide.md) · [权限对接](doc/CI权限管理前端对接文档.md)

`internal/core/` 聚合 CI 数据操作、校验、权限、变更记录和生命周期处理；历史架构说明见 [ROOT_README.md](ROOT_README.md)，其中阶段性完成标记需结合当前实现核对。

## 许可证与来源

本仓库采用 [Apache-2.0](LICENSE)。沿用现有服务框架的上游许可，保留文件中的原作者版权。第三方依赖遵循各自许可证，保留原有版权与许可声明。
