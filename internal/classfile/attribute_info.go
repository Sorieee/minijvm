// 属性表（M03 任务 4，对应书 3.6 节）。
package classfile

import "fmt"

// AttributeInfo 是属性表条目的统一接口。
//
// 接口带未导出方法 readInfo：封闭接口，只有本包能实现——防止外部
// 造出属性条目破坏「属性名 ↔ 类型」的对应关系。
//
// 属性在 class 文件里的通用布局（JVM 规范 §4.7）：
//
//	u2 attribute_name_index   → 常量池 Utf8，得到属性名
//	u4 attribute_length       → 属性内容字节数
//	u1 info[attribute_length] → 属性内容
//
// 关键：attribute_length 必须**严格按声明消费**——多读或少读都会让
// 后续结构整体错位。未知属性按 length 跳过即可（健壮性要求）。
type AttributeInfo interface {
	// Name 返回属性名（如 "Code"、"SourceFile"）。
	Name() string
	// Length 返回属性内容长度（u4）。
	Length() uint32
	// readInfo 读取属性内容（属性名与长度已由调用方读出）。
	readInfo(r *ClassReader) error
}

// readAttributeInfo 读取一个属性条目并分派到具体类型。
//
// name 与 length 由调用方（readAttributes）从属性表里读出后传入，
// 这样未知属性的跳过长度也由调用方统一掌控。
func readAttributeInfo(r *ClassReader, name string, length uint32) (AttributeInfo, error) {
	panic("TODO(M03): not implemented")
}

// readAttributes 读取整个属性表（u2 count + count 个属性条目）。
//
// 类、字段、方法、Code 属性内部都复用这个函数。
func readAttributes(r *ClassReader) ([]AttributeInfo, error) {
	panic("TODO(M03): not implemented")
}

// UnknownAttribute 是本解析器不认识的属性：只保留名字与长度，内容按
// length 跳过。
//
// 为什么要有它：class 文件里会出现本项目不关心的属性（Signature、
// StackMapTable、BootstrapMethods、InnerClasses……）。遇到它们必须
// 能安全跳过，而不是报错——否则连 java.lang.Object 都解析不了。
//
// M03 只要求「能跳过」；若将来某个属性对运行时有用（如 M09 的
// StackMapTable 用于验证），再为它加专门类型。
type UnknownAttribute struct {
	AttrName   string
	AttrLength uint32
}

// Name 返回属性名。
func (a *UnknownAttribute) Name() string { return a.AttrName }

// Length 返回属性内容长度。
func (a *UnknownAttribute) Length() uint32 { return a.AttrLength }

// readInfo 跳过属性内容（长度已知）。
func (a *UnknownAttribute) readInfo(r *ClassReader) error {
	if err := r.Skip(a.AttrLength); err != nil {
		return err
	}
	return nil
}

// attributeNameToString 是调试辅助：把属性名解析成字符串（供错误信息用）。
func attributeNameToString(cp ConstantPool, nameIndex uint16) string {
	return cp.GetUtf8(nameIndex)
}

// errUnexpectedAttribute 构造「属性长度与实际消费不符」的错误。
//
// 实现建议：各属性类型的 readInfo 记录调用前的 r.Pos()，读完��
// 比较差值与声明的 length 是否一致；不一致说明解析错位，必须报错
// 而不是继续往下读（错位会让后面所有结构都读成垃圾）。
func errUnexpectedAttribute(name string, want, got uint32) error {
	return fmt.Errorf("%w: 属性 %q 声明长度 %d，实际消费 %d 字节", ErrClassFormat, name, want, got)
}
