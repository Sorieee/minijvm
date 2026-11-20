# M05 指令集和解释器

## 目标

建立指令抽象与解释器循环，实现第一批指令（约 90 条：常量/加载/存储/栈操作/算术/转换/比较分支/跳转），能解释执行**纯计算**程序。

## 对应书章节

第 5 章「指令集和解释器」。

## 前置依赖

M03（Code 属性）、M04（帧结构）。

## 任务清单

### 1. 指令抽象 — `internal/instructions/base/`

```go
type BytecodeReader struct { Code []byte; PC int }
// ReadInt8/Uint8/Int16/Uint16/Int32/Int32s(n)/Int64

type Instruction interface {
    FetchOperands(reader *base.BytecodeReader)
    Execute(frame *rtda.Frame)
}

type NoOperandsInstruction struct{}       // 大多数指令无操作数
type Index8Instruction struct{ Index uint } // 1 字节索引（局部变量表）
type Index16Instruction struct{ Index uint } // 2 字节索引（ldc_w 等）
type BranchInstruction struct{ Offset int } // 分支：NextPC = 当前指令 PC + Offset
// 各自提供 FetchOperands 默认实现 + 嵌入复用
```

`Execute` 修改帧：`frame.NextPC = ...`（分支指令）或操作栈/局部变量表。

### 2. 指令类别子包（每包一文件族，每个类别一个文件或几个文件）

| 包 | 指令（M05 范围） |
|----|------------------|
| `constants` | nop(0x00)、aconst_null、iconst_m1..iconst_5、lconst_0/1、fconst_0/1/2、dconst_0/1、bipush、sipush |
| `loads` | iload(0x15)…aload(0x19) + `_<n>` 短格式全系列（iload_0..3、lload_0..3、fload、dload、aload 各 0..3） |
| `stores` | istore(0x36)…astore(0x4a) + `_0.._3` 短格式全系列 |
| `stack` | pop/pop2、dup/dup_x1/dup_x2/dup2/dup2_x1/dup2_x2、swap |
| `math` | i/l/f/d × add/sub/mul/div/rem/neg + 移位（shl/shr/ushr）+ 位运算（and/or/xor） |
| `conversions` | i2b/i2c/i2s/i2l/i2f/i2d、l2i/l2f/l2d、f2i/f2l/f2d、d2i/d2l/d2f |
| `comparisons` | lcmp、fcmpl/fcmpg、dcmpl/dcmpg、if_icmp<eq,ne,lt,le,gt,ge>、if<eq,ne,lt,le,gt,ge>（含引用判等 if_acmpeq/acmpne 在 references 补） |
| `control` | goto、tableswitch、lookupswitch（4 字节对齐 padding！） |

### 3. opcode 查表工厂 — `internal/instructions/factory.go`

```go
func NewInstruction(opcode byte) Instruction // 巨型 switch，每条指令注明助记符注释
```

未实现 opcode：返回带助记符的 `UnimplementedError`，解释器中断并输出「opcode 0x?? (mnemonic) not implemented」——**禁止静默跳过**。

### 4. 解释器循环 — `internal/instructions/interpreter.go`

```
Loop(thread, firstFrame):
  for thread 栈非空:
    frame = thread.CurrentFrame
    pc = frame.NextPC; thread.pc = pc
    code = frame 方法字节码
    inst = NewInstruction(code[pc]); reader 定位 pc+1
    inst.FetchOperands(reader); frame.NextPC = reader.PC
    inst.Execute(frame)
    // M05 阶段：ireturn 等返回指令把栈弹空 → 结束
```

除零（idiv/rem）M05 先返回明确错误（`ArithmeticException` 文案），M09 换真异常对象。

### 5. 临时调试入口 `-Xtest <class>`（本里程碑新增，M14 退役）

- 加载 class（复用 M02 classpath + M03 解析），找 `static int test()` 方法
- 建 Thread/Frame，解释执行，**打印返回的 int**
- 这是 M05–M07 阶段唯一的端到端可见出口（真 println 桥接在 M06，全链路在 M11）

### 6. math 细节规范

- int 运算按 int32 环绕、long 按 int64；除法向零截断（Go `/` 天然一致）
- `ishr/iushr`：无符号右移需 `uint32` 位移掩码 `& 0x1f`（long `& 0x3f`）
- 浮点 `NaN` 比较：fcmpl/fcmpg 的 -1/+1 区分必须实现

## 关键难点

- **tableswitch/lookupswitch 的 4 字节对齐**：操作数从 `(pc+1)` 向上对齐到 4 的倍数处开始读，default/low/high/匹配表都要按此偏移计算
- NextPC 管理：所有指令执行前 reader 已把 NextPC 推到下一指令，分支指令再覆盖
- dup_x1 等「插入式」复制指令容易写反，逐一配单测

## 验收标准（TDD 先行）

1. 单测：手工构造 code 字节 + Frame，跑 Loop 后断言栈/局部变量
2. Java 侧：`tests/java/arith/`（Sum1to100、FibRecursion、ArithMix、SwitchTest），每个类含 `public static int test()`

| # | 用例 | 期望 |
|---|------|------|
| 1 | 手写字节 `iconst_1 iconst_2 iadd ireturn` | 返回 3 |
| 2 | long 加法/位移（lconst 系 + lushr） | 值正确（含负数） |
| 3 | `-Xtest Sum1to100` | 5050 |
| 4 | `-Xtest FibRecursion`（n=20） | 6765 |
| 5 | `-Xtest SwitchTest` | 两分支 switch（table/lookup）都正确 |
| 6 | 除零 | 报 ArithmeticException 文案，非 panic |
| 7 | 遇到未实现 opcode | 中断并打印助记符 |
| 8 | instructions 核心包覆盖率 | ≥ 80% |

## 完成动作

- commit：`feat: 指令集与解释器（常量/加载/存储/栈/算术/转换/比较/控制）`
- tag：`m05`
- 更新 README 进度表
