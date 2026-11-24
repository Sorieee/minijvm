// SourceFile 属性（M03 任务 4）。
package classfile

// SourceFileAttribute 是源文件名属性：class 文件来自哪个 .java。
//
// 布局：u2 sourcefile_index（指向常量池 Utf8，内容是源文件名，
// 如 "Meta.java"）。
//
// 用途：M09 抛异常时打印「哪个源文件的哪一行」。
type SourceFileAttribute struct {
	SourceFileIndex uint16
}

// Name 返回属性名。
func (a *SourceFileAttribute) Name() string { return "SourceFile" }

// Length 返回属性内容长度（u4）——固定为 2。
func (a *SourceFileAttribute) Length() uint32 { return 2 }

// readInfo 读取 u2 sourcefile_index。
func (a *SourceFileAttribute) readInfo(r *ClassReader) error {
	fileIndex, err := r.ReadUint16()
	if err != nil {
		return err
	}
	a.SourceFileIndex = fileIndex
	return nil
}
