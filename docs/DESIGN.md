# MiniJVM 总体设计

> 参照《自己动手写Java虚拟机》（张秀宏，Go 版）的技术路线。本文档回答「整体怎么设计、为什么这么设计」；分阶段任务见 [PLAN.md](PLAN.md) 与 [milestones/](milestones/)。

## 1. 项目定位

MiniJVM 是一个**学习型**的字节码执行引擎：解释执行 Java 8 字节码（类文件版本 ≤ 52），能运行真实的 Java 控制台程序。不追求性能、不追求 JVM 规范的完整实现（不做 JIT、不支持 invokedynamic）。GC 为超出书范围的扩展：自研标记-清除（M15），分代收集列为 stretch。

核心设计思想（与书一致）：**JVM 不自带类库**。`java.lang.Object`、`java.lang.String`、`java.util.ArrayList` 等直接从真实 JDK 8 的 `rt.jar` 加载，由解释器执行其中的纯 Java 字节码。这样无需手写类库，就能跑"真"Java 程序；需要解释器亲手实现的只有 native 方法（如 `Object.hashCode`、`FileOutputStream.writeBytes`）。

## 2. 关键决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 实现语言 | Go | 与书完全一致，代码清单可直接对照 |
| 类文件目标版本 | ≤ v52（Java 8） | 结构最简单；用 JDK21 的 `javac --release 8` 编译测试程序 |
| 核心类库来源 | JDK 8 的 `rt.jar` | 书的关键取舍；JDK 9+ 的模块化类库（v61+、condy、nestmates）复杂度失控 |
| 测试程序编译 | `javac --release 8` | 产 v52 字节码；字符串拼接走 StringBuilder，天然规避 invokedynamic |
| 代码组织 | 单一代码树 + 里程碑 tag（m00…m14） | 持续演进得到完整成品，符合真实工程习惯 |
| 代码结构 | 单一 `internal/` 分包 | 书按 chXX 每章复制整棵树；我们改为演进式结构，每个里程碑 = 一次可运行的提交 |

## 3. 总体架构

```
┌────────────────────────────────────────────────────────────┐
│ cmd/minijvm —— 类 java 命令行（-cp / -Xjre / -Xss / -jar）  │
├────────────────────────────────────────────────────────────┤
│ classpath     四种 Entry（目录/jar/组合/通配符）定位类字节    │
├────────────────────────────────────────────────────────────┤
│ classfile     class 文件解析（常量池/字段/方法/属性）        │
├────────────────────────────────────────────────────────────┤
│ heap          运行时类/对象模型（Class/Method/Field/Object、 │
│               符号引用解析、字符串池、数组类）                │
├────────────────────────────────────────────────────────────┤
│ rtda          线程私有数据区（Thread/Stack/Frame/            │
│               LocalVars/OperandStack，依赖 heap）            │
├────────────────────────────────────────────────────────────┤
│ instructions  ~200 条指令，按类别分包 + opcode 查表工厂       │
│               + fetch-decode-execute 解释器循环              │
├────────────────────────────────────────────────────────────┤
│ native        本地方法注册表（Object/Class/System/IO/...）    │
└────────────────────────────────────────────────────────────┘
```

依赖方向（单向，避免循环 import）：

```
cmd → instructions → rtda → heap → classfile → classpath
                     ↑ (Frame 持有 *heap.Method)
        native → rtda/heap
```

**关键点：`heap` 包自足**——`Slot` 定义在 `heap` 包（`Object.fields []Slot`），`rtda` 的 LocalVars/OperandStack 引用 `heap.Slot` 和 `*heap.Object`。书第 4 章先在 rtda 内用占位 `Object`，第 6 章引入 heap 时才重构；我们沿用同样的演进（见 M04/M06）。

## 4. 核心机制设计

### 4.1 类加载生命周期（M08 实装，M06 有雏形）

```
加载 readClass → defineClass（解析 classfile 填充 heap.Class）
→ 链接 verify（简略跳过）+ prepare（静态字段默认值 + ConstantValue）
→ 初始化 <clinit>（防重入标记；先递归初始化父类；
   触发点：new / getstatic / putstatic / invokestatic）
```

类查找：`java/**` 等系统类走 `-Xjre` 指定的 jre 路径（`jre/lib/*` → `rt.jar`），用户类走 `-cp`（M14 显式化为 BootstrapClassLoader + 委派模型）。

### 4.2 运行时数据区

- **Thread**：`pc int` + 私有操作数栈帧链 `*Frame`（链表，push/pop）
- **Frame**：`method *heap.Method`、`localVars`（按 maxLocals 分配）、`operandStack`（按 maxStack 分配）、`nextPC`
- **Slot**：4 字节槽。`long/double` 占两个槽（高位槽空置）；引用类型存 `*heap.Object`
- **JvmStack**：按 `-Xss` 限制深度，超限抛 `StackOverflowError`（M09 起为真异常）

### 4.3 解释器循环（M05 实装）

```
loop:
  pc = frame.nextPC; thread.pc = pc
  opcode = code[pc]; decoder 重定位 nextPC
  inst = factory.NewInstruction(opcode)   // switch 大表
  inst.FetchOperands(reader)
  inst.Execute(frame)                     // 可能触发类加载/方法调用/natives
  if 栈空 → 结束
```

指令按 JVM 规范的类别分包：`constants / loads / stores / stack / math / conversions / comparisons / control / references / extended / reserved`。每条指令实现 `FetchOperands(reader)` + `Execute(frame)`。

### 4.4 方法调用（M06 实装）

- `invokestatic`：解析 MethodRef → （init 触发）→ 新帧压栈
- `invokespecial`：`<init>`、private、super 调用三种情形，按 JVM 解析规则选实际方法
- `invokevirtual`：沿类继承链 + 接口查找运行时实际方法
- 符号引用惰性解析：`ClassRef/MethodRef/FieldRef` 首次使用时解析并缓存（resolved 指针）

### 4.5 异常机制（M09 实装）

- `CodeAttribute` 携带 `ExceptionTable`（catchType/startPc/endPc/handlerPc）
- `athrow`：把异常对象作为特殊返回值上抛；解释器循环内对当前帧查表（异常类 isInstanceOf catch 类则跳 handlerPc），查不到 pop 帧继续向上
- 栈顶仍未处理：清空线程栈，Go 端打印 `java.lang.XxxException: msg` + 帧堆栈（引擎职责，永久保留）
- 除零 → `ArithmeticException`、栈超限 → `StackOverflowError`、数组越界 → `ArrayIndexOutOfBoundsException`（M09/M11/M13 逐步接入）

### 4.6 本地方法（M11 实装）

- 注册表：`map[className]map["name+descriptor"]func(frame)`；invoke 前查表命中即调 Go 实现
- 最小集合以「`System.<clinit>` 实际触发链」为准逐个补齐，预期包括：
  - `Object`：registerNatives / getClass / hashCode / clone
  - `System`：registerNatives / arraycopy / currentTimeMillis / nanoTime / identityHashCode / initProperties / setIn0 / setOut0 / setErr0
  - `Float/Double`：floatToRawIntBits / intBitsToFloat / doubleToRawLongBits / longBitsToDouble
  - `Throwable`：fillInStackTrace（M09 先给雏形返回 this）
  - `FileOutputStream`：writeBytes → fd 1/2 → stdout/stderr
- 移除哨兵桥接后，`HelloWorld` 走全真实链：`System.<clinit>` → `PrintStream.println`（纯 Java）→ `FileOutputStream.writeBytes`（native）

### 4.7 字符串与数组（M07 基础，M12/M13 完善）

- 字符串：JDK8 `String` 布局（`char[] value` + `int hash`）。`ldc` 字符串常量 → 构造 String 对象 + char 数组对象；M12 加 `StringPool`（Go `map[string]*Object` + 锁）与 `String.intern` native
- 数组：`newarray/anewarray/multianewarray`（M07）；M13 数组类为合成 `heap.Class`（名字 `[I` / `[Ljava/lang/String;`，superClass = Object，实现 Cloneable/Serializable），支持数组 clone 与 `java.lang.reflect.Array` natives

### 4.8 垃圾回收（M15 实装，超出书范围）

- **对象注册表 + 标记-清除**：分配点统一注册 `*Object` 并累计 `usedBytes`；`Slot.Ref` 保持 Go 指针，指令层零改动；GC 从注册表注销死对象，最终内存释放交宿主 GC
- **根集**：全部线程帧（局部变量表+操作数栈）→ 全部已加载类静态槽 → StringPool → PendingException → 永生对象（JClass、类元数据永不回收）
- **触发**：分配路径同步触发（STW）；回收后仍超 `-Xmx` → `OutOfMemoryError`；`System.gc()` 显式触发；`-verbose:gc` 打印统计
- **stretch**：分代收集（复制 + 写屏障 + 记忆集，`putfield/putstatic/aastore` 预留 `heap.WriteRef` 接缝）、Reference 类型、finalize()

## 5. 与书的两处调整

1. **println 哨兵桥接（M06 引入，M11 移除）**：书直到第 11 章才能 println。为让 M06–M10 里程碑有可见输出，`getstatic java/lang/System.out` 特判返回一个哨兵对象，`invokevirtual println` 命中哨兵直接由 Go 输出。书的第 9 章对 `Throwable.printStackTrace` 用的就是同类桥接手法。
2. **M05 增加临时 `-Xtest` 模式**：加载类 → 调用 `static int test()` → 打印返回值。让 M05 纯计算阶段就能端到端验证（书此时只能单测/调试），M06 桥接上线后保留作调试工具，M14 后退役。

**超出书范围的扩展**：M15 自研 GC（书依赖 Go 宿主 GC，jvmgo 只把 GC 列为练习方向）。我们的方案：对象注册表 + 标记-清除，`Slot.Ref` 保持 Go 指针、指令层零改动；GC 决定对象生死并维护 `-Xmx` 账本，最终内存释放交给宿主 GC。详见 [milestones/M15-gc.md](milestones/M15-gc.md)。

其余与书一致：不做 JIT、不实现 invokedynamic。

## 6. 目录结构（目标形态）

```
minijvm/
├── cmd/minijvm/main.go          # 入口：解析参数 → 起 JVM
├── internal/
│   ├── cmd.go                   # Cmd 参数结构（M01）
│   ├── classpath/               # Entry 接口 + 4 种实现 + Classpath（M02）
│   ├── classfile/               # 大端读取、常量池、成员、属性（M03）
│   ├── heap/                    # Class/Method/Field/Object/Slot/字符串池/数组类（M06+）
│   ├── rtda/                    # Thread/Stack/Frame/LocalVars/OperandStack（M04）
│   ├── instructions/            # base/ + 11 个类别子包 + factory + interpreter（M05-07）
│   ├── native/                  # 注册表 + java/lang、java/io 等（M11+）
│   └── jvm.go                   # JVM 组装与启动（M08 起）
├── tests/
│   ├── java/                    # 每里程碑一组 .java（提交）
│   └── classes/                 # javac --release 8 输出（不提交）
├── scripts/                     # 编译测试程序、跑集成测试（Windows bash/bat）
└── docs/                        # 本文档体系
```

文件遵循 200–400 行、函数 <50 行、显式错误处理（不静默吞错）、不可变优先。

## 7. 风险与对策

| 风险 | 对策 |
|------|------|
| `System.<clinit>` 触发链深（牵出 PrintStream/FileOutputStream/sun.misc 等） | M11 前哨兵桥接规避；M11 按"实际触发链"逐个补 native，跑通即收敛 |
| 循环 import（rtda ↔ heap） | heap 包自足（Slot 定义在 heap），依赖单向 rtda → heap |
| JDK8 rt.jar 类触达未实现路径（图形/网络/反射深层） | 只跑控制台程序；反射只实现 Array 相关子集 |
| GC 根集遗漏 → 活对象误回收（悬垂崩溃） | 根集清单化（见 §4.8）；每个根配「回收后仍存活」单测；JClass/字符串池/PendingException 是最易漏的三处 |
| invokedynamic 不支持 | 测试程序 `--release 8` + 避免 lambda/方法引用 |
| Long/Double 双槽、switch 4 字节对齐等字节码细节坑 | M03/M05 用 javap 对拍 + 单测覆盖 |
| Windows 与书的差异（classpath 分隔符 `;`、路径分隔符） | 全程用 `filepath.ListSeparator`、`filepath.Join`，不硬编码 `/` `:` |
| 性能 | 非目标；解释器 opcode switch 表即可 |
