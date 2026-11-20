# M15 垃圾回收（GC）

## 目标

为 MiniJVM 实现自研垃圾回收：**对象注册表 + 标记-清除（mark-sweep）**，支持 `-Xmx` 堆上限、`-verbose:gc` 日志、`System.gc()`、`java.lang.OutOfMemoryError`。书未覆盖此主题（jvmgo 直接依赖 Go 宿主 GC），属扩展里程碑。

## 对应书章节

书外扩展（书第 14 章后的练习方向）。

## 前置依赖

M14 完成。具体依赖：M06/M07 对象与数组模型、M09 异常机制（OOM 需要抛真异常）、M12 字符串池、M13 数组。

## 设计决策（实现前先通读）

### 1. 架构选择：对象注册表 + 标记-清除，**不引入句柄/偏移间接层**

- 新增 `heap` 包内账本：`Heap{maxBytes, usedBytes int64, table map[*Object]gcMeta}`
- **所有分配点统一注册**：`NewObject`、`newarray/anewarray/multianewarray`、`JString`、JClass 创建等，分配时累计 `usedBytes`
- `Slot.Ref` 保持 `*Object` Go 指针——**指令层与解释器零改动**
- 我们的 GC 决定对象的生死（标记可达 → 清除时从注册表注销）；真实内存的最终释放仍由 Go GC 完成。JVM 层语义完整：垃圾识别、回收、`-Xmx` 配额、OOM、GC 日志全部真实
- 备选方案（句柄表/字节 arena + free-list）更接近真实 JVM，但改造面覆盖全部指令，放弃；在本文档记录取舍即可

### 2. 触发与模式

- **STW**：引擎单线程，GC 只在分配路径同步触发（分配前 `used+need > max` → full GC），天然 stop-the-world
- 回收后仍不足 → 抛 `java.lang.OutOfMemoryError: Java heap space`（复用 M09 异常工厂）
- `System.gc()` native → 显式触发一次 full GC
- `-verbose:gc` 每次回收打印一行统计（根数/标记数/清除数/释放字节/耗时）

### 3. 永生对象（不参与回收，明示边界）

- 类元数据：`heap.Class/Method/Field`（≈ Metaspace，永不回收）
- 每个已加载类的 JClass 对象及其 char[] 名称数组
- 已知限制：不支持 `finalize()`；Weak/Soft/Phantom Reference 为 stretch（见文末）

## 任务清单

### 1. CLI 扩展 — `internal/cmd.go`

- `-Xmx <size>`（默认 64m）、`-Xms <size>`（仅预留语义，先记录）、`-verbose:gc`
- `ParseXss` 扩展 `g` 后缀（`-Xmx1g`）；**M01 测试不动**（新增后缀不影响既有用例），在 M15 测试补 `1g → 1073741824` 用例

### 2. 堆账本 — `internal/heap/gc.go`

```go
type gcMeta struct {
    size      int64  // 估算字节数
    kind      int    // instance / array / string / jclass
    permanent bool   // 永生对象
}
type Heap struct {
    maxBytes, usedBytes int64
    table map[*Object]*gcMeta
}
func (h *Heap) Register(obj *Object, size int64)  // 分配点调用
func (h *Heap) Collect(loader, threads, pools …) gcStats // full GC
```

- 大小估算：实例 = `槽数×8 + 16`；数组 = `元素大小×len + 16`（引用元素按 8）；够用即可，目标是确定性可测

### 3. 分配点接入（集中、少量）

所有创建 `*Object` 的路径收口到注册调用——预计 4–6 处（对象/数组/JString/JClass）。

### 4. 根集枚举（完备性是生命线）

| 根 | 来源 |
|----|------|
| 线程栈 | 所有 Thread 的每帧 LocalVars + OperandStack 的 ref 槽 |
| 类静态区 | 全部已加载 Class 的 staticVars ref 槽（含 System.in/out/err） |
| 字符串池 | StringPool 全部条目 |
| 引擎暂存 | Thread.PendingException |
| 永生对象 | JClass 及其名称数组 |

### 5. 标记阶段 — 迭代工作列表（禁递归，对象图可能深）

- 实例对象：沿**类继承链**全字段扫描，描述符为 `'L'`/`'['` 的槽是 ref（父类 ref 字段易漏，重点单测）
- 引用数组：逐元素
- JClass：永生已覆盖，不再深挖
- Class 对象 `extra`：只处理已知类型

### 6. 清除阶段

未标记 → 注册表注销 + `usedBytes -= size`；输出统计（swept、freedBytes）。

### 7. OOM

回收后 `used+need` 仍超限 → `ThrowException(loader, "java/lang/OutOfMemoryError", "Java heap space")`。

### 8. `-verbose:gc` 行格式（测试断言此格式）

```
[GC (Allocation Failure) roots=42 marked=17 swept=25 freed=120KB used=1.2M/64M pause=0.3ms]
```

### 9. `System.gc()` native — `native/java_lang_system.go`

注册 `gc()V` → 触发 full GC（无返回值）。

## 关键难点

- **根集完备性**：漏一个根 = 活对象被回收 = 悬垂指针崩溃。JClass、字符串池、PendingException 是最易漏的三处，逐项配"回收后仍存活"单测
- 实例对象 ref 槽定位依赖类字段布局（M06 的槽位计算），父类字段必须随继承链一起扫
- OOM 的可测确定性：`-Xmx` 固定 + 固定分配序列，使 OOM 时点可复现
- 分代（stretch）需要**写屏障**：`putfield/putstatic/aastore` 必须统一走 `heap.WriteRef`，M15 先留接缝不实现

## 验收标准（TDD 先行）

> 分工沿用 M01 起的约定：测试由 AI 编写，实现由用户编写。

Go 单测（`internal/heap/gc_test.go`）：

| # | 用例 | 期望 |
|---|------|------|
| 1 | 分配→断开引用→Collect | 注册表条目数下降、usedBytes 下降 |
| 2 | 栈/静态/字符串池持有的对象 Collect 后仍在 | 可达性保持 |
| 3 | 环引用（A.field=B, B.field=A）断开外部引用后 Collect | 双双回收（标记-清除正确性） |
| 4 | 父类 ref 字段指向的对象，子类实例存活时不被回收 | 继承链字段扫描正确 |
| 5 | `-Xmx` 极小 + 大数组分配 | OutOfMemoryError("Java heap space") |
| 6 | `-verbose:gc` 输出 | 行格式符合 §8 断言 |
| 7 | System.gc() | 触发一次 Collect |
| 8 | ParseXss("1g") | 1073741824 |

集成 Java 程序（`tests/java/gc/`）：

| # | 用例 | 期望 |
|---|------|------|
| 9 | `GcChurn`（循环分配丢弃 + System.gc() 计数） | 正常完成，输出 GC 次数 |
| 10 | `GcOOM`（持续 new 大数组直至捕获 OutOfMemoryError 并打印继续运行） | 与 JDK 行为一致（程序可继续） |
| 11 | `LiveList`（长存活链表 + 大量短命对象，`-Xmx4m` 跑完） | 不 OOM（证明回收有效） |
| 12 | 回归：M01–M14 全部测试程序 | 全绿 |

## 完成动作

- commit：`feat: 标记-清除垃圾回收（对象注册表/-Xmx/OOM/verbose:gc）`
- tag：`m15`
- README 进度表 + DESIGN.md 已知限制同步更新

## Stretch（可选，另立 M16）

- 分代收集：新生代复制 GC + 老年代标记-清除，写屏障 + 记忆集（依赖 M15 留下的 `WriteRef` 接缝）
- `java.lang.ref.WeakReference/SoftReference`（需要 Reference 队列语义）
- `finalize()` 支持（每对象至多一次，独立 finalizer 队列）
