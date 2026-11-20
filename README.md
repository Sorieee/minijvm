# MiniJVM

从零实现一个迷你 Java 虚拟机，参照张秀宏《自己动手写Java虚拟机》（Go 语言版）的技术路线。

- **语言**：Go
- **目标**：能运行真实的 Java 8 字节码（类文件版本 ≤ 52）
- **核心思想**：JVM 只做「字节码执行引擎」，`java.lang.*` 等类库不自己实现，直接从 JDK 8 的 `rt.jar` 加载真实类

## 文档导航

| 文档 | 内容 |
|------|------|
| [docs/DESIGN.md](docs/DESIGN.md) | 总体设计：架构分层、关键决策、核心机制、目录结构、风险对策 |
| [docs/PLAN.md](docs/PLAN.md) | 总计划：环境准备、里程碑总览、测试策略、TDD 流程、协作方式 |
| [docs/milestones/](docs/milestones/) | 16 个里程碑（M00–M15），每个一个文件，含分步任务清单与验收标准；M15 为书外扩展（自研 GC） |

## 里程碑进度

| 里程碑 | 名称 | 对应书章节 | 状态 |
|--------|------|-----------|------|
| M00 | 环境准备 | - | ✅ 完成（Go 1.27.4 + Temurin 8u512b09） |
| M01 | 命令行工具 | 第 1 章 | ⬜ 未开始 |
| M02 | 搜索 class 文件 | 第 2 章 | 🔨 进行中（骨架就绪，待实现） |
| M03 | 解析 class 文件 | 第 3 章 | ⬜ 未开始 |
| M04 | 运行时数据区 | 第 4 章 | ⬜ 未开始 |
| M05 | 指令集和解释器 | 第 5 章 | ⬜ 未开始 |
| M06 | 类和对象 | 第 6 章 | ⬜ 未开始 |
| M07 | 求值和栈帧操作 | 第 7 章 | ⬜ 未开始 |
| M08 | 类加载 | 第 8 章 | ⬜ 未开始 |
| M09 | 异常处理 | 第 9 章 | ⬜ 未开始 |
| M10 | 打包和部署 | 第 10 章 | ⬜ 未开始 |
| M11 | 本地方法调用 | 第 11 章 | ⬜ 未开始 |
| M12 | 字符串 | 第 12 章 | ⬜ 未开始 |
| M13 | 数组 | 第 13 章 | ⬜ 未开始 |
| M14 | 系统类加载器 | 第 14 章 | ⬜ 未开始 |
| M15 | 垃圾回收（扩展） | 书外扩展 | ⬜ 未开始 |

状态标记：⬜ 未开始 / 🔨 进行中 / ✅ 完成（完成时同步打 tag `mXX`）

## 环境要求

> 已就绪（2026-11-23，本机）：Go 1.27.4（`C:\Program Files\Go`，系统 PATH），
> Temurin 17.0.20.1（`E:\tools\jdk-17.0.20.1+1`，javac 用前先 `source /e/tools/env.sh`），
> Temurin 8u512b09（`E:\tools\jdk8`，仅作 rt.jar 类库来源，未进 PATH）。

- Go ≥ 1.21
- JDK 9+（仅用 `javac --release 8` 编译测试程序；本机用 JDK 17，已验证产出 v52）
- JDK 8 / Temurin 8（仅作 `rt.jar` 类库来源，不必进 PATH）

详见 [docs/milestones/M00-environment.md](docs/milestones/M00-environment.md)。

## 运行（规划中的最终形态）

```bash
# 编译
go build -o bin/minijvm.exe ./cmd/minijvm

# 运行 Java 程序（classpath 用 ; 分隔，Windows）
bin/minijvm.exe -Xjre E:\tools\jdk8\jre\lib -cp tests\classes HelloWorld

# 运行 jar
bin/minijvm.exe -Xjre ... -jar app.jar
```
