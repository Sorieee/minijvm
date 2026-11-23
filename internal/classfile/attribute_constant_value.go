// ConstantValue 属性（M03 任务 4）。
package classfile

// ConstantValueAttribute 是常量值属性：字段的编译期常量初值。
//
// 布局：u2 constantvalue_index（指向常量池里的 Integer/Float/Long/
// Double/String 条目）。
//
// 只出现在 **static final** 字段上；取值时用
// cp.GetConstantValue(index) 拿到 Go 原生类型的字面值。
type ConstantValueAttribute struct {
	ConstantValueIndex uint16
}

// Name 返回属性名。
func (a *ConstantValueAttribute) Name() string { return "ConstantValue" }

// Length 返回属性内容长度（u4）——固定为 2。
func (a *ConstantValueAttribute) Length() uint32 { return 2 }

// readInfo 读取 u2 constantvalue_index。
func (a *ConstantValueAttribute) readInfo(r *ClassReader) error {
	panic("TODO(M03): not implemented")
}
