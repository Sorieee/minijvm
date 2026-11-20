# tests/java —— Java 测试程序集

MiniJVM 的集成测试程序源码（编译产物 `tests/classes/` 不入库）。

## 组织方式

每个里程碑一个子目录，目录内每个可执行类都包含 `public static void main(String[] args)`：

| 目录 | 里程碑 | 内容 |
|------|--------|------|
| `meta/` | M03 | 结构对拍 javap 用的固定类（常量池/属性覆盖面广） |
| `arith/` | M05 | 纯计算（含 `static int test()`，供 `-Xtest` 调试入口） |
| `basic/` | M06 | HelloWorld、静态/虚方法、继承多态 |
| `object/` | M07 | 对象/字段/数组/类型检查 |
| `clinit/` | M08 | 类初始化顺序 |
| `exception/` | M09 | try/catch/finally、运行时异常 |
| `native/` | M11 | println 全链路、arraycopy、hashCode |
| `string/` | M12 | intern、拼接、常用方法 |
| `array/` | M13 | 数组 clone、reflect.Array |
| `final/` | M14 | ArrayList/HashMap 等复杂程序 |

## 编写约束（重要）

- 一律 `javac --release 8` 编译（保证 v52 字节码、字符串拼接走 StringBuilder）
- **禁止 lambda / 方法引用 / stream**（解释器不实现 invokedynamic）
- 只做控制台输出；每类顶部注释写明期望输出，便于与真 JDK 比对

## 编译

```bash
./scripts/build-tests.sh     # 输出到 tests/classes/
```
