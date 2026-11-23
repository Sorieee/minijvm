// CONSTANT_String（tag 8）。
package classfile

// ConstantStringInfo 是 CONSTANT_String 条目：字符串字面量。
//
// StringIndex 指向 CONSTANT_Utf8，内容是字符串内容本身
// （不是类名——类名走 CONSTANT_Class，别混淆）。
type ConstantStringInfo struct {
	StringIndex uint16
}

// Tag 返回常量池 tag。
func (c *ConstantStringInfo) Tag() uint8 { return ConstantTagString }

// readInfo 读取 u2 的 string_index。
func (c *ConstantStringInfo) readInfo(r *ClassReader) error {
	idx, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.StringIndex = idx
	return nil
}
