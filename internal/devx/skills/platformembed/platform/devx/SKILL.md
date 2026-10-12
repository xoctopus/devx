---
name: devx
description: >-
  归纳 devx 生成的 Makefile 与 hack 依赖容器约定: 带 MySQL/TiDB/Redis 等的集成测试,
  本地/CI 单测, cmd 二进制构建与 Docker 镜像打包.
  当任务涉及 devgen, make test/cover, hack/docker compose, target_*/image_*, 或 .devx/config.json 时使用.
triggers:
  - "devgen"
  - ".devx/config.json"
  - "make test"
  - "hack/docker-compose"
  - "target_*"
  - "image_*"
  - "HACK_TEST"
---

# devx (Makefile / hack / 镜像)

Go 业务仓库通常由 **[devx](https://github.com/xoctopus/devx) `devgen`** 生成根 `Makefile` 与各 `cmd/<name>/Makefile`.
本 skill 描述 **生成之后** 的用法; CLI 与 JSON 字段以 [devx README](https://github.com/xoctopus/devx/blob/main/README.md) 为准.

参考实现: `github.com/xoctopus/devx/internal/devx/static/*.tpl.mk`, 示例仓库根 `Makefile` (如 multi-app-go).

详细模板对照见 [makefile-guideline.md](references/makefile-guideline.md).

## 0. devgen 入口 (生成物从哪来)

在 **Go 模块根** (含 `go.mod`):

```bash
go install github.com/xoctopus/devx/cmd/devgen@latest
# 或 go get -tool github.com/xoctopus/devx/cmd/devgen@vX.Y.Z && go tool devgen ...

devgen init          # 创建 .devx/config.json
# 编辑 config 后
devgen all           # 常用: Makefile + lint + CI
# 或 devgen make     # 仅 Makefile / cmd Makefile / Dockerfile
```

- 勿手改 devgen 生成的 `Makefile` / `Dockerfile`; 改 `.devx/config.json` 后重新 `devgen make` 或 `devgen all`.
- `devgen lint` / `devgen ci` 生成 `.golangci.yml` 与 `.github/*` (见 devx README).

### Agent skills (业务仓)

1. `go.mod` 标注 `// +skill:<name>`, 运行 **`go tool skill-install`** (genx) → 填充 **`.agents/skills/`** (devgen 不代劳).
2. 在 **go.mod 同级** 运行 **`devgen skill install`**:
   - **项目级**: 项目 skill 只存在于 **`.agents`**; 在模块根 (或 **`--path`** 指定目录) 创建 **`.cursor` → `.agents`**、**`.claude` → `.agents`** (相对 symlink, 非逐 skill 链接).
   - **用户级**: 将 **devgen 内嵌** 的 platform skills (如 `devx`, `base`) 写入 **`~/.cursor/skills/<name>`** 等 (Codex: **`~/.agents/skills/<name>`**). 内容与 devgen 版本一致; 要新 SKILL 先升级 devgen.
3. 可选 **`devgen skill install --path <dir>`**: 在 `<dir>` 安装桥接, **`.agents` 仍为执行目录下**; **`--name cursor|claude|codex`** 只处理指定 Agent.

业务仓 **不必** `require` devx 模块来拿 devx skill; 使用 **`devgen skill install`** 即可装 platform skill.

## 1. 带依赖的单测 / 集成测

### Makefile 侧

在 `.devx/config.json` 的 `make` 段开启 **`hack_test: true`** 时, devx 会:

- 在 Makefile 中 `export HACK_TEST=true`
- 为 `test` / `cover` 增加前置: `dep tidy hack_dep_run`
- 生成 `hack_dep_run` / `hack_dep_stop`

默认 **`hack_test: false`**: 无 `hack_dep_*`, Makefile 通常也不 export `HACK_TEST`.

```make
hack_dep_run:
	@cd hack && docker compose up -d --remove-orphans

hack_dep_stop:
	@cd hack && docker compose down -v
```

根目标示例 (`hack_test: true` 时):

```make
test: dep tidy hack_dep_run
	go test ./... -race -failfast -parallel 1 -gcflags="all=-N -l"
```

**`HACK_TEST` 与 `make test`**

| 场景 | 做法 |
|------|------|
| 全量集成测 (要 compose) | `make test` 或 `export HACK_TEST=true` 后 `go test ./...` |
| 只跑不依赖容器的包 | **`HACK_TEST=false go test ./pkg/foo/...`** (推荐) |
| 长期不要 compose 前置 | `.devx/config.json` 设 `hack_test: false`, 再 `devgen make` |

注意: 在 **`hack_test: true`** 已生成的前提下, **`HACK_TEST=false make test` 仍会执行 `hack_dep_run`** (compose 会起). 仅 Skip 集成用例请用 **`go test` 单包**, 或关闭 `hack_test` 后重新生成 Makefile.

### hack/ 目录约定

| 文件                      | 职责                                                        |
|---------------------------|-------------------------------------------------------------|
| `hack/docker-compose.yml` | 测试用 MySQL/TiDB/Redis/Pulsar 等, 端口映射到本机非默认端口 |
| `hack/scripts/`           | 初始化 SQL/入口脚本                                         |
| `hack/testing*.go`        | 可选: 封装 `WithMysql` 等, 供 `*_test.go` 调用              |

集成测 helper 常见模式 (confx/sqlx 等库):

- `hack.Check(t)`: `HACK_TEST != "true"` 时 `t.Skip`
- 用 retry 等待 compose 就绪后再 `Endpoint.Init`

### Agent 操作清单

1. 确认 `hack/docker-compose.yml` 已包含被测依赖, 端口与 DSN 与测试一致.
2. 本地集成测: `make hack_dep_run` 或 `make test` (需 `hack_test: true`).
3. 改 compose: `make hack_dep_stop` 再 `make hack_dep_run`.
4. 纯单元测: `HACK_TEST=false go test ./pkg/foo/...` (勿依赖 `make test` 跳过 compose).

## 2. 构建二进制 (targets)

devx 在 `make.target[]` 中为每个 `cmd/<name>` 生成:

| 根 Makefile     | 作用                         |
|-----------------|------------------------------|
| `target_<name>` | `make -C cmd/<name> install` |
| `targets`       | 所有 `target_*` 的聚合       |

`cmd/<name>/Makefile` (模板 `target.tpl.mk`):

- `build`: 当前 GOOS/GOARCH 下 `go build`, 产出 `<name>` 与 `<name>-<goos>-<goarch>`
- `install`: 写入仓库根 `dist/<name>/` (二进制, `version` 文件, 可选 `config/`)

版本注入: `-ldflags` 写入 `main.Name/Feature/Version/CommitID/CommitAt/BuildAt` (与 `appx.Meta` 字段对应).

```bash
make target_srv-mm28          # 单服务
make targets                  # config 里声明的全部 cmd
```

## 3. Docker 镜像 (images)

在 `make.target[]` 中为条目配置 `image` 对象时, devx 额外生成:

- `cmd/<name>/Dockerfile` (`target.tpl.dockerfile`)
- 根 `image_<name>` / `images`

`cmd/<name>/Makefile`:

```make
IMAGE_NAME ?= $(NAME)
IMAGE_TAG  ?= $(if $(GIT_TAG),$(GIT_TAG),latest)-$(GIT_COMMIT)

image:
	docker build -f Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) ../..
```

镜像构建流程 (Dockerfile):

1. **builder**: `COPY` 全仓库, `make target_<name>` → `dist/<name>/<name>`
2. **runtime**: 拷贝 `/app/app`, `version`, 可选 `config/`, 非 root 用户, 可选 `EXPOSE` / `ENTRYPOINT`

```bash
make image_srv-mm28
make images
# 或
make -C cmd/srv-mm28 image
```

镜像构建上下文为 **仓库根** (`../..`), 修改依赖或 cmd 后需重新 `make image_*`.

## 4. 其它常用根目标

| 目标                   | 说明                                                          |
|------------------------|---------------------------------------------------------------|
| `show`                 | 打印 module/git/工具链/env (含 `HACK_TEST` 若已 export)       |
| `dep` / `upgrade-dep`  | 安装 golangci-lint, goimports-reviser, git-chglog 等          |
| `tidy`                 | `go mod tidy` (+ vendor 若存在)                               |
| `lint`                 | golangci-lint + go vet                                        |
| `cover` / `view-cover` | 覆盖率 (`hack_test: true` 时同样前置 `hack_dep_run`)          |
| `ci-cover`             | lint + cover                                                  |
| `fmt` / `fmt-check`    | goimports-reviser                                             |
| `clean`                | 清理 cover.out / build 等                                     |
| `changelog`            | git-chglog                                                    |
| `pre-commit`           | dep fmt lint view-cover changelog; 有 target 时追加 `targets` |

## 5. 配置入口 (devx)

`.devx/config.json` 片段:

```json
{
	"make": {
		"hack_test": true,
		"target": [
			{
				"name": "srv-mm28",
				"image": {
					"runtime": "alpine:latest",
					"expose": "8080"
				}
			},
			{
				"name": "srv-robot"
			}
		]
	}
}
```

变更后执行: `devgen make` 或 `devgen all` 重新生成 Makefile/Dockerfile.

## 参考

- [devx README](https://github.com/xoctopus/devx/blob/main/README.md) — devgen 安装, 配置字段, 生成文件列表
- [makefile-guideline.md](references/makefile-guideline.md) — 模板字段与 multi-app 对照
- [genx manual](https://github.com/xoctopus/genx/blob/main/.agents/skills/genx/references/manual.md) — 业务仓 `internal/cmd/gen` 与 `go tool skill-install`
- `github.com/xoctopus/devx/internal/devx/devgen_make.go`
- `github.com/xoctopus/confx/hack/` — compose + testing helper 样例
