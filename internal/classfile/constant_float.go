// CONSTANT_Float（tag 4）。
package classfile

import (
	"encoding/binary"
	"math"
)

// ConstantFloatInfo 是 CONSTANT_Float 条目：4 字节浮点常量。
type ConstantFloatInfo struct {
	Value float32 // IEEE 754 单精度
}

// Tag 返回常量池 tag。
func (c *ConstantFloatInfo) Tag() uint8 { return ConstantTagFloat }

// readInfo 读取 4 字节并按 IEEE 754 单精度解码。
func (c *ConstantFloatInfo) readInfo(r *ClassReader) error {
	b, err := r.ReadBytes(4)
	if err != nil {
		return err
	}
	c.Value = math.Float32frombits(binary.BigEndian.Uint32(b))
	return nil
}
