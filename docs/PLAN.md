# MiniJVM 总计划

> 分阶段任务清单。每个里程碑的详细任务见 [milestones/](milestones/) 下对应文件；设计依据见 [DESIGN.md](DESIGN.md)。

## 1. 目标与范围

- 用 Go 实现解释执行 Java 8 字节码（类文件 ≤ v52）的迷你 JVM
- 核心类库复用 JDK 8 `rt.jar`，覆盖《自己动手写Java虚拟机》全部 14 章
- **扩展里程碑 M15：自研 GC**（标记-清除 + `-Xmx` + OOM；书未实现 GC，分代收集为 stretch）
- 不做：JIT、invokedynamic、图形/网络类库路径

## 2. Phase 0：环境准备（M00）

| 事项 | 说明 |
|------|------|
| Go ≥ 1.21 | `winget install GoLang.Go` 或官网 MSI；`go version` 验证 |
| JDK 8 / Temurin 8 | zip 解压到 `E:\tools\jdk8`，**不进 PATH**；确认 `E:\tools\jdk8\jre\lib\rt.jar` 存在 |
| JDK 21+（已有） | 仅用 `javac --release 8` 编译测试程序 |
| 验证 | `javac --release 8` 编译 HelloWorld，`javap -v` 确认 `major version: 52`，字符串拼接反编译无 invokedynamic |

## 3. 里程碑总览

| # | 里程碑 | 书章节 | 核心内容 | 验收要点 |
|---|--------|--------|----------|----------|
| M00 | [环境准备](milestones/M00-environment.md) | - | Go / JDK8 / javac 验证 | 三条验证命令通过 |
| M01 | [命令行工具](milestones/M01-cmdline.md) | 第 1 章 | Cmd 结构、java 风格参数解析 | `-version` / `-h` / 参数回显 |
| M02 | [搜索 class 文件](milestones/M02-classpath.md) | 第 2 章 | Entry 接口 + 目录/jar/组合/通配符 | 三种来源定位类字节 |
| M03 | [解析 class 文件](milestones/M03-classfile.md) | 第 3 章 | 大端读取、常量池、成员、属性 | 单测对拍 javap |
| M04 | [运行时数据区](milestones/M04-rtda.md) | 第 4 章 | Slot/局部变量表/操作数栈/帧/线程 | 单测：栈语义、双槽、溢出 |
| M05 | [指令集和解释器](milestones/M05-instructions.md) | 第 5 章 | 指令抽象、查表工厂、解释器循环、7 类指令 | `-Xtest` 跑纯计算程序 |
| M06 | [类和对象](milestones/M06-class-object.md) | 第 6 章 | heap 模型、引用解析、三种 invoke、println 桥接 | HelloWorld + 多态程序 |
| M07 | [求值和栈帧操作](milestones/M07-evaluation.md) | 第 7 章 | new/字段读写/数组/ldc/instanceof/wide | 对象+数组程序 |
| M08 | [类加载](milestones/M08-classloader.md) | 第 8 章 | define/link/init 生命周期、`<clinit>` 防重入 | 初始化顺序程序 |
| M09 | [异常处理](milestones/M09-exception.md) | 第 9 章 | ExceptionTable、athrow、逐帧上抛 | try/catch/finally 全覆盖 |
| M10 | [打包和部署](milestones/M10-packaging.md) | 第 10 章 | `-jar`（MANIFEST）、Go 交叉编译 | 运行真实 jar |
| M11 | [本地方法调用](milestones/M11-native.md) | 第 11 章 | native 注册表、System 链、移除桥接 | HelloWorld 全真实链 |
| M12 | [字符串](milestones/M12-string.md) | 第 12 章 | StringPool、intern | 拼接/==/intern 程序 |
| M13 | [数组](milestones/M13-array.md) | 第 13 章 | 数组类、clone、reflect.Array natives | 多维数组/clone/反射程序 |
| M14 | [系统类加载器](milestones/M14-bootclasspath.md) | 第 14 章 | BootstrapClassLoader、委派模型 | ArrayList/HashMap 程序 + 全量回归 |
| M15 | [垃圾回收](milestones/M15-gc.md) | 书外扩展 | 对象注册表 + 标记-清除、`-Xmx`、OOM、`-verbose:gc`、`System.gc()` | 回收正确性/OOM/环引用用例 + 全量回归 |

每个里程碑：**讲解 → TDD 实现 → 测试通过 → commit + tag（mXX）→ README 进度更新**。

## 4. 测试策略

### 4.1 Go 单元测试（核心包覆盖 ≥ 80%）

- `classfile`：用 `javac --release 8` 编译的固定类文件字节 + `javap -v` 人工确认的期望值做断言（常量池条目、major version、方法表）
- `rtda`：操作数栈 push/pop、long/double 双槽、局部变量读写、栈深度限制
- `heap`：方法描述符解析、类引用解析
- `instructions`：手工构造 code 字节 → 跑解释器 → 断言栈顶结果（M05 起持续补充）
- `heap/gc`（M15）：回收正确性（可达性保持/环引用/继承链 ref 字段）、OOM、`-verbose:gc` 行格式

```bash
go test ./... -cover
```

### 4.2 Java 集成测试程序集

- `tests/java/` 下每里程碑一个子目录（如 `tests/java/arith/`、`tests/java/exception/`）
- 脚本统一 `javac --release 8 -d tests/classes` 编译，`minijvm -cp tests/classes <Main>` 运行，**输出与 JDK21 `java` 的运行结果逐行比对**
- 程序编写约束：不用 lambda / 方法引用 / stream（规避 invokedynamic）；只做控制台输出

### 4.3 端到端验收（M14 收尾）

HelloWorld、递归 fib、继承多态、try/catch/finally、字符串、二维数组、ArrayList/HashMap —— minijvm 输出与真 JDK 一致。

## 5. TDD 流程（每个里程碑）

1. 先写该里程碑的验收测试（解析器先写对拍用例；解释器先写 Java 程序期望输出）
2. 跑测试确认失败（RED）
3. 最小实现（GREEN）
4. 重构 + 覆盖率核对 ≥ 80%（IMPROVE）
5. 大里程碑（M03/M05/M06/M11）追加一轮代码自查
6. commit（conventional commits）+ tag + 更新 README 进度表

## 6. Git 规范

- 分支：`main` 单分支演进（学习型项目）
- 提交信息：`feat: 解析 class 文件常量池` / `fix: 修正 long 双槽偏移` / `docs: ...` / `test: ...` / `refactor: ...`
- 标签：`m00`…`m14`，每个里程碑完成打一次
- 推送远端前需明确确认（全局规则）

## 7. 协作方式

逐里程碑推进，节奏：**该章原理讲解 → 一起过关键设计 → 实现 → 跑测试 → 提交**。任一节点可暂停、提问、调整方向。

## 8. 全局风险提示

| 风险 | 对策 |
|------|------|
| `System.<clinit>` 触发链深 | M11 前哨兵桥接规避；M11 按实际触发链补 native |
| 循环 import | heap 自足、依赖单向（见 DESIGN.md §3） |
| Windows 差异（classpath `;`、路径） | `filepath.ListSeparator` 全程统一 |
| invokedynamic | `--release 8` + 避免 lambda |
| 反射/深层类库路径 | 只实现 Array 反射子集，只跑控制台程序 |
| GC 根集遗漏 → 活对象误回收 | 根集清单化（M15 文档 §4）；每个根配「回收后仍存活」单测 |
