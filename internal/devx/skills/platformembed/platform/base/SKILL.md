---
name: base
description: >-
  说明如何撰写文档与仓库提交信息.
  当需要编写或修改 `doc.go` / Markdown 文档, 或撰写 git commit message 时使用.
---




## 文档

1. 每个领域/模块包应该有 `doc.go`. 描述领域概况和对外契约用法.
2. 中文英文选择按照仓库惯例
3. `doc.go`, `*.md` 文档如果是中文, 使用半角标点, 切勿使用全角标点. 
4. `*.md` 注意表格对齐并控制表格宽度, 避免表格过宽影响阅读.

## 提交信息

1. 遵循 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) 规范
2. 中文英文选择按照仓库惯例
3. 如果是中文, 使用半角标点, 切勿使用全角标点. 
4. 多个改动点提交格式例子如下: summary 和 detail 需要提炼重点, 切勿啰嗦.
5. 给两个结果 1) 命令行 2) 提交信息文案

```
git commit -m "feat(module): summary mainly" \
-m "- change detail 1
- change detail 2
..."
```

6. 通常不要主动做提交动作, 除非明确要求