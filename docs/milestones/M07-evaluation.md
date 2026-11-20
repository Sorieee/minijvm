# M07 求值和栈帧操作

## 目标

补全引用类指令（对象/数组/字段）与常量加载指令，使对象创建、数组、字段读写、类型检查全面可用。

## 对应书章节

第 7 章「求值和栈帧操作」。

## 前置依赖

M06（heap 模型与解析）。

## 任务清单

### 1. 对象创建 — `instructions/references/new_.go`

- `new(0xbb)`：解析 ClassRef → NewObject 压栈（**不调用构造器**，构造器由程序显式 invokespecial）
- 同名包名规避（Go 文件 `new_.go`）

### 2. 数组创建（基础版，M13 完善）

- `newarray(0xbc)`：8 种基本类型（atag 4..11）→ `heap.Array`（`Object{class=合成数组类雏形, data []int32/[]int64/...}`）——M07 用统一 `Array{Data interface{}, Len int}` 抽象即可
- `anewarray(0xbd)`：引用数组 `[]*Object`
- `multianewarray(0xc5)`：维度描述符递归构建（内层 `null`，逐层 new），count 校验负数报错
- `arraylength(0xbe)`
- 数组类合成：按描述符 `[I`/`[D`/`[Ljava/lang/String;` 生成 Class（M13 完整化：继承 Object、Cloneable/Serializable）

### 3. 字段访问

- `getfield(0xb4)` / `putfield(0xb5)`：receiver + FieldRef 解析（含继承链）→ 按 descriptor 读写实例字段槽
- `getstatic(0xb2)` / `putstatic(0xb3)`：读写 Class 的 staticVars；**补齐 long/double 支持**（M06 只做了 int/ref）
- 静态字段的解析同样走继承链（父类静态字段）

### 4. 类型检查

- `instanceof(0xc1)`：null → 0；沿类/接口/数组维度的 isInstanceOf 规则（heap.IsInstanceOf：类继承、接口实现、数组协变、Object 兜底）
- `checkcast(0xc0)`：失败抛 `ClassCastException` 文案错误（M09 换真异常）

### 5. 常量加载 — `instructions/constants/ldc.go`

- `ldc(0x12)` / `ldc_w(0x13)` / `ldc2_w(0x14)`
- Integer/Float → 直接入栈；**Long/Double（仅 ldc2_w）**→ 双槽入栈
- String → 构造字符串对象（复用 M06 桥接中的构造逻辑，抽到 heap：`heap.JString(class, goStr)`）
- Class 常量（ldc `Foo.class`）→ 返回该类的 Class 对象雏形（`Object{class=java/lang/Class, extra=*heap.Class}`，M11 native getClass 同构）

### 6. extended 包（增强指令）

- `wide(0xc4)`：iload/istore/... + iinc 的 16 位索引前缀展开
- `iinc(0x84)`：局部变量 += 常量（在 M05 应已实现，若未放本里程碑）
- `ifnull(0xc6)` / `ifnonnull(0xc7)` / `goto_w(0xc8)` / `jsr`/`ret` 不实现（v52 编译器不产出）

## 关键难点

- **字段槽定位**：实例字段偏移 = 父类字段数 + 本类序号；Class 转换时必须自顶向下累加计算（书同款：`NewObject` 时 fields 按继承链总长分配）
- multianewarray 维度描述符解析（`[[[I` 逐维取 count，剩余维度置 null）
- ldc String 的构造链会触发 `java/lang/String` 类加载——确保类加载器雏形对 boot 段生效（M06 已就位）
- instanceof 数组规则：`Object[] o = new String[1]` 合法、`int[]` 不是 `Object[]`——按 JVM 规范表实现

## 验收标准（TDD 先行）

`tests/java/object/`：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `FieldRW`（实例字段 int/long/double/String 读写） | 输出与 java 一致 |
| 2 | `StaticField`（跨类静态字段读写 + long） | 一致 |
| 3 | `IntArray`（newarray + 求和/最值） | 一致 |
| 4 | `StrArray`（anewarray + 遍历） | 一致 |
| 5 | `MultiArray`（int[3][4] 与不规则维度） | 一致 |
| 6 | `InstanceOf`（继承/接口/数组协变矩阵） | 一致 |
| 7 | `ClassCast`（错误转型 → 文案报错） | 中断输出 ClassCastException 文案 |
| 8 | `LongConst`（ldc2_w + 运算） | 一致 |
| 9 | heap.Object 字段偏移单测（三层继承混合 long） | 槽位正确 |
| 10 | IsInstanceOf 规则表单测 | 全矩阵通过 |

## 完成动作

- commit：`feat: 对象创建/字段访问/数组/ldc/类型检查指令`
- tag：`m07`
- 更新 README 进度表
