# M11 本地方法调用

## 目标

建立 native 方法注册表，实现第一批本地方法，移除 System.out 哨兵桥接——让 HelloWorld 走**全真实执行链**（`System.<clinit>` → PrintStream → FileOutputStream.writeBytes）。全书风险最高的里程碑，按「实际触发链」驱动实现。

## 对应书章节

第 11 章「本地方法调用」。

## 前置依赖

M10。

## 任务清单

### 1. Native 注册表 — `internal/native/`

```go
type NativeMethod func(frame *rtda.Frame)
var registry = map[string]map[string]NativeMethod{} // className → "name+descriptor" → fn

func Register(className, methodDesc string, fn NativeMethod)
func Lookup(className, method, descriptor string) NativeMethod
// package init() 里按类注册；解释器 invoke 前先查表（命中则不压新帧，直接调 Go 实现）
```

命名建议：`native/java_lang_object.go`、`java_lang_class.go`、`java_lang_system.go`、`java_lang_float_double.go`、`java_lang_throwable.go`、`java_io.go`。

### 2. invoke 接入

- `invokestatic` / `invokevirtual` / `invokespecial` 执行前查注册表；命中：参数已在栈上，按描述符弹参调 fn，fn 用 `frame.OperandStack` 放返回值
- 未命中 → 走解释执行；类/方法都解析不到且方法 ACC_NATIVE → 明确报错「native method not implemented: X.y(desc)」

### 3. 实现第一批本地方法

**java/lang/Object**：
- `registerNatives()V` → 空实现
- `getClass()Ljava/lang/Class;` → 返回 receiver 类的 JClass（M08 已有）
- `hashCode()I` → 身份哈希（对象指针/自增 id 映射表）
- `clone()Ljava/lang/Object;` → 浅拷贝（非 Cloneable 抛 CloneNotSupportedException；数组 clone M13 扩展）
- `toString` 不需要（Object.toString 是 Java 代码）

**java/lang/Class**（按测试触发链按需）：`getSuperclass`、`getInterfaces0`、`getName`、`isInterface`、`isPrimitive`、`desiredAssertionStatus`、`registerNatives`、`getComponentType` 等——先空缺，报错驱动补齐

**java/lang/System**：
- `registerNatives()V` → 空
- `arraycopy(Ljava/lang/Object;ILjava/lang/Object;II)V` → 同类型数组拷贝 + 边界检查（ArrayIndexOutOfBoundsException / ArrayStoreException）
- `currentTimeMillis()J` / `nanoTime()J` → `time.Now()`
- `identityHashCode(Ljava/lang/Object;)I` → 同 Object.hashCode
- `initProperties(Ljava/util/Properties;)Ljava/util/Properties;` → 塞入最少必需键（`file.separator`、`path.separator`、`line.separator`、`java.version`、`user.dir`、`os.name` 等；Properties 是 Hashtable 子类，纯 Java 可直接调用其 put）
- `setIn0/setOut0/setErr0` → 直接设置 System 类静态字段（这三个正是 <clinit> 建 PrintStream 的落点）

**java/lang/Float / java/lang/Double**：
- `floatToRawIntBits(F)I` / `intBitsToFloat(I)F`、`doubleToRawLongBits(D)J` / `longBitsToDouble(J)D` → `math.Float32bits` 等

**java/lang/Throwable**：
- `fillInStackTrace()Ljava/lang/Throwable;` → 正式实现：把当前线程帧快照（类名+方法名+行号）存入异常对象 `extra`；printStackTrace 桥接改读该快照

**java/io/FileOutputStream**：
- `writeBytes([BII)V` → 读 receiver 的 `fd` 字段：1 → stdout、2 → stderr，其余报 IOException
- `write(I)V`（单字节）同理
- （FileDescriptor.out/err 是 JDK8 纯 Java 构造 fd=1/2，无需 native）

### 4. 移除哨兵桥接

- 删除 `getstatic System.out` 特判与 println 哨兵特判
- 此时 `System.out.println` 真链路：`System.<clinit>`（触发链见下）→ `setOut(new PrintStream(...))` → `println`（PrintStream/Writer 纯 Java）→ `FileOutputStream.writeBytes`（native）
- M06 引入的 `-Xtest` 保留；字符串哨兵相关代码一并清理

### 5. `System.<clinit>` 触发链收敛（核心攻坚点）

预期会牵出（逐一补到能跑通为止，报错驱动）：
`System` → `sun/misc/VM`?（JDK8 `System.<clinit>` 可能引用）→ `String`/`Integer`（纯 Java）→ `PrintStream` → `FilterOutputStream` → `BufferedOutputStream`/`OutputStreamWriter`/`StreamEncoder` → `FileOutputStream` → `FileDescriptor` → `Unsafe`?（StreamEncoder 可能用 `sun/misc/Unsafe.compareAndSwapObject`）→ 按报错逐个注册（Unsafe 常用的几个 CAS/put 方法给简化实现）。

**纪律：每补一个 native 都要有触发它的失败输出记录，不提前猜实现。**

## 关键难点

- <clinit> 深链的不可预测性——用「跑 → 报 native 缺失 → 补 → 再跑」循环收敛，预计 10–20 轮
- `initProperties` 需要 Properties 对象可构造（Hashtable.<init> 纯 Java 链）
- identityHashCode 需要稳定映射（map[*Object]int + 自增，注意 Go map 键的指针等价性）
- printStackTrace 快照时机：fillInStackTrace 在 Throwable.<init> 内调用（比我们的 pendingException 流转更早），实现为注册表 native 而非引擎特判

## 验收标准（TDD 先行）

`tests/java/native/`：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `HelloWorld`（无任何桥接） | 输出正确——**全链路验收** |
| 2 | `PrintAll`（println 打印 int/long/boolean/char/String/Object/null） | 与 java 一致 |
| 3 | `ArrayCopy`（int[]/String[]/越界报错） | 一致 |
| 4 | `TimeMillis`（前后两次 currentTimeMillis 单调递增） | 通过 |
| 5 | `HashCode`（同一对象两次 hashCode 相同；不同对象不同） | 通过 |
| 6 | `ExceptionTrace`（printStackTrace 输出含类/方法/行号） | 格式一致 |
| 7 | `FloatBits`（Float.floatToIntBits 回转） | 一致 |
| 8 | `-verbose:natives`（可选）：打印已注册 native 清单 | 注册表非空 |
| 9 | 全量回归 M01–M10 测试程序 | 通过（哨兵移除后仍全绿） |

## 完成动作

- commit 1：`feat: native 方法注册表框架`
- commit 2：`feat: Object/Class/System/IO 等本地方法，移除 println 桥接`
- tag：`m11`
- 更新 README 进度表
