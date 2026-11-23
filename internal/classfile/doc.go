// Package classfile 实现 class 文件的解析（JVM 规范 §4）。
//
// 里程碑 M03。对应书第 3 章「解析class文件」。
//
// 结构（与书的 ch3 布局一一对应）：
//
//	class_reader.go          大端读取器（u1/u2/u4/u8）+ ErrClassFormat 哨兵
//	constant_info.go         ConstantInfo 接口 + 17 种 tag 常量 + 分派
//	constant_pool.go         ConstantPool（按槽位索引）+ 便捷查询
//	constant_*.go            17 种常量条目类型（一类一文件）
//	member_info.go           MemberInfo（字段与方法的公共结构）
//	attribute_info.go        AttributeInfo 接口 + 属性表读取 + 未知属性跳过
//	attribute_*.go           Code / ConstantValue / Exceptions /
//	                         LineNumberTable / SourceFile 属性
//	class_file.go            ClassFile 顶层结构 + Parse 入口
//
// 两条贯穿全包的约定：
//
//  1. **常量池按槽位索引**：long/double 占两个槽，第二个槽为 nil。
//     按「第 n 个条目」理解会让所有索引错位（书里的经典坑）。
//  2. **属性长度必须严格消费**：多读少读都会让后续结构整体错位；
//     未知属性按声明长度跳过，保证能解析 java.lang.Object 这种
//     带一堆陌生属性的真实类文件。
//
// 测试夹具：tests/java/meta/Meta.java（静态 final 常量、long/double
// 常量、throws 方法、多接口），先用 scripts/build-testdata.sh 编译到
// internal/classfile/testdata/，再由测试用 go:embed 内嵌字节。
package classfile
