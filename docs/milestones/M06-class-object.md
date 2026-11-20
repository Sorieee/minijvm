# M06 类和对象

## 目标

引入堆区运行时模型（Class/Method/Field/Object），实现符号引用惰性解析与三种方法调用指令，跑通第一个完整 Java 程序（HelloWorld）。

## 对应书章节

第 6 章「类和对象」。

## 前置依赖

M05（解释器循环）。

## 任务清单

### 1. heap 包建立 + Slot 重构（对 M04 的小重构，单独 commit）

- 新建 `internal/heap/`：**heap 包自足**（定义 `Slot`、`Object`），依赖方向单向 `rtda → heap`（见 DESIGN.md §3）
- `rtda.Slot` 迁入 heap；rtda 的 LocalVars/OperandStack/Frame 改用 `heap.Slot` / `*heap.Object`
- Frame 增加 `Method *heap.Method`

### 2. 运行时类模型

| 文件 | 内容 |
|------|------|
| `heap/class.go` | `Class{cp, accessFlags, name, super *Class, interfaces, fields, methods, loader, initStarted}`；由 classfile.ClassFile 转换（NewClass） |
| `heap/method.go` | `Method{Class, maxStack, maxLocals, code, name, descriptor, isStatic/... }`；`IsPublic` 等 flag 判断 |
| `heap/field.go` | `Field` + `IsLong/IsDouble`（描述符判断） |
| `heap/method_descriptor.go` | 解析 `(II)Ljava/lang/String;` → 参数类型列表 + 返回类型（write descriptor parser，纯字符串状态机） |
| `heap/object.go` | `Object{class *Class, fields []Slot}`；`NewObject`、字段读写按 descriptor 定槽（非 long/double 占 1 槽） |

### 3. 符号引用与解析 — `heap/class_ref.go`、`method_ref.go`、`field_ref.go`

```go
type ConstantPool 内提取：
type ClassRef { SymRef }        // SymRef{cp, className} → Resolved() *Class（惰性）
type MethodRef { SymRef; name, descriptor; resolved *Method }
type FieldRef  { SymRef; name, descriptor; resolved *Field }
```

- 解析顺序：ClassRef → 加载类（M06 用「简单加载器」：`java/**` 走 `-Xjre` boot 段，其余走 `-cp`，复用 M02 Classpath；完整生命周期 M08 再做）
- `LookupMethod`：类继承链向上 + 接口；`LookupField`：类继承链
- 方法解析规则（invokevirtual 语义）：先本类，再父类递归，最后接口

### 4. 方法调用指令 — `instructions/references/`

- `invokestatic(0xb8)`：解析 MethodRef → 参数从调用方栈弹出按描述符逆序装入新帧 → 新帧压栈（解释器循环以帧链驱动）
- `invokespecial(0xb7)`：`<init>` / private / super 调用解析规则（当前类、父类、接口默认方法三段判断的简化版，书同款）
- `invokevirtual(0xb6)`：从栈取 receiver → 按实际类解析 → 传参 → 新帧
- 参数传递复用：`CopyOperandStack`（逆序弹出到新帧局部变量表 0..n）

### 5. println 哨兵桥接（M06 新增，M11 移除）

- `getstatic` 特判：常量池解析出 `java/lang/System.out` → **不加载 System**，直接返回哨兵对象（`heap.Object` 标记 `extra = SENTINEL_OUT`）
- `invokevirtual println` 特判：receiver 是哨兵 → 按描述符 `(I)V/(J)V/(Z)V/(C)V/(Ljava/lang/String;)V/(Ljava/lang/Object;)V` 从栈取参，Go 端格式化输出（字符串对象读 char[] 还原；Object 打印 className@伪 hash）
- `ldc` 字符串常量同步就位（本里程碑的最小实现：创建 String 对象 + char[] value + hash=0，intern 在 M12）

### 6. getstatic/putstatic 最小实现

M06 只支持读写 `int`/引用型静态字段（静态字段槽在 Class 上：`staticVars`）；`long/double` 与 init 触发在 M07/M08 补齐。

## 关键难点

- **invokespecial 解析规则**：`<clinit>/<init>` 看实际类；private 看当前类；super 调用（ACC_SUPER）从父类起查
- **System.out 哨兵的短路时机**：必须在「解析 ClassRef → 可能触发加载/初始化」**之前**判定，否则会提前踩进 System.<clinit> 深链
- 循环 import 防线：rtda → heap 单向；instructions → rtda/heap
- String 对象构造依赖 `java/lang/String` 类的加载与字段布局——它来自 rt.jar（boot 段），这是 M06 起就在跑真类库的第一个验证点

## 验收标准（TDD 先行）

`tests/java/basic/`（每个类 public static void main(String[])）：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `HelloWorld`（ldc + println） | 输出 `Hello, MiniJVM!` |
| 2 | `StaticCall`（跨类静态方法） | 输出计算结果 |
| 3 | `Poly`（父类引用/子类实现 invokevirtual） | 多态输出 |
| 4 | `Ctor`（new + <init> 链 + this） | 构造顺序输出 |
| 5 | `StringBasic`（String.length/equals/charAt——rt.jar 纯 Java） | 与真 java 一致 |
| 6 | method_descriptor 单测 | 8 组描述符解析对拍 |
| 7 | LookupMethod/Field 单测（手工构建继承链） | 解析命中正确 |
| 8 | 集成：`minijvm -Xjre ... -cp tests/classes Poly` | 输出与 `java Poly` 一致 |

## 完成动作

- commit 1：`refactor: Slot 迁入 heap 包，依赖单向化`
- commit 2：`feat: 类/对象/字段/方法模型与三种方法调用`
- tag：`m06`
- 更新 README 进度表
