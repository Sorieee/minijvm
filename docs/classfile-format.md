# Class 文件结构速查

依据：JVM 规范（JVMS SE 8）§4.1、§4.7；对应书第 3 章与 [M03 里程碑](milestones/M03-classfile.md)。

## 顶层布局

class 文件是**扁平、顺序**的字节流：所有结构一段接一段，没有跳转。解析时用同一个 `ClassReader`（单一游标 `pos`）从开头一路往后读。

```
ClassFile:
  u4 magic                      ← Parse 读（校验 CAFEBABE）
  u2 minor_version
  u2 major_version
  u2 constant_pool_count + cp[] ← Parse / readConstantPool 读
  u2 access_flags
  u2 this_class
  u2 super_class
  u2 interfaces_count + 接口表
  u2 fields_count + 字段表      ← 每个 field 末尾自带一张属性表
  u2 methods_count + 方法表     ← 每个 method 末尾自带一张属性表
  u2 attributes_count + 属性表  ← 类级的属性表，到这里才轮到 readAttributes
```

## 要点

- `u1` / `u2` / `u4` 分别是 1 / 2 / 4 字节无符号整数，**大端**序。
- 变长表统一是「u2 count + count 个条目」打头：常量池、接口表、字段表、方法表、属性表，以及 Code 内部的异常表、行号表等。count 是读取边界的唯一依据，属性条目是变长的，没有结束哨兵。
- `long` / `double` 常量在常量池里占**两个槽位**，第二个槽为 nil。所以常量池要按槽位索引，不能按"第 n 个条目"数。
- 属性表布局在四个位置完全一致，因此 `readAttributes` 可复用：ClassFile 末尾、field 末尾、method 末尾、Code 属性末尾。每次调用前由调用方把 reader 推到该段的起点。
- 单条属性的通用头是 `u2 attribute_name_index + u4 attribute_length + u1 info[attribute_length]`；名字索引指向常量池的 Utf8 条目。不认识的属性按声明长度跳过（`UnknownAttribute`），且必须严格消费 length 字节，否则后续结构整体错位。

## 版本约定

本项目只支持 Java 8 及更早：`major_version <= 52`。测试夹具由 `scripts/build-testdata.sh` 用 `javac --release 8` 生成到 `internal/classfile/testdata/`。
