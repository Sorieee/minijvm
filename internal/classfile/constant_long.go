// CONSTANT_Long（tag 5）。
package classfile

import "encoding/binary"

// ConstantLongInfo 是 CONSTANT_Long 条目：8 字节长整型常量。
//
// 注意：本条目在常量池里**占两个槽位**（自身 + 一个空槽），
// 槽位处理见 constant_pool.go 的 readConstantPool。
type ConstantLongInfo struct {
	Value int64 // 有符号 64 位
}

// Tag 返回常量池 tag。
func (c *ConstantLongInfo) Tag() uint8 { return ConstantTagLong }

// readInfo 读取 8 字节（大端）并按有符号 64 位解码。
func (c *ConstantLongInfo) readInfo(r *ClassReader) error {
	b, err := r.ReadBytes(8)
	if err != nil {
		return err
	}
	c.Value = int64(binary.BigEndian.Uint64(b))
	return nil
}
