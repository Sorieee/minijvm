# M08 类加载

## 目标

把 M06 的「简单加载器雏形」升级为正式的类加载生命周期：加载 → 链接（准备/校验）→ 初始化，含 `<clinit>` 防重入与触发时机。

## 对应书章节

第 8 章「类加载」。

## 前置依赖

M07。

## 任务清单

### 1. ClassLoader — `internal/heap/class_loader.go`

```go
type ClassLoader struct {
    cp       *classpath.Classpath
    verbose  bool          // -verbose:class
    classMap map[string]*Class // 已加载缓存
    bootLoader *ClassLoader   // M14 拆分双加载器，M08 先单一实例 + 查找顺序约定
}
func NewClassLoader(cp, verbose) *ClassLoader
func (l *ClassLoader) LoadClass(name string) *Class
func (l *ClassLoader) LoadArrayClass(name string) *Class // [I / [Ljava/lang/String;
```

`LoadClass` 步骤：

1. 缓存命中直接返回
2. 数组类名（`[` 开头）→ 合成数组类（M07 雏形转正：super=Object、实现 `[Ljava/lang/Cloneable, Serializable` 接口名、className 等元数据）
3. `readClass`：查找顺序 **`java/**` 等系统类 → boot 段（jre 路径，含 rt.jar）；其余 → ext → user（-cp）**（M14 显式化为 BootstrapClassLoader + 委派）
4. `defineClass`：classfile.ClassFile → heap.Class 转换（填充方法/字段/父类/接口引用——**父类、接口递归 LoadClass**；循环继承防护）
5. `link`（M08 简化）：verify 跳过（学习型实现，书同款）+ prepare
6. `init`（见下）

### 2. prepare（链接）

- 实例字段：全部赋零值（引用 null、数值 0、boolean false、char '\0'）
- 静态字段：赋零值 + **`ConstantValue` 属性**（static final）直接赋常量
- 类对象 `java/lang/Class` 关联：`Class.JClass` 惰性创建（`Object{class=java/lang/Class, extra=*heap.Class}`，M11 的 `Object.getClass` native 复用）

### 3. init（初始化）— `heap/class_init.go`

```go
func (c *Class) Init() {
    if !c.initStarted.compareAndSet(false, true) { return } // 防重入
    if c.SuperClass() != nil { c.SuperClass().Init() }       // 父类先初始化
    if clinit := c.GetClinitMethod(); clinit != nil {
        // 在当前线程压帧执行 <clinit>（由调用点驱动，见下）
    }
}
```

- `initStarted` 用 Go 原子布尔（单线程执行引擎，简单即可）
- **触发时机接入**（M08 起强制）：`new` / `getstatic` / `putstatic` / `invokestatic` 在解析出目标类后先 `Init()`——但**哨兵桥接路径除外**（System.out 的 getstatic 特判继续短路）
- `<clinit>` 执行方式：Init() 把方法帧压入当前线程（解释器自然驱动），或解释器循环特判——选「压帧」方案，与普通调用同构

### 4. main 接入 — `internal/jvm.go`

正式组装 JVM：`Cmd → Classpath → ClassLoader → 加载主类 → 找 main([Ljava/lang/String;)V → 压帧 → Interpreter.Loop`。传参：String 数组构造（复用 heap.JString）。`-verbose:class` 打印 `[Loaded <name> from <entry>]`。

## 关键难点

- **触发时机正确性**：getstatic 读 final 常量（ConstantValue）**不**触发初始化（JVM 规范）；本实现为简化可不区分（书同款），文档记录该偏差即可
- 类初始化顺序测试：父 <clinit> → 子 <clinit>；同类只一次
- `java/lang/String` 等系统类在主类之前就被牵出（ldc/哨兵），验证 boot 段查找覆盖所有 jvm 内部触发点
- 循环依赖防护：defineClass 先注册到 classMap 再解析父类引用，防 `class A extends B` 与 `B extends A` 死循环

## 验收标准（TDD 先行）

`tests/java/clinit/`：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `InitOrder`（父/子 <clinit> 各 println 一句） | 输出顺序：父先子后 |
| 2 | `InitOnce`（两次触发静态访问） | <clinit> 只输出一次 |
| 3 | `ConstantValue`（static final int/long/String） | 值正确 |
| 4 | `StaticDefault`（未显式赋值的 static int/long/ref） | 零值/null |
| 5 | `MainArgs`（main 接收 args 并回显） | 参数透传正确 |
| 6 | `-verbose:class` | 打印加载清单，含 java/lang/Object 来自 rt.jar |
| 7 | 循环继承（手工构造坏类文件） | 明确报错不死循环 |
| 8 | heap 包覆盖率 | ≥ 80% |

## 完成动作

- commit：`feat: 类加载生命周期（define/link/init）与 JVM 组装`
- tag：`m08`
- 更新 README 进度表
