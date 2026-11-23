// CONSTANT_InterfaceMethodref（tag 11）。
package classfile

// ConstantInterfaceMethodrefInfo 是 CONSTANT_InterfaceMethodref 条目：
// 接口方法引用（invokeinterface 指令专用）。
//
// 结构与 Methodref 相同；区别在语义——被引用者是**接口**上的方法。
type ConstantInterfaceMethodrefInfo struct {
	ClassIndex       uint16 // 指向 CONSTANT_Class（通常是接口）
	NameAndTypeIndex uint16 // 指向 CONSTANT_NameAndType
}

// Tag 返回常量池 tag。
func (c *ConstantInterfaceMethodrefInfo) Tag() uint8 { return ConstantTagInterfaceMethodref }

// readInfo 读取 u2 class_index + u2 name_and_type_index。
func (c *ConstantInterfaceMethodrefInfo) readInfo(r *ClassReader) error {
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
