// CONSTANT_Fieldref（tag 9）。
package classfile

// ConstantFieldrefInfo 是 CONSTANT_Fieldref 条目：字段引用。
//
// 由「类索引 + 名字和描述符索引」两部分组成：
// 解析出的语义是「<ClassIndex 这个类>.<name>:<descriptor> 字段」。
type ConstantFieldrefInfo struct {
	ClassIndex       uint16 // 指向 CONSTANT_Class
	NameAndTypeIndex uint16 // 指向 CONSTANT_NameAndType
}

// Tag 返回常量池 tag。
func (c *ConstantFieldrefInfo) Tag() uint8 { return ConstantTagFieldref }

// readInfo 读取 u2 class_index + u2 name_and_type_index。
func (c *ConstantFieldrefInfo) readInfo(r *ClassReader) error {
	classIndex, err := r.ReadUint16()
	if err != nil {
		return err
	}
	natIndex, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.ClassIndex = classIndex
	c.NameAndTypeIndex = natIndex
	return nil
}
