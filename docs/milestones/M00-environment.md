# M00 环境准备

## 目标

搭好 MiniJVM 全程所需的构建与验证环境，为 M01 之后的编码扫清障碍。

## 对应书章节

无（书的第 0 步准备工作）。

## 前置依赖

无。

## 任务清单

### 1. 安装 Go（≥ 1.21）

- Windows：`winget install GoLang.Go` 或官网 MSI 安装包
- 验证：`go version` 输出版本号
- 本机现状：未检测到 Go，需安装

### 2. 准备 JDK 8（类库来源，不进 PATH）

- 下载 Temurin 8（或任意 OpenJDK 8）zip 版（免安装），解压到 `E:\tools\jdk8`
- 验证文件存在：`E:\tools\jdk8\jre\lib\rt.jar`
- 注意：**不要**把 JDK8 加进 PATH / JAVA_HOME，避免与现有 JDK 21 冲突；它只被 `-Xjre` 参数引用

### 3. 确认 JDK 21 的 javac 可用（已有环境）

- `javac -version` 输出 21.x

### 4. 验证 `--release 8` 产出符合目标的字节码

- 写最小 `HelloWorld.java`（含 `System.out.println("...")`）
- `javac --release 8 -d /tmp/m00check HelloWorld.java`
- 检查两项：
  - `javap -v HelloWorld | grep major` → `major version: 52`
  - `javap -c HelloWorld` 中字符串拼接（若有）走 `StringBuilder`，**无** `invokedynamic`
- 检查产物可直接在 JDK 21 运行：`java -cp /tmp/m00check HelloWorld`

### 5. 初始化仓库结构

- `git init -b main`（如尚未初始化）
- `.gitignore` 就位
- `go mod init minijvm`
- 目录骨架：`cmd/minijvm/`、`internal/`、`tests/java/`、`scripts/`

## 关键难点

- JDK 8 与 JDK 21 共存：只解压不配置环境变量，M02 起 `-Xjre` 指向 `E:\tools\jdk8\jre\lib`
- `--release 8` 是硬要求：不带它会产 v65 字节码，且字符串拼接走 `invokedynamic`（我们不支持）

## 验收标准（TDD 先行）

| # | 验证项 | 期望结果 |
|---|--------|----------|
| 1 | `go version` | ≥ 1.21 |
| 2 | rt.jar 存在 | `E:\tools\jdk8\jre\lib\rt.jar` |
| 3 | `javap -v` major | `52` |
| 4 | 拼接反编译 | 无 `invokedynamic` |
| 5 | `go mod init` | `go.mod` 生成，`module minijvm` |

## 完成动作

- commit：`chore: 初始化仓库与环境准备`
- tag：`m00`
- 更新 README 里程碑进度表
