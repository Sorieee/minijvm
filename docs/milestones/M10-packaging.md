# M10 打包和部署

## 目标

完善类 `java` 命令的部署能力：支持运行 jar（`-jar` 读 MANIFEST），并输出跨平台可执行产物。

## 对应书章节

第 10 章「打包和部署」。

## 前置依赖

M09。

## 任务清单

### 1. `-jar` 模式 — `internal/cmd.go` + `internal/jvm.go`

- 新参数 `-jar xxx.jar`（与 `-cp` 互斥语义：jar 自身成为 user classpath）
- 读 jar 内 `META-INF/MANIFEST.MF` 解析 `Main-Class: com.x.Main`（注意 `Manifest-Version` 头与行折叠 RFC822 续行 ` 空格开头`）
- 主类名 `.` → `/` 转换后走既有加载流程

### 2. classpath 组合细化

- `-jar` 时 userClasspath = 该 jar 的 ZipEntry（其它 -cp 值忽略，与 java 行为一致）
- `-cp` 多段（`;` 分隔）已在 M02 支持，回归确认

### 3. 退出码规范化

- 正常结束 0；未捕获异常退出码 1（与 java 的 `Exit 1` 行为对齐——java 是异常直接退出码 1）
- `System.exit(n)` 在 M11 有 native 后生效，M10 先记录规范

### 4. Go 交叉编译与版本注入

```bash
# scripts/build.sh
GOOS=windows GOARCH=amd64 go build -ldflags "-X minijvm/internal.VERSION=$(cat VERSION)" -o dist/minijvm-windows-amd64.exe ./cmd/minijvm
GOOS=linux   GOARCH=amd64 go build ... -o dist/minijvm-linux-amd64
GOOS=darwin  GOARCH=amd64 go build ... -o dist/minijvm-darwin-amd64
GOOS=darwin  GOARCH=arm64 go build ... -o dist/minijvm-darwin-arm64
```

- `VERSION` 文件单一来源；`-version` 输出注入值
- 产物清单写进发布说明（本仓库 README 附录即可）

### 5. 用法文档

README「运行」章节补 `-jar` 用法与 `-Xss/-Xjre` 完整示例。

## 关键难点

- MANIFEST 续行解析（`\r\n` + 首空格续行）
- `-jar` 与 `-cp` 的互斥语义对齐真 java（给了 -cp 又给 -jar 时 java 会忽略 -cp，我们打印警告即可）

## 验收标准（TDD 先行）

| # | 用例 | 期望 |
|---|------|------|
| 1 | MANIFEST 解析单测（含续行、多余空行、CRLF） | Main-Class 正确 |
| 2 | 测试程序打成 jar：`minijvm -Xjre ... -jar app.jar a b` | main 收到 ["a","b"]，输出正确 |
| 3 | jar 内多包结构（com.test.*） | 正常加载执行 |
| 4 | `-jar` + `-cp` 同给 | 打印警告并按 java 语义（-cp 忽略）执行 |
| 5 | 未捕获异常退出码 | 1 |
| 6 | 交叉编译 4 个平台产物 | 均成功，`-version` 注入生效 |
| 7 | 集成回归 | M01–M09 全部测试程序仍通过 |

## 完成动作

- commit：`feat: 支持 -jar 运行与跨平台打包`
- tag：`m10`
- 更新 README 进度表
