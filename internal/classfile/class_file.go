// ClassFile 顶层结构（M03 任务 5，对应书 3.8 节）。
package classfile

import "fmt"

// magic 是 class 文件的魔数：0xCAFEBABE。
const magic uint32 = 0xCAFEBABE

// ClassFile 是解析后的 class 文件（对应一个类/接口）。
//
// 字段全部导出：M06 起 heap 层的运行时类模型要直接读这些字段
// （方法表、字段表、常量池），不导出就得在 classfile 包里再包一层
// 访问器，纯属多余。
//
// 命名说明：文档草稿里写的是小写字段（minorVersion 等），这里改成
// 导出形式 MinorVersion 等——Go 惯例是导出字段用大写开头。
type ClassFile struct {
	MinorVersion uint16 // 次版本号（v52 是 0）
	MajorVersion uint16 // 主版本号（Java 8 = 52；本项目只支持 ≤ 52）

	ConstantPool ConstantPool // 常量池（按槽位索引）

	AccessFlags uint16 // 类访问标志（public/final/abstract/interface…）
	ThisClass   uint16 // 指向本类的常量池 Class 索引
	SuperClass  uint16 // 指向父类的常量池 Class 索引（0 = 只有 Object）

	Interfaces []uint16 // 直接实现的接口，常量池 Class 索引列表

	Fields  []MemberInfo // 字段表
	Methods []MemberInfo // 方法表

	Attributes []AttributeInfo // 类级属性（SourceFile / Signature / BootstrapMethods…）
}

// Parse 解析 class 文件字节，返回结构化的 ClassFile。
//
// 完整布局（JVM 规范 §4.1）：
//
//	u4 magic = 0xCAFEBABE
//	u2 minor_version
//	u2 major_version
//	u2 constant_pool_count + 常量池
//	u2 access_flags
//	u2 this_class
//	u2 super_class
//	u2 interfaces_count + interfaces[]
//	u2 fields_count + field_info[]
//	u2 methods_count + method_info[]
//	u2 attributes_count + attribute_info[]
//
// 错误约定：格式非法一律返回包装 ErrClassFormat 的错误（含出错偏移），
// 调用方用 errors.Is(err, ErrClassFormat) 判定。
//
// 版本校验：major_version > 52 时应报错（本项目只支持 Java 8 及更早）。
func Parse(classData []byte) (*ClassFile, error) {
	panic("TODO(M03): not implemented")
}

// ClassName 返回本类的内部形式类名（如 "java/lang/Object"）。
//
// 内部形式用斜杠分隔；面向用户的点分形式（java.lang.Object）在
// 打印错误信息时转换。
func (cf *ClassFile) ClassName() string {
	panic("TODO(M03): not implemented")
}

// SuperClassName 返回父类的内部形式类名；本类无父类（只有 Object）时
// 返回空串。
func (cf *ClassFile) SuperClassName() string {
	panic("TODO(M03): not implemented")
}

// InterfaceNames 返回直接实现的接口的内部形式类名列表。
func (cf *ClassFile) InterfaceNames() []string {
	panic("TODO(M03): not implemented")
}

// String 返回便于调试的一行摘要（类名 + 版本）。
func (cf *ClassFile) String() string {
	return fmt.Sprintf("ClassFile{%s, v%d.%d, methods=%d, fields=%d}",
		cf.ClassName(), cf.MajorVersion, cf.MinorVersion, len(cf.Methods), len(cf.Fields))
}
