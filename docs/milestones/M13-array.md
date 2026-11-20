# M13 数组

## 目标

数组体系完整化：合成数组类转正、数组 clone、`java.lang.reflect.Array` natives、数组相关异常全接真异常。

## 对应书章节

第 13 章「数组」。

## 前置依赖

M12。

## 任务清单

### 1. 数组类转正 — `heap/array_class.go`（M08 雏形完善）

- 数组类元数据：className（`[I`、`[Ljava/lang/String;`、`[[J`）、superClass = `java/lang/Object`、实现 `java/lang/Cloneable` + `java/io/Serializable` 接口名、accessFlags = ACC_PUBLIC|ACC_FINAL|ACC_ABSTRACT
- 数组对象数据抽象：`Object{data}`，`data` 为 `[]int32 / []int64 / []float32 / []float64 / []int16(或int8)/ []uint16(char) / []*Object`，统一 `heap.Array` 访问器：`Len() int`、`GetInt/SetInt(i)`、`GetRef/SetRef(i)`……
- 越界访问统一抛 `ArrayIndexOutOfBoundsException`（M09 已接，回归确认全覆盖）

### 2. 数组 clone

- `Object.clone()` native 扩展：receiver 是数组 → 数据深浅拷贝（一维引用数组浅拷贝元素指针，与 JVM 一致）
- 类型保留：clone 出的对象 class 与原对象一致

### 3. reflect natives — `native/java_lang_reflect_array.go`

- `Array.getLength(Ljava/lang/Object;)I` → 数组长度（非数组抛 IllegalArgumentException）
- `Array.get(Ljava/lang/Object;I)Ljava/lang/Object;` → 基本类型读出要**装箱**（int → Integer 对象，走 `Integer.valueOf` 解释调用或直接构造+缓存注意：装箱必须走 Java 的 `Integer.valueOf` 以对齐缓存语义）
- `Array.set(Ljava/lang/Object;ILjava/lang/Object;)V` → 拆箱写入 + 类型检查（ArrayStoreException）
- `Array.newInstance(Ljava/lang/Class;I)Ljava/lang/Object;` → 按类合成数组类并创建

### 4. System.arraycopy 增强（M11 基础版完善）

- 同数组自我拷贝重叠区间正确性（Go copy 已保证语义方向，验证即可）
- 引用数组跨类型赋值检查 → `ArrayStoreException`
- null 输入 → NullPointerException

### 5. Objects/Arrays 常用工具回归确认（纯 Java，能跑即行）

- `java.util.Arrays.toString/sort(int[])`、`Objects.equals/hash`——不需要 native，作为集成测试对象

## 关键难点

- **装箱路径**：reflect.Array.get 返回 Integer/Long 等包装对象，必须复用 rt.jar 的 `Integer.valueOf`（含 -128..127 缓存）——实现为「构造 Integer 对象 + 压帧调 valueOf」，或在验证无缓存依赖场景时直接 newObject + value 字段赋值（记录取舍）
- 多维数组 clone 只拷贝最外层（JVM 语义）
- char 数组（`[]C`）的存储选择：`[]uint16`（UTF-16 code unit），String 互转时保持一致

## 验收标准（TDD 先行）

`tests/java/array/`：

| # | 用例 | 期望 |
|---|------|------|
| 1 | `ArrayClone`（int[]、String[]、int[2][3] 最外层 clone） | 与 java 一致 |
| 2 | `ReflectArray`（Array.getLength/get/set/newInstance） | 一致 |
| 3 | `ArrayStoreEx`（String[] 放入 Object 引用 → 捕获 ArrayStoreException） | 一致 |
| 4 | `ArraysTool`（Arrays.toString/sort） | 一致 |
| 5 | `Matrix`（int[5][5] 矩阵乘法） | 一致 |
| 6 | 数组类元数据单测（`[I`/`[[Ljava/lang/String;` 的 super/接口） | 规范一致 |
| 7 | heap 包覆盖率保持 ≥ 80% | 通过 |
| 8 | 全量回归 | 通过 |

## 完成动作

- commit：`feat: 数组类/clone/reflect.Array natives`
- tag：`m13`
- 更新 README 进度表
