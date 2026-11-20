# M04 运行时数据区

## 目标

实现线程私有的运行时数据结构：槽、局部变量表、操作数栈、帧、栈、线程。堆（heap 包）尚未出现，用占位对象。

## 对应书章节

第 4 章「运行时数据区」。

## 前置依赖

M03（能解析出 Code 属性的 maxStack/maxLocals）。

## 任务清单

### 1. 槽 — `internal/rtda/slot.go`

```go
type Slot struct {
    Num int32
    Ref *Object // M04 阶段为 rtda 包内占位 type Object struct{}；
                  // M06 引入 heap 包后重构为 *heap.Object（见 DESIGN.md §3）
}
```

### 2. 局部变量表 — `internal/rtda/local_vars.go`

按 `maxLocals` 分配的 `[]Slot`，方法族：

```
SetInt/GetInt(idx)          SetLong/GetLong(idx)      // long 占 idx、idx+1 两槽
SetFloat/GetFloat(idx)      SetDouble/GetDouble(idx)
SetRef/GetRef(idx)
```

越界访问必须报错（不静默）。

### 3. 操作数栈 — `internal/rtda/operand_stack.go`

按 `maxStack` 分配 `[]Slot` + `size` 指针：

```
PushInt/PopInt ... PushLong/PopLong（占两槽）... PushRef/PopRef
PushSlot/PopSlot（M06 用于传参复制）
Clear()（调用返回时复用帧）
```

### 4. 栈帧 — `internal/rtda/frame.go`

```go
type Frame struct {
    lower        *Frame   // 链表
    LocalVars    *LocalVars
    OperandStack *OperandStack
    Thread       *Thread
    NextPC       int
    // M06 加 Method *heap.Method
}
```

### 5. JVM 栈 — `internal/rtda/jvm_stack.go`

链式 push/pop；构造时传 maxSize（来自 `-Xss`），超过限制 `panic("java.lang.StackOverflowError")`（M09 改为抛真异常对象）。 pop 空栈返回 nil（解释器以栈空为结束条件）。

### 6. 线程 — `internal/rtda/thread.go`

```go
type Thread struct {
    pc    int
    stack *JvmStack
}
func NewThread(xss int) *Thread
func (self *Thread) PC() int / SetPC(int)
func (self *Thread) PushFrame / PopFrame / CurrentFrame / TopFrame
```

## 关键难点

- **long/double 双槽**：PushLong 时 `size += 2`，第二槽留空；PopLong 读两槽。Get/Set 局部变量同理。这是后续所有算术指令正确性的地基，必须用单测钉死
- 栈帧链表方向（lower 指向下层帧）与 pop 空判定
- `-Xss` 帧数上限：M04 用「帧数」近似（书的做法），不做字节级计量

## 验收标准（TDD 先行）

`internal/rtda/*_test.go`：

| # | 用例 | 期望 |
|---|------|------|
| 1 | LocalVars int 写读 | 值一致 |
| 2 | LocalVars long 双槽写读，且相邻槽不受污染 | 值一致 |
| 3 | LocalVars double/ref 写读 | 值一致 |
| 4 | 越界索引 | 返回错误/panic（定一种，一致贯彻） |
| 5 | OperandStack 100 个 int 进出 | LIFO 顺序正确 |
| 6 | PushLong 后 PushInt | PopInt 先出、PopLong 后出且值正确 |
| 7 | JvmStack 超过 maxSize push | panic StackOverflowError 文案 |
| 8 | Thread 帧链 push/pop/current | 链序正确 |
| 9 | rtda 包覆盖率 | ≥ 80% |

## 完成动作

- commit：`feat: 运行时数据区（槽/局部变量表/操作数栈/帧/线程）`
- tag：`m04`
- 更新 README 进度表
