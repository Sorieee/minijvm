// 常量池（M03 任务 2，对应书 3.2 节）。
package classfile

// ConstantPool 是常量池，**按槽位索引**（JVM 规范 §4.4）。
//
// 关键点（书里的经典坑）：long 与 double 占**两个**槽位——条目本身放在
// 第一个槽，紧随其后的槽留空（nil）。所以按索引取条目时必须直接用槽位
// 号索引数组，绝不能理解成「第 n 个条目」，否则常量池索引会整体错位，
// 解析出的类名/方法名全是乱的。
//
// 下标 0 恒为 nil（规范规定从 1 开始编号）。
type ConstantPool []ConstantInfo

// readConstantPool 读取整个常量池。
//
// 布局：u2 constant_pool_count + constant_pool_count-1 个槽位。
// 注意 constant_pool_count 是**槽位数**而非条目数——它已经包含了
// long/double 占用的额外槽位，所以循环上界直接用它。
func readConstantPool(r *ClassReader) (ConstantPool, error) {
	count, err := r.ReadUint16()
	if err != nil {
		return nil, err
	}

	cp := make(ConstantPool, count) // 下标 0 留空
	for i := 1; i < int(count); i++ {
		tag, err := r.ReadUint8()
		if err != nil {
			return nil, err
		}
		info, err := readConstantInfo(r, tag)
		if err != nil {
			return nil, err
		}
		cp[i] = info

		// long/double 占两个槽：跳过紧随其后的空槽。
		if tag == ConstantTagLong || tag == ConstantTagDouble {
			i++
		}
	}
	return cp, nil
}

// GetConstantInfo 按槽位索引取常量条目。
//
// 索引越界、或该槽是 long/double 的空槽时返回 nil（调用方需判空）。
func (cp ConstantPool) GetConstantInfo(index uint16) ConstantInfo {
	if int(index) >= len(cp) {
		return nil
	}
	return cp[index]
}

// GetUtf8 取 CONSTANT_Utf8 条目的字符串内容。
// 索引无效或类型不符时返回空串。
func (cp ConstantPool) GetUtf8(index uint16) string {
	info, ok := cp.GetConstantInfo(index).(*ConstantUtf8Info)
	if !ok {
		return ""
	}
	return string(info.Bytes)
}

// GetClassName 取 CONSTANT_Class 条目的类名（如 "java/lang/Object"）。
// 索引无效或类型不符时返回空串。
func (cp ConstantPool) GetClassName(index uint16) string {
	info, ok := cp.GetConstantInfo(index).(*ConstantClassInfo)
	if !ok {
		return ""
	}
	return cp.GetUtf8(info.NameIndex)
}

// GetNameAndType 取 CONSTANT_NameAndType 条目的 (name, descriptor)。
// 索引无效或类型不符时返回 ("", "")。
func (cp ConstantPool) GetNameAndType(index uint16) (name string, descriptor string) {
	info, ok := cp.GetConstantInfo(index).(*ConstantNameAndTypeInfo)
	if !ok {
		return "", ""
	}
	return cp.GetUtf8(info.NameIndex), cp.GetUtf8(info.DescriptorIndex)
}

// GetConstantValue 取常量条目的字面值。
//
// 返回 Go 原生类型：int32 / float32 / int64 / float64 / string；
// CONSTANT_Class、NameAndType 这类引用型常量返回 nil（它们不是字面值）。
// 供 ConstantValue 属性与静态字段初值使用（M06 的 prepare 阶段）。
func (cp ConstantPool) GetConstantValue(index uint16) any {
	switch info := cp.GetConstantInfo(index).(type) {
	case *ConstantIntegerInfo:
		return info.Value
	case *ConstantFloatInfo:
		return info.Value
	case *ConstantLongInfo:
		return info.Value
	case *ConstantDoubleInfo:
		return info.Value
	case *ConstantUtf8Info:
		return string(info.Bytes)
	default:
		return nil
	}
}
