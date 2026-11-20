# M14 系统类加载器

## 目标

把隐式的「查找顺序约定」显式化为双加载器结构（Bootstrap + 用户加载器）与委派模型，收尾全书；跑通最复杂验收程序集。

## 对应书章节

第 14 章「系统类加载器」。

## 前置依赖

M13。

## 任务清单

### 1. BootstrapClassLoader — `heap/class_loader.go` 拆分

- `bootstrapLoader`：仅从 `-Xjre` 的 jre 路径（`jre/lib/*` → rt.jar，经 M02 缓存版 ZipEntry）加载
- `appLoader`（用户加载器，单一实例）：从 `-cp` 加载；**parent = bootstrap**
- 委派：`LoadClass` 先查 parent → 未命中再自己找（学习型简化：无自定义 ClassLoader 字节码支持，仅引擎内置两层；文档记录该简化）

### 2. `-Xbootclasspath` 前置路径

- `Xbootclasspath/a:<path>` 追加到 boot 段尾部（Override 类调试用，可选实现）
- `-Xbootclasspath:<path>` 完全替换（实现其一即可，优先 `/a:`）

### 3. ZipEntry 性能收尾

- bootstrap 段 rt.jar 全程共用单个缓存 reader（M02 已做条目索引，回归确认无重复打开）
- `-verbose:class` 输出加载来源：`[Loaded java/lang/Object from rt.jar]` / `[Loaded Foo from file:...\tests\classes]`

### 4. `-Xtest` 退役

- M05 引入的临时调试入口移除（printStackTrace/全链路输出已够用）；如需保留改为 `-Xdebugtest` 并从文档移除推荐用法

### 5. 最终回归与端到端验收

`tests/java/final/`（复杂程序集）：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `Collections`（ArrayList/HashMap/TreeMap 增删改查遍历排序） | 输出与 java 一致 |
| 2 | `Recursion`（递归数据结构：二叉树构建+三种遍历） | 一致 |
| 3 | `MazeSolver`（BFS/DFS + 二维数组） | 一致 |
| 4 | `StringFormat`（StringBuilder 格式化报表，不用 String.format——Formatter 链超范围则记录） | 一致（或标注 stretch 未过） |
| 5 | `PolymorphismSuite`（接口/抽象类/多态分派/instanceof 矩阵） | 一致 |
| 6 | 全部历史测试程序集（M01–M13）回归 | 全绿 |
| 7 | `go test ./... -cover` | 核心包 ≥ 80% |
| 8 | `go vet ./...` | 无告警 |

### 6. 收尾文档

- README：里程碑全勾、运行示例、已知限制（不做 GC/JIT/indy、反射仅 Array 子集、单线程）
- docs/DESIGN.md：补记实现偏差清单（与书/与 JVM 规范的所有简化点）

## 关键难点

- 委派模型下数组类与 String 等基础类的加载入口全部要换到 bootstrapLoader（排查所有 `LoadClass` 调用点）
- 历史行为回归：M06 起依赖「java/** 走 boot 段」的隐式约定，拆分后必须逐个集成测试回归
- Collections/TreeMap 纯 Java 实现复杂（红黑树 + 泛型擦除 + Comparator 接口调用）——是全引擎正确性的终极压力测试

## 完成动作

- commit：`feat: 双加载器与委派模型`
- tag：`m14`
- 更新 README 进度表 + 已知限制清单
- （可选扩展，不在 14 章范围）后续练习：Thread.start/run native + monitorenter/monitorexit + wait/notify → 多线程程序支持
