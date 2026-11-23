// CONSTANT_Utf8（tag 1）。
package classfile

// ConstantUtf8Info 是 CONSTANT_Utf8 条目：类名、方法名、描述符等字符串。
type ConstantUtf8Info struct {
	Bytes []byte // 原始字节（MUTF-8 编码）
}

// Tag 返回常量池 tag。
func (c *ConstantUtf8Info) Tag() uint8 { return ConstantTagUtf8 }

// readInfo 读取 u2 长度 + 长度个字节。
//
// 编码说明：JVM 用的是 **MUTF-8**（modified UTF-8），与标准 UTF-8 有两处
// 差异——U+0000 编码为 0xC0 0x80、增补平面字符用代理对（surrogate pair）
// 形式。类名/描述符/属性名全是 ASCII，实际不受影响；M03 直接按字节转
// string 处理即可（真要严格解码，留到 M12 字符串里程碑统一处理）。
func (c *ConstantUtf8Info) readInfo(r *ClassReader) error {
	length, err := r.ReadUint16()
	if err != nil {
		return err
	}
	b, err := r.ReadBytes(uint32(length))
	if err != nil {
		return err
	}
	c.Bytes = b
	return nil
}
