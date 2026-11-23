// CONSTANT_NameAndType（tag 12）。
package classfile

// ConstantNameAndTypeInfo 是 CONSTANT_NameAndType 条目：名字 + 描述符。
//
// 它是 Fieldref/Methodref/InterfaceMethodref 的组成部分，也是
// InvokeDynamic 的组成部分。描述符形如：
//
//	I     int
//	Ljava/lang/String;   引用类型
//	(I)V  方法：(int)void
//	[[I  二维 int 数组
type ConstantNameAndTypeInfo struct {
	NameIndex       uint16 // 指向 CONSTANT_Utf8（方法名/字段名）
	DescriptorIndex uint16 // 指向 CONSTANT_Utf8（描述符）
}

// Tag 返回常量池 tag。
func (c *ConstantNameAndTypeInfo) Tag() uint8 { return ConstantTagNameAndType }

// readInfo 读取 u2 name_index + u2 descriptor_index。
func (c *ConstantNameAndTypeInfo) readInfo(r *ClassReader) error {
	nameIndex, err := r.ReadUint16()
	if err != nil {
		return err
	}
	descIndex, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.NameIndex = nameIndex
	c.DescriptorIndex = descIndex
	return nil
}
