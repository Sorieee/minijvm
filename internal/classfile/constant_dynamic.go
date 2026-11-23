// CONSTANT_Dynamic（tag 17，v55+）。
package classfile

// ConstantDynamicInfo 是 CONSTANT_Dynamic 条目：动态常量（v55+）。
//
// v52（Java 8）的 class 文件里**不会出现**本类型，解析器支持它只为
// 完整性；结构与 InvokeDynamic 相同。
type ConstantDynamicInfo struct {
	BootstrapMethodAttrIndex uint16
	NameAndTypeIndex         uint16
}

// Tag 返回常量池 tag。
func (c *ConstantDynamicInfo) Tag() uint8 { return ConstantTagDynamic }

// readInfo 读取 u2 bootstrap_method_attr_index + u2 name_and_type_index。
func (c *ConstantDynamicInfo) readInfo(r *ClassReader) error {
	bsmIndex, err := r.ReadUint16()
	if err != nil {
		return err
	}
	natIndex, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.BootstrapMethodAttrIndex = bsmIndex
	c.NameAndTypeIndex = natIndex
	return nil
}
