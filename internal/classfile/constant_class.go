// CONSTANT_Class（tag 7）。
package classfile

// ConstantClassInfo 是 CONSTANT_Class 条目：指向一个类/接口/数组类型。
//
// NameIndex 指向常量池里的 CONSTANT_Utf8，内容是**内部形式**的类名
// （斜杠分隔，如 "java/lang/Object"；数组类型是 "[Ljava/lang/String;"）。
type ConstantClassInfo struct {
	NameIndex uint16
}

// Tag 返回常量池 tag。
func (c *ConstantClassInfo) Tag() uint8 { return ConstantTagClass }

// readInfo 读取 u2 的 name_index。
func (c *ConstantClassInfo) readInfo(r *ClassReader) error {
	idx, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.NameIndex = idx
	return nil
}
