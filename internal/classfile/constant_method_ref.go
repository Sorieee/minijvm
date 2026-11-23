// CONSTANT_Methodref（tag 10）。
package classfile

// ConstantMethodrefInfo 是 CONSTANT_Methodref 条目：方法引用。
//
// 与 Fieldref 结构完全相同，只是语义不同（指向方法而非字段），
// 解释器执行 invoke* 指令时用它解析「谁的方法、叫什么、什么签名」。
type ConstantMethodrefInfo struct {
	ClassIndex       uint16 // 指向 CONSTANT_Class
	NameAndTypeIndex uint16 // 指向 CONSTANT_NameAndType
}

// Tag 返回常量池 tag。
func (c *ConstantMethodrefInfo) Tag() uint8 { return ConstantTagMethodref }

// readInfo 读取 u2 class_index + u2 name_and_type_index。
func (c *ConstantMethodrefInfo) readInfo(r *ClassReader) error {
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
