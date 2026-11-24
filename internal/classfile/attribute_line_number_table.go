// LineNumberTable 属性（M03 任务 4）。
package classfile

// LineNumberTableAttribute 是行号表属性：字节码 pc ↔ 源码行号映射。
//
// 布局：u2 line_number_table_length +
// line_number_table[line_number_table_length]，
// 每项是 (u2 start_pc, u2 line_number)。
//
// 用途：M09 抛异常时打印「异常发生在源文件的哪一行」。
type LineNumberTableAttribute struct {
	LineNumberTable []LineNumberTableEntry
}

// LineNumberTableEntry 是行号表的一项。
type LineNumberTableEntry struct {
	StartPc    uint16 // 起始 pc（字节码偏移）
	LineNumber uint16 // 对应的源码行号
}

// Name 返回属性名。
func (a *LineNumberTableAttribute) Name() string { return "LineNumberTable" }

// Length 返回属性内容长度（u4）：2 字节 count + 每项 4 字节。
func (a *LineNumberTableAttribute) Length() uint32 {
	return 2 + 4*uint32(len(a.LineNumberTable))
}

// readInfo 读取 u2 数量 + 每项 (u2 start_pc, u2 line_number)。
func (a *LineNumberTableAttribute) readInfo(r *ClassReader) error {
	count, err := r.ReadUint16()
	if err != nil {
		return err
	}
	table := make([]LineNumberTableEntry, count)
	for i := 0; i < int(count); i++ {
		startPc, err := r.ReadUint16()
		if err != nil {
			return err
		}
		lineNumber, err := r.ReadUint16()
		if err != nil {
			return err
		}
		table[i] = LineNumberTableEntry{StartPc: startPc, LineNumber: lineNumber}
	}
	a.LineNumberTable = table
	return nil
}
