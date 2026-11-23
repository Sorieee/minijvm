// CONSTANT_Module（tag 19，v55+）与 CONSTANT_Package（tag 20，v55+）。
package classfile

// ConstantModuleInfo 是 CONSTANT_Module 条目：模块名（v55+ 的 JPMS 引入）。
//
// v52 的 class 文件里不会出现，解析器支持它只为完整性。
type ConstantModuleInfo struct {
	NameIndex uint16 // 指向 CONSTANT_Utf8（模块名）
}

// Tag 返回常量池 tag。
func (c *ConstantModuleInfo) Tag() uint8 { return ConstantTagModule }

// readInfo 读取 u2 name_index。
func (c *ConstantModuleInfo) readInfo(r *ClassReader) error {
	idx, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.NameIndex = idx
	return nil
}

// ConstantPackageInfo 是 CONSTANT_Package 条目：包名（v55+ 的 JPMS 引入）。
//
// v52 的 class 文件里不会出现，解析器支持它只为完整性。
type ConstantPackageInfo struct {
	NameIndex uint16 // 指向 CONSTANT_Utf8（包名，内部形式用 '/' 分隔）
}

// Tag 返回常量池 tag。
func (c *ConstantPackageInfo) Tag() uint8 { return ConstantTagPackage }

// readInfo 读取 u2 name_index。
func (c *ConstantPackageInfo) readInfo(r *ClassReader) error {
	idx, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.NameIndex = idx
	return nil
}
