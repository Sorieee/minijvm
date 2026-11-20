# M01 命令行工具

## 目标

实现类 `java` 命令的参数解析入口 `minijvm`，为后续所有里程碑提供命令入口。

## 对应书章节

第 1 章「命令行工具」。

## 前置依赖

M00（go.mod 就位）。

## 任务清单

### 1. Cmd 参数结构 — `internal/cmd.go`

```go
type Cmd struct {
    HelpFlag    bool     // -h / -help
    VersionFlag bool     // -version
    CpOption    string   // -cp / -classpath（Windows 分隔符 ';')
    XjreOption  string   // -Xjre <path>  指向 E:\tools\jdk8\jre\lib
    XssOption   string   // -Xss <size>   栈大小，如 -Xss1m / -Xss512k
    VerboseClassFlag bool // -verbose:class（M14 用）
    Class       string   // 主类全限定名
    Args        []string // 传给 main 的参数
}

func ParseCmd() *Cmd
func (cmd *Cmd) PrintUsage()
```

- 用标准库 `flag` 包实现，自定义 flag 顺序支持（java 允许 flag 与类名混排，M01 先按 flag 包默认规则即可）
- `-Xss` 解析成字节数：支持 `1m`、`512k`、纯数字；非法值报错退出

### 2. main 入口 — `cmd/minijvm/main.go`

- 调 `ParseCmd()`
- `-version` → 打印 `minijvm version 0.0.1` 后退出
- `-h` / 无类名 → 打印用法后退出
- 否则临时打印解析结果（classpath / xjre / class / args），占位提示「类加载在 M02 实现」

### 3. 版本号常量 — `internal/version.go`

```go
const VERSION = "0.0.1"
```

（M10 交叉编译时用 `-ldflags "-X minijvm/internal.VERSION=..."` 注入）

## 关键难点

- **classpath 分隔符**：书用 Linux 的 `:`，Windows 是 `;`。M01 只透传字符串，M02 拆分时必须用 `filepath.ListSeparator`，不要在文档/代码里硬编码 `:`
- `-Xss` 数值解析：`strconv.ParseInt` + 单位后缀处理，错误输入给友好报错

## 验收标准（TDD 先行）

先写 `internal/cmd_test.go`：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `-version` | VersionFlag = true |
| 2 | `-cp a;b;c Foo bar1 bar2` | CpOption="a;b;c"，Class="Foo"，Args=["bar1","bar2"] |
| 3 | `-Xss1m` / `-Xss512k` / `-Xss1024` | 解析为 1048576 / 524288 / 1024 |
| 4 | `-Xssabc` | 解析报错（友好信息退出码 ≠ 0） |
| 5 | `minijvm -version`（集成） | stdout 含 `minijvm version 0.0.1` |
| 6 | `minijvm`（无参数，集成） | 打印用法，退出码 ≠ 0 |

## 完成动作

- commit：`feat: 命令行参数解析`
- tag：`m01`
- 更新 README 进度表
