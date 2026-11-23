// CONSTANT_Double（tag 6）。
package classfile

import (
	"encoding/binary"
	"math"
)

// ConstantDoubleInfo 是 CONSTANT_Double 条目：8 字节双精度浮点常量。
//
// 注意：本条目在常量池里**占两个槽位**（自身 + 一个空槽），
// 槽位处理见 constant_pool.go 的 readConstantPool。
type ConstantDoubleInfo struct {
	Value float64 // IEEE 754 双精度
}

// Tag 返回常量池 tag。
func (c *ConstantDoubleInfo) Tag() uint8 { return ConstantTagDouble }

// readInfo 读取 8 字节并按 IEEE 754 双精度解码。
func (c *ConstantDoubleInfo) readInfo(r *ClassReader) error {
	b, err := r.ReadBytes(8)
	if err != nil {
		return err
	}
	c.Value = math.Float64frombits(binary.BigEndian.Uint64(b))
	return nil
}
