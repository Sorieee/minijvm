# M12 字符串

## 目标

完善字符串支持：字符串常量池、`String.intern`，使字符串拼接、比较、intern 语义与真 JVM 一致。

## 对应书章节

第 12 章「字符串」。

## 前置依赖

M11（String 相关基础 native 链已通）。

## 任务清单

### 1. 字符串池 — `internal/heap/string_pool.go`

```go
var stringPool = struct {
    mu sync.Mutex
    m  map[string]*Object // Go 字符串 → 已 intern 的 String 对象
}{m: map[string]*Object{}}

func InternString(loader, goStr string) *Object // 查池 → 未命中则 JString 构造后入池
func GoString(jStr *Object) string              // String 对象 → Go 字符串（读 char[] value）
```

- JDK8 String 布局：字段 `char[] value` + `int hash`（char 数组对象由 `heap.JString` 创建）

### 2. ldc 字符串走池

- M06/M07 的 `ldc` 字符串常量改为 `InternString`（字符串常量自动 intern，JVM 语义）

### 3. `String.intern` native — `native/java_lang_string.go`

- `intern()Ljava/lang/String;` → 读对象内容 → `InternString` 命中池中对象返回

### 4. 辅助：引擎内部字符串交换统一走 GoString/InternString

- println 桥接残留（printStackTrace）、`-Xtest`、异常消息（`ThrowException` msg）统一用这两个函数，避免散落的字段解析代码

## 关键难点

- Go map 键是内容字符串（拷贝语义），intern 语义天然成立；但对象字段内容变了不会失效——JDK8 String value 实际不可变，安全
- 字符池与类加载器关系：池跨类加载器共享（单加载器实现下无歧义，M14 后注明该简化）

## 验收标准（TDD 先行）

`tests/java/string/`：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `InternTest`（两个同名常量 `==` 为 true；`new String("a").intern() == "a"` 为 true） | 输出与 java 一致 |
| 2 | `ConcatTest`（`+` 拼接、StringBuilder 链式、`+=` 循环拼接） | 一致 |
| 3 | `StringMethods`（equals/hashCode/substring/indexOf/replace/trim/split/toUpperCase） | 一致 |
| 4 | `CharTest`（char 数组构造 String、charAt、遍历） | 一致 |
| 5 | `StringCompare`（compareTo 排序一组词） | 一致 |
| 6 | StringPool 并发安全（`go test -race` 跑池单测） | 无竞争 |
| 7 | 全量回归 | 通过 |

## 完成动作

- commit：`feat: 字符串常量池与 String.intern`
- tag：`m12`
- 更新 README 进度表
