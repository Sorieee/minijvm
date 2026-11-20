# M09 异常处理

## 目标

实现 Java 异常机制端到端：异常表、athrow、处理器匹配、逐帧上抛、未捕获异常打印，并把此前的「文案报错」换成真异常对象。

## 对应书章节

第 9 章「异常处理」。

## 前置依赖

M08（类加载与对象模型齐备）。

## 任务清单

### 1. 异常表解析转正 — `classfile/attribute_info.go`（M03 已读入字段，本里程碑接通语义）

```go
type ExceptionTableEntry struct {
    StartPc, EndPc, HandlerPc uint16
    CatchType                 uint16 // 常量池索引，0 = catch-all（finally）
}
```

### 2. athrow — `instructions/references/athrow.go`

- 弹栈取异常对象引用（null → NullPointerException）
- **当前帧内查表**：`FindExceptionHandler(exClass, pc)` → 命中则清空操作数栈、异常引用压栈、NextPC=handlerPc
- 未命中：把异常对象存入 `thread.PendingException`，**pop 当前帧**；解释器循环每轮开头检查 PendingException → 在新当前帧继续查表 → 仍未命中继续 pop……（上抛）
- 栈底（main 帧也未命中）：清空线程栈 → **引擎打印**：

```
java.lang.ArithmeticException: / by zero
	at com.test.Main.main(Main.java:12)
```

（LineNumberTable 定位行号；引擎级打印永久保留——未捕获异常本来就该由引擎输出）

### 3. 引擎异常对象工厂 — `heap/exception.go`

```go
func ThrowException(loader, exClassName, msg string) *Object
// newObject(异常类) → 调用 <init>(Ljava/lang/String;)V（压帧执行构造器）
// 特例：Throwable.<init> 链里的 fillInStackTrace()（JDK8 是 native）→ M09 给雏形
```

- 构造器执行依赖解释器递归——实现为「把构造帧压栈 + PendingException 流转」与普通调用同构
- **Throwable.fillInStackTrace 雏形**：注册最小 native（返回 this），正式版 M11

### 4. 运行时错误接入（替换此前的文案 panic）

| 时机 | 异常 |
|------|------|
| idiv/irem/ldiv/lrem 除零（int/long） | ArithmeticException: / by zero |
| JvmStack 超限 | StackOverflowError |
| checkcast 失败 | ClassCastException |
| athrow null | NullPointerException |
| 数组负长度/越界 | NegativeArraySizeException / ArrayIndexOutOfBoundsException（读取时校验） |

注意：抛真异常本身需要类加载+构造器执行，**在引擎自身报错路径上要防递归**（加载异常类失败 → 退回 Go 层文案打印）。

### 5. printStackTrace 桥接（书同款手法）

`invokevirtual Throwable.printStackTrace()`：拦截并用 Go 打印我们维护的帧堆栈（LineNumberTable 定位）。M11 后此桥接**保留**（正式实现走 backtrace 太深，桥接是书的等价做法）。

## 关键难点

- **异常表 catch 匹配**：异常类 isInstanceOf catchType（继承链匹配）；catchType=0 的 finally 语义
- PendingException 在 pop 帧后必须保留传递（帧清空但异常对象是堆上的，安全）
- 构造异常对象时的递归保护：ArithmeticException 构造又触发了异常 → fallback 直接 Go 打印退出
- LineNumberTable 反查：code offset → 行号（取 ≤ pc 的最大 startPc）

## 验收标准（TDD 先行）

`tests/java/exception/`：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `TryCatch`（捕获自定义异常，类型精确命中） | catch 分支输出 |
| 2 | `CatchHierarchy`（catch 父类捕获子类异常） | 命中父类 handler |
| 3 | `FinallyTest`（try/finally、try/catch/finally、return 前执行 finally） | 输出顺序与 java 一致 |
| 4 | `Rethrow`（catch 后再 throw 上抛） | 上层捕获 |
| 5 | `Propagate`（5 层调用栈深处抛出） | 顶层捕获 |
| 6 | `Uncaught`（main 直接抛） | 引擎打印 `java.lang.RuntimeException: msg` + 堆栈行号 |
| 7 | `DivZero`（int 除零） | 捕获 ArithmeticException |
| 8 | `DeepRecursion`（无限递归） | 捕获/打印 StackOverflowError |
| 9 | 异常表 FindExceptionHandler 单测 | pc 边界（startPc≤pc<endPc）正确 |
| 10 | 上述全部与 JDK21 `java` 运行输出逐行一致（脚本断言） |

## 完成动作

- commit：`feat: 异常处理（异常表/athrow/上抛/真异常接入）`
- tag：`m09`
- 更新 README 进度表
