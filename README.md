# devx

Go 项目的开发脚手架生成器.通过 `devgen` 从统一配置生成 `Makefile`,target 构建脚本,Dockerfile,CI 与 lint 配置,减少各仓库重复维护成本.

## 安装

在模块根目录安装 CLI(推荐与项目 `go` 版本一致):

```bash
go install github.com/xoctopus/devx/cmd/devgen@latest
```

或在仓库内直接运行:

```bash
go run ./cmd/devgen <command>
```

## 快速开始

```bash
# 1. 在 Go 模块根目录初始化配置(生成 .devx/config.json)
devgen init

# 2. 按需编辑 .devx/config.json(target,镜像,依赖工具等)

# 3. 生成全部产物
devgen all

# 4. 使用生成的 Makefile
make test
make target_<name>    # 构建单个 cmd 产物到 dist/<name>/
make image_<name>     # 构建 Docker 镜像(需 target 配置了 image)
```

`devgen all` 会依次生成:

| 产物                         | 说明                             |
|------------------------------|----------------------------------|
| `Makefile`                   | 项目根构建,测试,lint,target 聚合 |
| `cmd/<name>/Makefile`        | 单个 target 的 build / install   |
| `cmd/<name>/Dockerfile`      | 可选,配置了 `image` 时生成       |
| `.golangci.yml`              | golangci-lint 配置               |
| `.github/workflows/ci.yml`   | GitHub Actions CI                |
| `.github/workflows/lint.yml` | GitHub Actions Lint              |
| `.github/dependabot.yml`     | Dependabot                       |

## 配置

配置文件路径固定为 **`.devx/config.json`**.顶层结构:

```json
{
  "ci": { "CI": true },
  "make": { ... },
  "lint": { "lint": true }
}
```

### `make`:Makefile 与构建 target

| 字段                 | 类型       | 说明                                           |
|----------------------|------------|------------------------------------------------|
| `test_ignore`        | `[]string` | 覆盖率统计排除模式                             |
| `format_ignore`      | `[]string` | `goimports-reviser` 排除模式                   |
| `env`                | `[]string` | 全局环境变量,**键值对**(如 `["GOWORK","off"]`) |
| `hack_test`          | `bool`     | 是否启用 hack 测试依赖(docker compose)         |
| `enable_bench_cover` | `bool`     | `cover` 时是否带 `-bench`                      |
| `depends`            | `[]Tool`   | 额外开发依赖工具(与默认工具合并)               |
| `target`             | `[]Target` | `cmd/<name>` 构建目标                          |

未配置时,`init` 会写入与 devx 自身相近的默认值(test/format ignore,`GOWORK=off`,golangci-lint 等).

#### `depends` / `Tool`

```json
{
  "name": "golangci-lint",
  "repo": "github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
  "version": "latest",
  "type": "LINTER",
  "description": "code static checking"
}
```

`type` 可选:`LINTER`,`FORMATER`,`DOC`,`DEVGEN`.

#### `target`:cmd 入口与镜像

`name` 必须与 `cmd/<name>` 目录名一致.

```json
{
  "name": "demo",
  "image": {
    "runtime": "alpine:latest",
    "timezone": "Asia/Shanghai",
    "cgo_enabled": false,
    "go_proxy": "https://goproxy.cn,direct",
    "expose": "80",
    "run_args": ["run"]
  }
}
```

| `image` 字段  | 说明                                                                  |
|---------------|-----------------------------------------------------------------------|
| `runtime`     | 运行阶段基础镜像,默认 `alpine:latest`                                 |
| `timezone`    | 容器时区,设置后安装 `tzdata`                                          |
| `cgo_enabled` | 构建阶段 `CGO_ENABLED`                                                |
| `go_proxy`    | 构建阶段 `GOPROXY`                                                    |
| `expose`      | `EXPOSE` 端口                                                         |
| `run_args`    | `ENTRYPOINT` 附加参数,如 `["run"]` → `ENTRYPOINT ["/app/app", "run"]` |

未配置 `image` 时,只生成 `cmd/<name>/Makefile`,不生成 Dockerfile.

### `lint`

```json
{ "lint": true }
```

为 `false` 时不生成 `.golangci.yml`.

### `ci`

```json
{ "CI": true }
```

为 `false` 时不生成 GitHub workflows 与 dependabot.

## 命令

| 命令          | 说明                                   |
|---------------|----------------------------------------|
| `devgen init` | 创建 `.devx/config.json`(已存在则跳过) |
| `devgen show` | 打印当前配置;无文件时等同 `init`       |
| `devgen all`  | `init` + 按配置生成全部产物(**常用**)  |
| `devgen make` | 仅生成 Makefile 与 target 相关文件     |
| `devgen lint` | 仅生成 `.golangci.yml`                 |
| `devgen ci`   | 仅生成 `.github` 下 CI 配置            |

`all` 支持 `--file` 指定配置文件,默认 `.devx/config.json`.

子命令 `make` / `lint` / `ci` 也可通过命令行 flags 覆盖配置字段(见 `devgen make -h`).

## 生成后的 Makefile 用法

### 根目录 `Makefile`

| 目标                        | 说明                                     |
|-----------------------------|------------------------------------------|
| `make test`                 | 单元测试                                 |
| `make cover` / `view-cover` | 覆盖率                                   |
| `make fmt` / `fmt-check`    | 格式化                                   |
| `make lint`                 | golangci-lint + go vet                   |
| `make tidy`                 | go mod tidy(有 vendor 则 vendor)         |
| `make dep`                  | 安装缺失的开发工具                       |
| `make target_<name>`        | 构建并安装到 `dist/<name>/`              |
| `make targets`              | 构建全部 target                          |
| `make image_<name>`         | `docker build`(在 `cmd/<name>` 目录)     |
| `make pre-commit`           | fmt + lint + cover + changelog + targets |

### `cmd/<name>/Makefile`

| 目标           | 说明                                                         |
|----------------|--------------------------------------------------------------|
| `make build`   | 编译当前 target                                              |
| `make install` | 安装到 `../../dist/<name>/`(二进制,`version`,可选 `config/`) |
| `make image`   | 在 `cmd/<name>` 下构建镜像(需生成 Dockerfile)                |

可在 `cmd/<name>` 内直接 `make install`;Makefile 内嵌 git 元数据逻辑,无需依赖根 Makefile 导出变量.

`dist/<name>/` 典型内容:

```
dist/demo/
  demo                 # 可执行文件
  demo-linux-amd64     # 带平台后缀的副本
  version              # 版本文本
  config/              # 若 cmd/<name>/config 存在
```

## 版本信息注入

target 构建通过 **ldflags** 写入 `main` 包变量(需在 `cmd/<name>/main.go` 中声明同名变量):

- `Name`,`Feature`(分支),`Version`(tag),`CommitID`,`CommitAt`,`BuildAt`

构建阶段设置 `GOFLAGS=-buildvcs=false`,版本由 Makefile 的 git 查询 + ldflags 负责,不依赖 Go 内置 VCS stamping.

在容器 / Kaniko 中构建时:

- Dockerfile 会配置 `git safe.directory`,便于 COPY 进来的 `.git` 可被 Makefile 读取
- 若构建上下文不含 `.git`,需在 CI 中通过环境变量等方式注入版本,或保证上下文包含完整仓库

## Docker 镜像构建

生成的 `cmd/<name>/Dockerfile` 采用 **builder + runtime 两阶段**:

1. **builder**:`make target_<name>`,产物打包到 `/out/`
2. **runtime**:`COPY --from=builder`,二进制安装为 `/app/app`

设计为兼容 **Kaniko** 等无 Docker daemon 环境:使用经典 `COPY --from` 与普通 `RUN`,避免 heredoc,`RUN --mount` 等 BuildKit 高级写法.

构建前会 `test -x` 校验二进制,避免"镜像构建成功但缺少 `/app/app`"导致 Pod 启动失败.

本地构建示例:

```bash
make image_demo
# 或
docker build -f cmd/demo/Dockerfile .
```

## 项目结构示例

```
.
├── .devx/
│   └── config.json
├── Makefile
├── cmd/
│   ├── demo/
│   │   ├── main.go
│   │   ├── Makefile      # devgen 生成
│   │   ├── Dockerfile    # 配置了 image 时生成
│   │   └── config/       # 可选,会打进 dist 与镜像
│   └── devgen/
│       └── Makefile
├── dist/
│   └── demo/
└── .github/workflows/
```

## 开发与发布

```bash
make test
make pre-commit
make target_devgen
```

`devgen` 自身的配置见 `.devx/config.json`.

## 相关仓库

- CLI:`github.com/xoctopus/devx/cmd/devgen`
- 库:`github.com/xoctopus/devx/internal/devx`
