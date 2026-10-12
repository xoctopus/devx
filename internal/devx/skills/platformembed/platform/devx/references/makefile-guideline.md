# Makefile 与 devx 模板对照

## 来源

| 层级     | 文件                    | 生成方式                                         |
|----------|-------------------------|--------------------------------------------------|
| 仓库根   | `Makefile`              | `devgen make` → `project.tpl.mk` + `meta.tpl.mk` |
| 命令入口 | `cmd/<name>/Makefile`   | 每个 `make.target[].name` → `target.tpl.mk`      |
| 镜像     | `cmd/<name>/Dockerfile` | `target[].image` 非空 → `target.tpl.dockerfile`  |

`meta.tpl.mk` 统一提供: `MODULE_PATH`, `GIT_COMMIT`, `GIT_TAG`, `GIT_BRANCH`, `BUILD_AT` 等.

## multi-app-go 根 Makefile 模式

下列片段对应 **`.devx/config.json` 中 `make.hack_test: true`** 时 devx 生成的行为.
devx 默认 `hack_test: false`, 无 `hack_dep_*`, 也不 export `HACK_TEST`; multi-app-go 等业务仓需在 config 中显式开启.

```make
HACK_TEST ?= true
export HACK_TEST

test: dep tidy hack_dep_run
	go test ./... -race -failfast -parallel 1 -gcflags="all=-N -l"

hack_dep_run:
	@cd hack && docker compose up -d --remove-orphans

hack_dep_stop:
	@cd hack && docker compose down -v

target_srv-mm28:
	@make -C cmd/srv-mm28 --no-print-directory install

image_srv-mm28:
	@make -C cmd/srv-mm28 --no-print-directory image

targets: target_srv-mm28 target_srv-robot
images: image_srv-mm28 image_srv-robot
```

命名规则: `target_<cmd目录名>`, `image_<cmd目录名>` (与 `cmd/` 下文件夹名一致).

## devx `Makefile` JSON 字段

| 字段                 | 默认            | 作用                                                     |
|----------------------|-----------------|----------------------------------------------------------|
| `hack_test`          | false           | true → export `HACK_TEST`, `hack_dep_*`, test/cover 前置 |
| `target[]`           | []              | 每个 `name` 对应 `cmd/<name>`                            |
| `target[].image`     | 省略            | 省略则只生成 `target_*`, 无 Dockerfile/`image_*`         |
| `test_ignore`        | 内置模式        | 覆盖率 grep 排除                                         |
| `format_ignore`      | 内置模式        | goimports-reviser 排除                                   |
| `env`                | `GOWORK=off` 等 | 写入 Makefile `export`                                   |
| `depends`            | git-chglog 等   | `dep` 目标安装 CLI 工具 (与内置默认合并)                 |
| `enable_bench_cover` | false           | cover 时加 `-bench=.`                                    |

`ImageOption` 常用字段: `runtime`, `timezone`, `cgo_enabled`, `go_proxy`, `expose`, `run_args`.

完整 JSON 与 `devgen` 命令: [devx README](https://github.com/xoctopus/devx/blob/main/README.md).

## cmd/<name>/Makefile 行为

### install 产物布局

```
dist/<name>/
  <name>              # 可执行文件
  <name>-<os>-<arch>  # 带平台后缀副本
  version             # 两行: meta 字符串 + module path
  config/             # 若 cmd 目录存在 config/ 则复制
```

### image 变量

- `IMAGE_NAME`: 默认等于 cmd 目录名 (`NAME`)
- `IMAGE_TAG`: 模板为 `$(if $(GIT_TAG),$(GIT_TAG),latest)-$(GIT_COMMIT)` (无 tag 时前缀为 `latest`)

Docker build 必须在 **cmd 目录** 执行 (Makefile 已写 `-f Dockerfile` 且 context 为 `../..`).

## hack 与 HACK_TEST

仅当 **`hack_test: true`** 且已 `devgen make` 后:

```
make test
  → dep (工具)
  → tidy
  → hack_dep_run (compose up)
  → go test ./...
```

`hack_test: false` 时 `make test` 为 `dep tidy` 后直接 `go test`, 无 compose 目标.

测试代码约定:

```go
func TestIntegration(t *testing.T) {
hack.Check(t) // HACK_TEST != "true" 时 Skip
ctx := hack.WithMysql(t, ...) // 或项目自有 helper
// ...
}
```

CI 若已有 service container: 用
**`HACK_TEST=false go test ./...`** 或直连 CI DSN, 并避免在 CI 里依赖 `hack_dep_run` (可设 `hack_test: false` 生成 CI 专用 Makefile, 或 CI 只跑 `go test`).

## 单测 flags 说明

模板默认:

- `-race`: 数据竞争检测
- `-failfast`: 首个失败后停止
- `-parallel 1`: 与共享 DB/全局状态集成测兼容
- `-gcflags="all=-N -l"`: 便于调试/与某些 mock 兼容

纯单元测可不经过 Makefile, 自行 `go test` 调整并行度.

## 排查

| 现象                        | 检查                                                                      |
|-----------------------------|---------------------------------------------------------------------------|
| 测试 Skip `HACK_TEST=false` | 集成测需 `export HACK_TEST=true` 或 `make test`; `make show` 看 env       |
| `make test` 仍起 compose    | 预期行为 (`hack_test: true`); 单包用 `HACK_TEST=false go test ./pkg/...`  |
| compose 端口冲突            | `hack/docker-compose.yml` host 端口是否占用                               |
| `target_*` 不存在           | `.devx/config.json` 是否声明 target; 是否重新 `devgen make`               |
| 镜像找不到二进制            | builder 阶段 `make target_<name>` 是否成功; `dist/<name>/<name>` 是否存在 |
| vendor 模式                 | 根与 cmd Makefile 均可能加 `-mod=vendor`                                  |
