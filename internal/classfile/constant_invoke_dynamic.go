// CONSTANT_InvokeDynamic（tag 18）。
package classfile

// ConstantInvokeDynamicInfo 是 CONSTANT_InvokeDynamic 条目：动态调用点。
//
// BootstrapMethodAttrIndex 指向 BootstrapMethods 属性（类级属性之一），
// NameAndTypeIndex 指向被调用方法的 (name, descriptor)。
//
// 注意：本项目**不支持 invokedynamic**（不实现 lambda / 方法引用 /
// string concat 走 indy 的场景；测试程序一律用 --release 8 编译，
// 字符串拼接走 StringBuilder）。此类型仅为完整性而存在——遇到它
// 应在 M05 解释器实现时明确报「不支持」。
type ConstantInvokeDynamicInfo struct {
	BootstrapMethodAttrIndex uint16
	NameAndTypeIndex         uint16
}

// Tag 返回常量池 tag。
func (c *ConstantInvokeDynamicInfo) Tag() uint8 { return ConstantTagInvokeDynamic }

// readInfo 读取 u2 bootstrap_method_attr_index + u2 name_and_type_index。
func (c *ConstantInvokeDynamicInfo) readInfo(r *ClassReader) error {
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
