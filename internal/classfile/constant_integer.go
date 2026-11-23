// CONSTANT_Integer（tag 3）。
package classfile

import "encoding/binary"

// ConstantIntegerInfo 是 CONSTANT_Integer 条目：4 字节整型常量。
type ConstantIntegerInfo struct {
	Value int32 // 有符号 32 位
}

// Tag 返回常量池 tag。
func (c *ConstantIntegerInfo) Tag() uint8 { return ConstantTagInteger }

// readInfo 读取 4 字节（大端）并按有符号整数解码。
func (c *ConstantIntegerInfo) readInfo(r *ClassReader) error {
	b, err := r.ReadBytes(4)
	if err != nil {
		return err
	}
	c.Value = int32(binary.BigEndian.Uint32(b))
	return nil
}
