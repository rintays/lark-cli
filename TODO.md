# TODO

## P0 - 影响脚本/JSON/可用性（必须先修）
- [x] stdout/stderr 分离：新增 `appState.ErrWriter`，stdout 仅输出数据，提示/日志走 stderr
- [x] 退出码区分：usage error -> 2，runtime error -> 1，成功 -> 0
- [x] Context 贯穿：根命令用可取消 context，所有命令/SDK 调用改用 `cmd.Context()`
- [x] 导出轮询可取消：`pollExportTask` 用 `select { case <-ctx.Done(): ... }`
- [x] config 保存原子写入（临时文件 + fsync + rename）

## P1 - CLI 交互与可脚本化
- [x] 分页截断时保留 `next_page_token` / `has_more`（至少 Drive list / Users search）
- [x] 列表型 flag 统一用 `StringSlice`（支持 `--x a,b` 与重复）：
  - [x] `drive --type`
  - [x] `messages search --from-id/--chat-id/--at-id`
- [x] 资源引用解析：支持直接贴 Lark/飞书 URL（docx/sheet/file 等）
  - [x] drive info/download/export
  - [x] docs get/info/overwrite/export
  - [x] sheets read/info/update/append/clear/delete
- [x] 时间参数更友好：支持 unix seconds / RFC3339 / 相对时间（-24h/-7d）
  - [x] messages search --start-time/--end-time
- [x] stdin/stdout 约定：
  - [x] 读文件 flag 支持 `-` 表示 stdin（content-file/raw-file 等）
  - [x] 写文件 flag 支持 `--out -` 输出 stdout（drive/docs export/download）
- [x] completion 子命令：`lark completion bash|zsh|fish|powershell`

## P2 - 维护性/错误信息
- [x] SDK 初始化错误更可诊断：保存 init 错因，统一 `requireSDK(...)` 返回具体提示
- [x] `token` 刷新等 verbose 信息走 stderr
- [x] JSON 参数本地校验：
  - [x] base record search 的 `--filter/--filter-json/--sort` 校验 JSON 有效性与互斥
- [x] 结构化执行模板（runWithToken）：统一 token 获取/SDK 调用/输出/scope hint（逐步迁移）
- [x] authregistry 元数据靠近命令定义（annotations 或生成器）
- [x] app_secret 支持 keyring（可选；CI 仍可 env 注入）

## 参考对标（gogcli）
- [x] stdout/stderr 输出边界：data 输出与 UI/日志分层
- [x] exit code 区分 usage/runtime
- [x] completion 子命令/脚本输出
- [ ] 非交互/强制执行模式（本轮不做）
