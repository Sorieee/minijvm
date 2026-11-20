# M03 解析 class 文件

## 目标

完整解析 class 文件结构（JVM 规范 §4），输出结构化的 `ClassFile`，为运行时模型供料。

## 对应书章节

第 3 章「解析class文件」。

## 前置依赖

M02（能拿到类字节）。

## 任务清单

### 1. 大端读取器 — `internal/classfile/class_reader.go`

```go
type ClassReader struct { data []byte }
func (r *ClassReader) ReadUint8() uint8      // u1
func (r *ClassReader) ReadUint16() uint16    // u2
func (r *ClassReader) ReadUint32() uint32    // u4
func (r *ClassReader) ReadUint64() uint64
func (r *ClassReader) ReadBytes(n uint32) []byte
```

内部用 `encoding/binary.BigEndian`。越界读返回 error 或 panic 由上层 recover（选择：返回 error 风格贯穿，坏文件给 `ClassFormatError` 文案）。

### 2. 常量池 — `internal/classfile/constant_pool.go` + 各常量类型文件

- 17 种 tag 全覆盖（v52 实际会出现：Utf8/Integer/Float/Long/Double/Class/String/Fieldref/Methodref/InterfaceMethodref/NameAndType/MethodHandle/MethodType/InvokeDynamic；Dynamic/Module/Package 是 v55+，解析器支持但断言不出现）
- **Long/Double 占两个槽位**（cpCount 计数偏移的经典坑）
- 便捷查询方法：`GetClassName(index)`、`GetUtf8(index)`、`GetNameAndType(index)`、`GetConstantInfo(index)`

### 3. 成员信息 — `internal/classfile/member_info.go`

```go
type MemberInfo struct {
    cp              ConstantPool
    accessFlags     uint16
    name            string
    descriptor      string
    attributes      []AttributeInfo
}
func (mi *MemberInfo) CodeAttribute() *CodeAttribute
func (mi *MemberInfo) ConstantValueAttribute() ConstantValueAttribute
func (mi *MemberInfo) ExceptionsAttribute() ExceptionsAttribute
```

### 4. 属性表 — `internal/classfile/attribute_info.go` + 子文件

- 必须实现：`Code`（maxStack/maxLocals/code/**ExceptionTable 留 M09 解析字段但先读长度跳过→直接本里程碑一起解析掉，字段就位**）、`ConstantValue`、`Exceptions`、`LineNumberTable`（异常栈打印要用）、`SourceFile`
- 未知属性：读长度后跳过（健壮性要求）

### 5. ClassFile 顶层 — `internal/classfile/class_file.go`

```go
type ClassFile struct {
    minorVersion, majorVersion uint16
    constantPool   ConstantPool
    accessFlags    uint16
    thisClass, superClass uint16
    interfaces     []uint16
    fields, methods []MemberInfo
    attributes     []AttributeInfo
}
func Parse(classData []byte) (*ClassFile, error)  // 校验 magic CAFEBABE
func (cf *ClassFile) ClassName() string
func (cf *ClassFile) SuperClassName() string
func (cf *ClassFile) InterfaceNames() []string
```

### 6. main 接入（可选调试输出）

`-verbose:class` 时打印 `Parsed: <ClassName> (super=<SuperClassName>, methods=<n>)`。

## 关键难点

- Long/Double 双槽导致后续 cp 索引与条目数组下标错位——常量池必须按**槽位**索引而非条目下标
- 属性长度字段（u4）必须严格按声明消费，否则错位到下一个结构
- accessFlags 差异：类/字段/方法三套 flag 值不同，解析阶段只存原始 u16（语义判断放 heap 层）

## 验收标准（TDD 先行）

1. 用 `javac --release 8` 编译固定测试类 `tests/java/meta/Meta.java`（刻意包含：静态 final 常量、long/double 常量、抛异常方法、多接口），`javap -v` 人工记录期望值
2. 测试文件内嵌该 class 的字节（`go:embed` 或测试前脚本编译到 `testdata/`）

| # | 用例 | 期望 |
|---|------|------|
| 1 | Parse 成功 | 无错误 |
| 2 | magic 错误（前 4 字节改掉） | 返回格式错误 |
| 3 | majorVersion | 52 |
| 4 | ClassName / SuperClassName | 与 javap 输出一致 |
| 5 | 常量池 Long/Double 后的下一条目索引 | 槽位正确（对拍 javap 的 `#n` 编号） |
| 6 | 某方法的 Code 属性 | maxStack/maxLocals/code 长度与 javap 一致 |
| 7 | ConstantValue（static final int） | 值一致 |
| 8 | Exceptions 属性（抛 IOException 的方法） | 异常类名一致 |
| 9 | 接口数与接口名 | 与 javap 一致 |
| 10 | 覆盖率 | classfile 包 ≥ 80% |

## 完成动作

- commit：`feat: 解析 class 文件（常量池/成员/属性）`
- tag：`m03`
- 更新 README 进度表
