// CONSTANT_MethodHandle（tag 15）。
package classfile

// ConstantMethodHandleInfo 是 CONSTANT_MethodHandle 条目：方法句柄。
//
// ReferenceKind 的取值（JVM 规范 §4.4.8）：
//
//	1..4  对应 getField/putField/getStatic/putStatic
//	5..8  对应 invokevirtual/invokespecial/invokestatic/invokeinterface
//	9     invokeinterface（interface 方法引用，v52 新增）
//
// ReferenceIndex 的含义随 ReferenceKind 而变：1..4 指向 Fieldref，
// 5..9 指向 Methodref/InterfaceMethodref。
type ConstantMethodHandleInfo struct {
	ReferenceKind  uint8
	ReferenceIndex uint16
}

// Tag 返回常量池 tag。
func (c *ConstantMethodHandleInfo) Tag() uint8 { return ConstantTagMethodHandle }

// readInfo 读取 u1 reference_kind + u2 reference_index。
func (c *ConstantMethodHandleInfo) readInfo(r *ClassReader) error {
	kind, err := r.ReadUint8()
	if err != nil {
		return err
	}
	index, err := r.ReadUint16()
	if err != nil {
		return err
	}
	c.ReferenceKind = kind
	c.ReferenceIndex = index
	return nil
}
