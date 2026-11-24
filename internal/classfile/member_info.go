// 字段与方法的成员信息（M03 任务 3，对应书 3.5 节）。
package classfile

// MemberInfo 描述 class 文件里的一个字段或一个方法。
//
// 字段与方法在 class 文件里的结构**完全相同**，所以共用一个结构体，
// 由所在位置（fields 表 / methods 表）区分语义。
//
// CP 是所属 ClassFile 的常量池（同一个池），放在成员里是为了让成员
// 能自己解析引用（方法体里的常量池索引、属性名等），不必到处传 cp。
type MemberInfo struct {
	CP          ConstantPool
	AccessFlags uint16 // 访问标志（字段/方法/类三套语义不同，解析阶段只存原值）
	Name        string // 成员名（来自 Utf8）
	Descriptor  string // 描述符（来自 Utf8）
	Attributes  []AttributeInfo
}

// readMemberInfo 读取一个字段或方法。
//
// 布局：u2 access_flags + u2 name_index + u2 descriptor_index +
// u2 attributes_count + attributes。
func readMemberInfo(r *ClassReader) (MemberInfo, error) {
	accessFlags, err := r.ReadUint16()
	if err != nil {
		return MemberInfo{}, err
	}
	nameIndex, err := r.ReadUint16()
	if err != nil {
		return MemberInfo{}, err
	}
	descriptorIndex, err := r.ReadUint16()
	if err != nil {
		return MemberInfo{}, err
	}
	attributes, err := readAttributes(r)
	if err != nil {
		return MemberInfo{}, err
	}

	return MemberInfo{
		CP:          r.cp,
		AccessFlags: accessFlags,
		Name:        r.cp.GetUtf8(nameIndex),
		Descriptor:  r.cp.GetUtf8(descriptorIndex),
		Attributes:  attributes,
	}, nil
}

// CodeAttribute 返回该方法的 Code 属性；没有则返回 nil。
//
// 只有方法才有 Code 属性；字段调此方法恒返回 nil。
func (mi *MemberInfo) CodeAttribute() *CodeAttribute {
	for _, attr := range mi.Attributes {
		if code, ok := attr.(*CodeAttribute); ok {
			return code
		}
	}
	return nil
}

// ConstantValueAttribute 返回该成员的 ConstantValue 属性；没有则返回 nil。
//
// 只有「static final」字段才有（编译期常量）；方法恒为 nil。
func (mi *MemberInfo) ConstantValueAttribute() *ConstantValueAttribute {
	for _, attr := range mi.Attributes {
		if constant, ok := attr.(*ConstantValueAttribute); ok {
			return constant
		}
	}
	return nil
}

// ExceptionsAttribute 返回该方法的 Exceptions 属性；没有则返回 nil。
//
// 只有「throws 子句」的方法才有（且可能为空表）。
func (mi *MemberInfo) ExceptionsAttribute() *ExceptionsAttribute {
	for _, attr := range mi.Attributes {
		if exceptions, ok := attr.(*ExceptionsAttribute); ok {
			return exceptions
		}
	}
	return nil
}
