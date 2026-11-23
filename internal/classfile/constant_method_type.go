// CONSTANT_MethodType（tag 16）。
package classfile

// ConstantMethodTypeInfo 是 CONSTANT_MethodType 条目：方法类型。
//
// DescriptorIndex 指向一个方法描述符（如 "(I)Ljava/lang/String;"）。
// 它与 invokedynamic 的 lambda 引导方法配套使用——本项目不支持
// invokedynamic（见 tests/java 编写约束），所以此类型在 v52 的
// 普通 class 文件里几乎不会出现，但解析器要能认。
type ConstantMethodTypeInfo struct {
	DescriptorIndex uint16
}

// Tag 返回常量池 tag。
func (c *ConstantMethodTypeInfo) Tag() uint8 { return ConstantTagMethodType }

// readInfo 读取 u2 descriptor_index。
func (c *ConstantMethodTypeInfo) readInfo(r *ClassReader) error {
	idx, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.DescriptorIndex = idx
	return nil
}
