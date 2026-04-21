# emget — Emby 命令行下载器设计文档

- **项目代号**:emget
- **日期**:2026-04-21
- **作者**:mhpsy + Claude (brainstorming 会话)
- **目标读者**:后续实现者

## 1. 背景与目标

用户有一台 Emby 服务器(`https://v1.uhdnow.com`),希望在本地机器上用一个 Go 命令行工具完成:

- 以 TUI 方式搜索影片(电影、剧集)
- 查看某部影片可用的版本(MediaSources)与字幕(MediaStreams)
- 选择单个版本 + 勾选多个字幕,下载到本地
- 剧集支持批量下载(一整部剧、选定季、选定集)
- 下载过程须**串行**进行,避免同时多连接触发服务端风控
- 支持断点续传、任务状态持久化、有限次智能重试

### 非目标(YAGNI)

- 不做完整的 Emby SDK(只封装用到的端点)
- 不做后台 daemon 模式;TUI 关闭即停止
- 不做多线程/多连接并发下载
- 不追求 Windows 兼容(首发聚焦 Linux/macOS)
- 不做自动化 E2E(连接用户私有服务器)

## 2. 关键需求决策

brainstorming 阶段已锁定的选择:

| 领域 | 决策 |
|------|------|
| 交互方式 | 纯 TUI(bubbletea) |
| 配置管理 | `~/.config/emget/config.yaml`(XDG),`.env` 不再使用 |
| 下载路径 | 配置文件默认 + 按影片类型自动组织(`Movies/`、`TV/`) |
| 电影详情 | 版本单选 + 字幕多选,外挂字幕与视频同目录同主名加语言后缀 |
| 剧集详情 | 按规则自动匹配版本/字幕,支持多选季/多选集,异常集跳过并记日志 |
| 版本优先级 | 分辨率顺序为主 + 关键词加权作 tie-breaker |
| 断点续传 | 支持;任务状态持久化到 `~/.local/share/emget/state.json` |
| 重试 | 4xx 不重试(鉴权/404 等直接报错);5xx/网络/超时重试 3 次,指数退避带抖动 |
| 下载节流 | 每两次下载之间 `delay + [0, jitter)` 随机睡眠,默认 3s + 0~2s |

## 3. 总体架构

```
emget/
├── cmd/emget/main.go                 # 入口:加载配置 → 启动 TUI
├── internal/
│   ├── config/                       # 配置读写:~/.config/emget/config.yaml
│   │   ├── config.go                 # Config 结构体、加载、默认值
│   │   └── session.go                # ~/.cache/emget/session.json 的读写
│   ├── emby/                         # Emby REST 客户端(只封装用到的端点)
│   │   ├── client.go                 # HTTP 客户端、认证、token 管理
│   │   ├── search.go                 # 搜索 items
│   │   ├── items.go                  # 获取电影/剧集/季/集详情
│   │   ├── stream.go                 # 构造视频/字幕下载 URL
│   │   └── types.go                  # Emby 返回 JSON 的 Go 结构体
│   ├── downloader/                   # 下载核心
│   │   ├── downloader.go             # 单文件下载:Range 续传、重试、进度回调
│   │   ├── queue.go                  # 任务队列:串行 + 节流 + 持久化
│   │   └── naming.go                 # 输出文件/目录命名规则
│   ├── state/                        # 任务状态持久化
│   │   └── store.go                  # ~/.local/share/emget/state.json 原子读写
│   ├── matcher/                      # 剧集批量下载的版本/字幕匹配规则引擎
│   │   └── matcher.go                # PickVersion / PickSubtitles(纯函数)
│   └── tui/                          # bubbletea 视图层
│       ├── app.go                    # 根 Model(screen 路由)
│       ├── search.go                 # 搜索框 + 结果列表
│       ├── detail_movie.go           # 电影详情:选版本 + 多选字幕
│       ├── detail_series.go          # 剧集详情:选季/集/规则
│       └── progress.go               # 下载进度 + 队列视图
└── go.mod
```

### 数据流

```
TUI (bubbletea) ─┬─► emby.Client ──► Emby 服务器
                 └─► downloader.Queue ──► downloader.Downloader ──► Emby 服务器
                                      └──► state.Store (持久化)
```

### 模块边界

- `emby.Client` 不感知 TUI、不感知下载;纯 REST 调用,返回 DTO
- `downloader` 不调 Emby;接收 URL + headers,做文件 I/O、续传、重试
- `tui` 不做业务,只装配数据、把用户选择转成 `Task` 入队
- `state.Store` 是 `downloader.Queue` 的依赖;Queue 对 JSON 格式无感
- `matcher` 纯函数,无副作用

## 4. TUI 设计

### 屏幕地图

```
┌─────────────┐    enter     ┌────────────────┐
│  Search     │─────────────►│  Result List   │
│  (输入框)    │              │  (电影/剧集)    │
└─────────────┘              └────────┬───────┘
       ▲                              │ enter
       │ esc                          ▼
       │           ┌──────────────────┴────────────────┐
       │           │                                   │
       │           ▼                                   ▼
       │  ┌────────────────┐              ┌────────────────────┐
       │  │ Detail: Movie  │              │ Detail: Series     │
       │  │ - 版本单选       │              │ - 规则(版本优先级) │
       │  │ - 字幕多选       │              │ - 字幕语言多选       │
       │  │ - [d] 加入队列   │              │ - 季/集多选          │
       │  └────────┬───────┘              │ - [d] 加入队列        │
       │           │                      └──────────┬─────────┘
       │           └──────────┬──────────────────────┘
       │                      ▼
       │           ┌──────────────────────┐
       │           │  Progress / Queue    │  ◄── [p] 任意位置可打开
       │           │  - 当前下载进度       │
       │           │  - 等待队列           │
       │           │  - 已完成 / 失败     │
       │           └──────────────────────┘
```

### 按键表

| 位置 | 键 | 动作 |
|------|------|------|
| 全局 | `q` / `ctrl+c` | 退出(有活跃任务时二次确认) |
| 全局 | `p` | 打开"进度/队列"面板 |
| Search | 输入 → `enter` | 发起搜索 |
| Result List | `↑/↓` `enter` `esc` | 导航 / 进入详情 / 返回 |
| Detail Movie | `↑/↓` 切版本,`space` 勾字幕,`d` 加队列,`esc` 返回 |
| Detail Series | `tab` 切换"规则 / 季集列表"区,`space` 勾选,`d` 加队列,`esc` 返回 |
| Queue | `r` 重试失败项,`x` 取消任务,`esc` 返回 |

### 启动流程

1. 读配置(必要字段缺失 → 打印路径 + 模板后退出)
2. 尝试读 session 缓存 → 调 `/System/Info` 验活
3. 未认证或验活失败 → 用配置里的账密做一次 `AuthenticateByName`;失败则提示并退出
4. 若 state.json 中有未完成任务 → 进 Search 前询问"检测到 N 个未完成任务,是否继续? [Y/n]"

### 交互细节

- 搜索/详情加载期间显示 spinner;失败用红字底栏提示并支持按 `r` 重试
- 加入队列后立刻把任务推入 Queue,state 即时落盘,TUI 底栏提示"已加入 N 个任务"
- 队列视图分三段:**正在下载**(进度条/速度/ETA) / **排队中** / **已完成 & 失败**
- 字幕下载作为独立子任务条目显示(与视频任务并排,通过 `parent_id` 关联到同一影片)

## 5. Emby 客户端层

```go
type Client struct {
    baseURL    *url.URL
    httpClient *http.Client
    session    *Session
}

type Session struct {
    AccessToken string    `json:"access_token"`
    UserID      string    `json:"user_id"`
    ServerID    string    `json:"server_id"`
    DeviceID    string    `json:"device_id"`    // 首次生成 UUID
    ExpiresAt   time.Time `json:"expires_at"`   // 客户端自估,保守
}

func NewClient(baseURL string, httpClient *http.Client) *Client
func (c *Client) Authenticate(ctx context.Context, username, password string) (*Session, error)
func (c *Client) SetSession(*Session)
func (c *Client) IsAuthenticated() bool
```

### 端点封装

| 方法 | Emby 端点 |
|------|----------|
| `Search(term, types, limit)` | `GET /Users/{UserID}/Items?SearchTerm=...&IncludeItemTypes=Movie,Series&Recursive=true` |
| `GetItem(id)` | `GET /Users/{UserID}/Items/{Id}?Fields=MediaSources,Path,Overview` |
| `GetSeasons(seriesID)` | `GET /Shows/{seriesID}/Seasons?UserId={UserID}` |
| `GetEpisodes(seriesID, seasonID)` | `GET /Shows/{seriesID}/Episodes?SeasonId={seasonID}&UserId={UserID}&Fields=MediaSources` |
| `VideoDownloadURL(item, sourceID)` | `GET /Items/{id}/Download?api_key={token}&mediaSourceId={sourceID}` |
| `SubtitleDownloadURL(item, sourceID, streamIndex, ext)` | `GET /Videos/{id}/{sourceID}/Subtitles/{streamIndex}/Stream.{ext}?api_key={token}` |

### 请求头

所有业务端点带以下头;**`AuthenticateByName` 除外**(预认证时无 token,此时 Authorization 头去掉 `Token=...` 部分):

```
Authorization: MediaBrowser Token="<AccessToken>", Client="emget",
               Device="<hostname>", DeviceId="<UUID>", Version="<build-version>"
```

### 认证刷新策略

- 401 **不自动重认证**(避免密码错用触发服务端锁定);返回 `ErrUnauthorized`
- 由 `main` 层捕获一次,用当前 config 凭据重调 `Authenticate`,再重试原请求一次
- 二次仍失败 → 写日志并终止 CLI,提示用户检查凭据

### 错误类型

```go
var (
    ErrUnauthorized = errors.New("emby: unauthorized")
    ErrForbidden    = errors.New("emby: forbidden")
    ErrNotFound     = errors.New("emby: not found")
)

type EmbyError struct {
    Status int
    Body   string
    URL    string
}
```

## 6. 下载器

### Task 结构

```go
type TaskKind string
const (
    KindVideo    TaskKind = "video"
    KindSubtitle TaskKind = "subtitle"
)

type Task struct {
    ID          string     // uuid
    ParentID    string     // 剧集任务分组,可空
    Kind        TaskKind
    ItemID      string     // Emby item id
    SourceID    string     // MediaSource id
    StreamIndex int        // 字幕流索引,视频任务为 -1
    URL         string
    OutputPath  string     // 最终绝对路径
    TotalSize   int64      // Content-Length,0 表示未知
    Downloaded  int64
    Status      Status     // queued | downloading | completed | failed
    Attempts    int
    LastError   string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 单文件流程

下载中使用**临时文件名** `<OutputPath>.part`,完成后重命名为 `OutputPath`。这样既避免播放器扫到半截文件,又让中断后下次启动能识别出"待续传"的目标。

1. Stat `<OutputPath>.part` 取得已下载字节数(文件不存在视为 0)
2. 若 >0:HEAD 确认服务端支持 Range 且 Content-Length ≥ 已下载;不满足则告警并截断重下
3. GET 带 `Range: bytes=<existing>-`
4. 以 `O_APPEND|O_WRONLY` 写入 `.part` 文件;每 ~1MB 触发 progress 回调
5. 完成后 `fsync` → `rename` 去掉 `.part` 后缀
6. 出错按重试策略处理(见下)

### 重试策略

```go
func shouldRetry(err error) bool {
    if errors.Is(err, emby.ErrUnauthorized) || errors.Is(err, emby.ErrForbidden) { return false }
    var ee *emby.EmbyError
    if errors.As(err, &ee) && ee.Status >= 400 && ee.Status < 500 { return false }
    // 网络错误、5xx、超时 → 重试
    return true
}
```

- 最多 **3 次重试**(加首次尝试,总计最多 4 次)
- 退避序列 1s / 2s / 4s + ±500ms 抖动(`retry_backoff` 为基数,每次 x2)
- 重试从当前 `Downloaded` 继续,不丢进度

### Queue 节流

```go
for task := range q.tasks {
    q.store.MarkDownloading(task.ID)
    q.emit(TaskStarted{task})
    err := q.downloader.Run(ctx, task, q.onProgress)
    if err != nil {
        q.store.MarkFailed(task.ID, err)
        q.emit(TaskFailed{task, err})
    } else {
        q.store.MarkCompleted(task.ID)
        q.emit(TaskCompleted{task})
    }
    sleep := q.delay + time.Duration(rand.Int63n(int64(q.jitter)))
    select {
    case <-ctx.Done(): return
    case <-time.After(sleep):
    }
}
```

### 文件命名

- 电影:`<output>/Movies/<Title> (<Year>)/<Title> (<Year>).<ext>`
- 电影字幕:`<output>/Movies/<Title> (<Year>)/<Title> (<Year>).<lang>.<ext>`
- 剧集:`<output>/TV/<Series>/Season <NN>/<Series> - S<NN>E<NN> - <EpisodeTitle>.<ext>`
- 剧集字幕:对应集同目录同主名 + 语言后缀
- 非法字符(`/\:*?"<>|`)替换为 `_`
- 路径超 255 字节截断
- `<lang>` 来自 `MediaStream.Language`(ISO 639-2/T);缺失则 `und`
- 同语言多条追加 `.forced` / `.sdh` / `.2` 等后缀

## 7. 配置与状态

### 7.1 config.yaml

路径:`$XDG_CONFIG_HOME/emget/config.yaml`(默认 `~/.config/emget/config.yaml`)

```yaml
emby:
  endpoint: https://v1.uhdnow.com
  username: mhpsy
  password: <fill-me>

download:
  output_dir: ~/Media
  movies_subdir: Movies
  tv_subdir: TV
  inter_download_delay: 3s
  jitter: 2s
  max_retries: 3
  retry_backoff: 1s
  user_agent: "emget/0.1.0"

subtitles:
  preferred_languages: [zho, chi, eng]   # 剧集批量下载的语言过滤

versions:
  resolution_order: [2160, 1080, 720, 480]
  keyword_boost: [BluRay, REMUX, WEB-DL]

logging:
  level: info
  file: ~/.local/share/emget/emget.log
```

加载优先级:**CLI flag > 环境变量(`EMGET_*`)> 配置文件 > 默认值**

首次启动若配置不存在 → 写模板 → 打印路径 → 退出(exit 0,用户填好再跑)。

### 7.2 session.json

路径:`$XDG_CACHE_HOME/emget/session.json`(默认 `~/.cache/emget/session.json`),权限 `0600`

```json
{
  "access_token": "...",
  "user_id": "...",
  "server_id": "...",
  "device_id": "stable-uuid",
  "expires_at": "2026-05-21T00:00:00Z"
}
```

### 7.3 state.json

路径:`$XDG_DATA_HOME/emget/state.json`(默认 `~/.local/share/emget/state.json`)

```json
{
  "version": 1,
  "tasks": [
    {
      "id": "uuid",
      "parent_id": "series-task-uuid",
      "kind": "video",
      "item_id": "emby-id",
      "source_id": "...",
      "stream_index": -1,
      "url": "https://...",
      "output_path": "/home/.../S01E01.mkv",
      "total_size": 1234567890,
      "downloaded": 456789000,
      "status": "downloading",
      "attempts": 0,
      "last_error": "",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

**写入原则**:
- 原子写:`state.json.tmp` → `rename`
- 状态变更(入队/开始/完成/失败)立即写
- 下载中每 2 秒批量更新 `downloaded` 字段
- 启动时把遗留的 `downloading` 任务重置为 `queued`(由 Queue 重新调度,从 `.part` 续传)

## 8. Matcher(规则引擎)

```go
type VersionRule struct {
    ResolutionOrder []int
    KeywordBoost    []string
}
func PickVersion(sources []MediaSource, rule VersionRule) (*MediaSource, error)

type SubtitleRule struct {
    Languages []string
    External  bool    // 默认 true:只选外挂
}
func PickSubtitles(streams []MediaStream, rule SubtitleRule) []MediaStream
```

### PickVersion

1. 对 source 推断分辨率(优先 `MediaStreams[video].Height`,否则从 source name 正则 `\b(4k|2160p?|1080p?|720p?|480p?)\b` 猜)
2. 每个 source 得分 = `resolution_score * 1000 + keyword_score`
   - `resolution_score = len(ResolutionOrder) - index`:`ResolutionOrder=[2160,1080,720,480]` 时,2160 得 4,1080 得 3,720 得 2,480 得 1;不在列表或推断不出 → 0
   - `keyword_score`:source name 大小写不敏感命中 `KeywordBoost` 的数量
3. 取最高分。全为 0 → `ErrNoMatch`(剧集批量时:跳过该集,记入失败日志)
4. 同分时按 `MediaSources` 原顺序稳定,取靠前者

### PickSubtitles

1. 过滤 `IsExternal==rule.External` 且 `Language in rule.Languages`
2. 按 `Languages` 顺序稳定排序返回(用户优先级明确)
3. 空列表正常返回(有视频无字幕也算成功)

## 9. 错误处理、日志、可观测性

### 错误分层

```
emby.Client  →  ErrUnauthorized / ErrNotFound / ErrServer / *EmbyError
downloader   →  *DownloadError{Kind, Task, Wrapped}
                Kind: auth | network | io | server | canceled
queue        →  事件广播,不再包装
tui          →  用户可读文案 + [r] 重试 / [x] 取消
```

**原则**:
- 用 `errors.Is` / `errors.As`,不做字符串匹配
- 原始错误原文存 state 的 `last_error`,用户文案单独存 `user_message`

### 日志

- `log/slog`(stdlib,现代 Go)
- 双 handler:
  - **文件**:JSON 格式,带 `run_id` + `task_id`,路径由 `logging.file` 指定
  - **终端**:TUI 启动期间**完全禁止**写 stdout/stderr;仅 TUI 未启动时(CLI 解析错误、启动失败)才打 error 级到 stderr
- 简单轮转:启动时若日志 >10MB,rename 为 `emget.log.1`(只保留一份旧日志)

### panic 处理

- `main` 和 downloader goroutine 都 `defer recover`
- TUI 崩溃:还原终端 → 写日志 → stderr 打印 → 退出码 2

### Ctrl+C 语义

- 第一次:取消 `ctx`,提示"正在停止当前任务并保存进度…";当前下载 fsync + 落盘 state → 退出(<2 秒)
- 第二次:强退(最多丢 2 秒的下载字节数,`.part` 文件仍可续传)

## 10. 测试策略

| 层 | 类型 | 工具 | 覆盖目标 |
|----|------|------|---------|
| `matcher` | 单元 | stdlib | table-driven,>90% |
| `naming` | 单元 | stdlib | table-driven,>90% |
| `emby` | 单元 | `httptest` | 请求头、URL、错误映射、JSON 解析(fixtures 放 `testdata/`) |
| `downloader` | 单元 | `httptest` | Range 续传、重试退避、.part rename、取消 |
| `downloader` | 单元 | 注入 clock | 节流时间窗 |
| `queue` | 单元 | mock downloader | 串行性、事件广播、state 持久化时机 |
| `state` | 单元 | 临时目录 | 原子写、并发读写、version 字段兼容 |
| `config` | 单元 | 临时目录 | 默认值、优先级、YAML 反序列化 |
| `tui` | 单元 | `teatest`(charmbracelet 官方) | 关键屏幕导航、按键响应 |
| E2E | 手动 | 真实 Emby | 开发者对 `https://v1.uhdnow.com` 跑一次搜索+下载,README 有 "manual verification" 清单 |

**不做**:
- 自动化 E2E(CI 连不到私有服务器,且有下载目录污染风险)
- Windows 支持(首发聚焦 Linux/macOS)

**覆盖目标**:不追求百分比,但每条错误分支都要有显式用例;续传、重试、节流、matcher 规则 —— 每条路径至少一个测试。

## 11. 依赖清单

| 库 | 用途 |
|------|------|
| `github.com/spf13/cobra` | CLI 骨架(即使纯 TUI,也留 `--config` 等 flag 口) |
| `github.com/charmbracelet/bubbletea` | TUI 框架 |
| `github.com/charmbracelet/bubbles` | 常用组件(spinner、list、textinput、progress) |
| `github.com/charmbracelet/lipgloss` | 样式 |
| `github.com/charmbracelet/x/exp/teatest` | TUI 单测 |
| `gopkg.in/yaml.v3` | 配置解析(不引 viper,够用就行) |
| `github.com/google/uuid` | 任务/设备 ID |
| `log/slog` | 日志(stdlib) |

## 12. 构建与分发

- `go build -o bin/emget ./cmd/emget`
- 使用 Go 1.22+,`go.mod` 中 `go 1.22` 或更高
- Makefile 提供 `build`、`test`、`lint`、`coverage` 目标

## 13. 路线图

**MVP(v0.1.0)**:
1. config 加载 + session 管理
2. emby.Client(Authenticate、Search、GetItem)
3. downloader.Downloader 单文件下载(Range、重试、.part)
4. naming.go
5. TUI:Search → Result List → Detail Movie → Progress
6. 单电影下载 + 字幕下载走通

**v0.2.0 — 剧集支持**:
7. GetSeasons / GetEpisodes
8. matcher.go(PickVersion / PickSubtitles)
9. Detail Series 屏幕 + 批量入队
10. 异常集跳过并记日志

**v0.3.0 — 健壮性**:
11. state.Store 持久化 + 启动恢复对话框
12. Queue 节流
13. Ctrl+C 优雅停止
14. 日志轮转

后续版本按需排期,非首发 scope。
