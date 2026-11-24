// Exceptions 属性（M03 任务 4）。
package classfile

// ExceptionsAttribute 是异常表属性：方法 throws 子句里声明的异常类型。
//
// 布局：u2 number_of_exceptions + u2 exception_index_table[number]。
// 每一项是常量池 Class 索引（异常类的内部形式类名）。
//
// 只出现在方法上；throws 子句为空时，属性本身通常不生成。
type ExceptionsAttribute struct {
	ExceptionIndexTable []uint16 // 常量池 Class 索引列表
}

// Name 返回属性名。
func (a *ExceptionsAttribute) Name() string { return "Exceptions" }

// Length 返回属性内容长度（u4）。
func (a *ExceptionsAttribute) Length() uint32 {
	return 2 + 2*uint32(len(a.ExceptionIndexTable))
}

// readInfo 读取 u2 数量 + u2 索引表。
func (a *ExceptionsAttribute) readInfo(r *ClassReader) error {
	attLen, err := r.ReadUint16()
	if err != nil {
		return err
	}
	exceptionIndexTable := make([]uint16, attLen)
	for i := uint16(0); i < attLen; i++ {
		exceptionIndexTable[i], err = r.ReadUint16()
		if err != nil {
			return err
		}
	}
	a.ExceptionIndexTable = exceptionIndexTable
	return nil
}
