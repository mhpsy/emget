# emget

> Emby 媒体下载器,TUI + CLI,单连接、断点续传、版本化。
>
> 语言: [English](README.md) · **中文**

[![CI](https://github.com/mhpsy/emget/actions/workflows/ci.yml/badge.svg)](https://github.com/mhpsy/emget/actions/workflows/ci.yml)

一个 Go 写的命令行工具,可以从 Emby 服务器下载电影和剧集。**严格串行**下载(不并发,避免服务端风控),支持 Range 断点续传和退避重试。除了 bubbletea TUI 之外,还提供四个独立的子命令,方便脚本集成。

## 安装

### 从源码编译

```sh
go install github.com/mhpsy/emget/cmd/emget@latest
```

或从仓库克隆编译:

```sh
make build   # 生成带 ldflags 版本信息的 bin/emget
```

需要 Go 1.24+。

### 从 Release 下载

到 [releases 页面](https://github.com/mhpsy/emget/releases) 下载对应平台的压缩包,解压后把 `emget`(Windows 为 `emget.exe`)放入 `PATH`。

支持平台:linux/amd64、linux/arm64、darwin/amd64、darwin/arm64、windows/amd64。

## 首次配置

首次运行时,emget 会写入配置模板然后退出:

```sh
emget
# emget: wrote template to <路径> — fill it in and re-run
```

编辑模板(至少填写 `emby.*` 段),然后再次运行。

### 各平台目录

| 系统    | Config                                        | Cache                        | Data                                      |
|---------|-----------------------------------------------|------------------------------|-------------------------------------------|
| Linux   | `$XDG_CONFIG_HOME/emget` 或 `~/.config/emget` | `~/.cache/emget`             | `~/.local/share/emget`                    |
| macOS   | `~/Library/Application Support/emget`         | `~/Library/Caches/emget`     | `~/Library/Application Support/emget`     |
| Windows | `%AppData%\emget`                             | `%LocalAppData%\emget\cache` | `%AppData%\emget`                         |

运行 `emget config --paths-only` 可打印本机的具体路径。

## 使用

```
emget                  启动 TUI(默认)
emget tui              启动 TUI(显式)
emget tasks [flags]    列出任务(按状态分组)
emget clean [flags]    清理任务记录
emget config [flags]   打印配置路径和内容
emget version          打印版本和构建信息
emget help             显示帮助
```

### TUI 按键

- `↑/↓` 或 `k/j` —— 移动光标
- `PgUp / PgDn` —— 翻半屏
- `Home / g`,`End / G` —— 跳到首/尾
- `Enter` —— 打开详情 / 展开季
- `Space` —— 多选切换
- `Tab` —— 展开 / 切换面板
- `d` —— 入队选中项
- `p` —— 打开进度屏(全局)
- `esc` —— 返回
- `Ctrl+C` —— 退出

进度屏额外:
- `r` —— 重试失败任务
- `x` —— 取消排队任务

### `emget tasks`

```sh
emget tasks                        # 全部任务,按状态分组
emget tasks --status=failed        # 只看失败
emget tasks --status=queued,completed
emget tasks --format=json          # JSON 输出(供脚本消费)
```

### `emget clean`

```sh
emget clean                # 交互确认后清空
emget clean --yes          # 跳过确认
emget clean --completed-only
emget clean --failed-only
```

两个互斥的过滤 flag 同时传会直接报错。

### `emget config`

```sh
emget config               # 打印路径 + 解析后的 YAML(密码已打码)
emget config --paths-only  # 只打印路径
emget config --raw         # 原样打印文件(密码不打码,带警告)
```

## 剧集匹配规则

在 Detail Series 屏按 `d` 时,emget 会用 `config.yaml` 里的规则为每一集自动挑一个版本和若干字幕流:

```yaml
subtitles:
  preferred_languages: [zho, chi, eng]   # 外挂字幕语言优先级(顺序敏感)

versions:
  resolution_order: [2160, 1080, 720, 480]   # 优先分辨率(从高到低)
  keyword_boost: [BluRay, REMUX, WEB-DL]     # 平分时的加权关键字(大小写不敏感)
```

每个 MediaSource 的得分是 `分辨率分 × 1000 + 关键字分`。某一集如果没有任何匹配版本,会被 *跳过*(不算失败),最终 flash 消息汇报跳过数量。

## 启动恢复

如果 emget 退出时还有未完成的下载,下次启动会弹出恢复屏:

- `[Y]` 恢复 —— 重新入队未完成任务(从 `.part` 文件按 Range 续传)
- `[N]` 清理 —— 清空 state 文件从零开始
- `[esc]` 跳过 —— 保留 state 不动,直接进入搜索屏

日志文件超过 10 MiB 时会自动轮转,保留一份旧文件 `<log>.1`。

## 版本演进

- **v0.1** —— 电影:搜索、版本单选、外挂字幕多选、串行下载 + 断点续传 + 重试
- **v0.2** —— 剧集:季/集多选、按规则自动匹配入队
- **v0.3** —— 稳健化:启动恢复、日志轮转、重试/取消按键、goroutine panic 兜底
- **v0.4** —— CLI 子命令(`tasks` / `clean` / `config` / `version`)、Windows/macOS 支持、列表滚动、GitHub 自动多平台构建

## 开发

```sh
make build    # 带 ldflags 编译到 bin/emget
make test     # 单测
make vet      # go vet ./...
```

## License

MIT(`LICENSE` 文件待补)。
