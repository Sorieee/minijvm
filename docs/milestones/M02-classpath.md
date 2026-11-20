# M02 搜索 class 文件

## 目标

实现 Classpath 体系：在文件目录、jar 包、通配符目录、组合路径中定位并读取类字节。

## 对应书章节

第 2 章「搜索class文件」。

## 前置依赖

M01（Cmd.CpOption / XjreOption）。

## 任务清单

### 1. Entry 抽象 — `internal/classpath/entry.go`

```go
type Entry interface {
    // 读取类字节；className 是全限定名（java/lang/Object，不含 .class）
    // 返回：字节、实际命中的 Entry（用于 -verbose 输出）、错误
    ReadClass(className string) ([]byte, Entry, error)
    String() string
}

func NewEntry(pathList string) Entry // 按分隔符拆分，多路径包成 CompositeEntry
func newSingleEntry(path string) Entry // 单路径分派：* 结尾→Wildcard；.jar/.zip→Zip；否则 Dir
```

### 2. 四种实现

| 文件 | 类型 | 行为 |
|------|------|------|
| `dir_entry.go` | DirEntry | 读 `<root>/<className>.class` |
| `zip_entry.go` | ZipEntry | `zip.OpenReader` 后按条目名读（M14 前每次打开即可，可缓存 *zip.ReadCloser 优化） |
| `composite_entry.go` | CompositeEntry | 依次尝试子 Entry |
| `wildcard_entry.go` | WildcardEntry | 目录下所有 `.jar/.zip` 组成 Composite（书不排序；我们加按文件名排序保证确定性） |

错误约定：读不到类返回统一 `ErrClassNotFound`，向上传递用于报错。

### 3. Classpath 三段结构 — `internal/classpath/classpath.go`

```go
type Classpath struct {
    bootClasspath Entry // jre/lib/*      ← -Xjre 指定目录
    extClasspath  Entry // jre/lib/ext/*  ← 无则空
    userClasspath Entry // -cp
}

func Parse(jreOption, cpOption string) *Classpath
func (cp *Classpath) ReadClass(className string) ([]byte, Entry, error)
// 查找顺序：boot → ext → user
```

- `-Xjre` 未指定时的兜底：依次探测 `JAVA_HOME/jre/lib`、常见安装路径；找不到给明确报错提示（M02 阶段要求显式传 `-Xjre E:\tools\jdk8\jre\lib`）
- jre 路径内通配：`<jre>/lib/*`（所有 jar）、`<jre>/lib/ext/*`

### 4. main 接入

- M02 阶段 `minijvm -Xjre ... -cp ... Foo`：三段查找 `Foo`，找到打印「Found class: Foo => <entry.String()>」+ 字节长度；找不到打印友好错误
- 打印 rt.jar 中 `java.lang.Object` 可读（验证 boot 段通配生效）

## 关键难点

- **Windows 路径**：全程 `filepath.Join` / `filepath.ListSeparator`；className 里的 `/` 与文件系统路径转换
- ZipEntry 每次读都开 zip 太慢：实现时直接读 central directory 缓存 `map[name]index`（书 M14 才优化，我们一步到位但保持接口不变）
- 通配符只匹配 `.jar`/`.zip`，不递归子目录（与书一致）

## 验收标准（TDD 先行）

先写 `internal/classpath/classpath_test.go`（测试数据：临时目录 + 用 `jar cf` 或 zip 库造测试 jar）：

| # | 用例 | 期望 |
|---|------|------|
| 1 | DirEntry 读存在的类 | 返回字节，首 4 字节 `CA FE BA BE` |
| 2 | 读不存在的类 | ErrClassNotFound |
| 3 | ZipEntry 读 jar 内类 | 返回字节 |
| 4 | CompositeEntry 依次尝试 | 第 2 个命中 |
| 5 | WildcardEntry（目录下 2 个 jar） | 都能命中各自类 |
| 6 | Classpath.Parse：jre 指向真实 jdk8 | ReadClass("java/lang/Object") 命中 rt.jar |
| 7 | user 优先于 boot？（反例：boot→ext→user 固定顺序） | 顺序固定，boot 先命中 |
| 8 | 集成：`minijvm -Xjre <jdk8> -cp <dir> HelloWorld` | 打印 Found class + 长度 |

## 完成动作

- commit：`feat: classpath 与四种 Entry 实现`
- tag：`m02`
- 更新 README 进度表
