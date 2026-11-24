// Code 属性（M03 任务 4，对应书 3.7 节）。
package classfile

// CodeAttribute 是方法体属性：字节码 + 局部变量/栈信息 + 异常表。
//
// 布局（JVM 规范 §4.7.3）：
//
//	u2 max_stack
//	u2 max_locals
//	u4 code_length
//	u1 code[code_length]
//	u2 exception_table_length
//	exception_table[exception_table_length]
//	u2 attributes_count
//	attributes[attributes_count]
//
// Code 属性**只出现在方法**里，且每个方法最多一个。
type CodeAttribute struct {
	MaxStack  uint16 // 操作数栈最大深度（解释器 M05 用它预分配栈）
	MaxLocals uint16 // 局部变量表最大槽位数（解释器 M05 用它预分配局部变量）
	Code      []byte // 字节码指令序列（M05 逐条取指执行）

	// ExceptionTable 是异常处理表（M09 真正使用）。
	//
	// M03 就把字段解析到位（而不是「读长度跳过」），这样 M09 只需接上
	// 异常匹配逻辑，不用回头改解析器。
	ExceptionTable []ExceptionTableEntry

	// LineNumberTable 是行号表（M09 抛异常时打印「哪一行出错了」用）。
	LineNumberTable []LineNumberTableEntry

	// LocalVariableTable / LocalVariableTypeTable：v52 里只有 javac 带
	// -g 编译才会生成。M03 先把字段位留好（属性按未知属性跳过即可），
	// 真正解析留到需要时（M04/M05 若要调试局部变量名）。
	LocalVariableTable     []LocalVariableTableEntry
	LocalVariableTypeTable []LocalVariableTypeTableEntry

	// attrLength 是属性头里的 u4 attribute_length；readInfo(r) 拿不到它，
	// 由 readAttributeInfo 构造本属性时回填，Length() 返回它。
	attrLength uint32
}

// ExceptionTableEntry 是异常处理表的一项（JVM 规范 §4.7.3）。
type ExceptionTableEntry struct {
	StartPc   uint16 // 受保护的代码起始 pc（含）
	EndPc     uint16 // 受保护的代码结束 pc（不含）
	HandlerPc uint16 // 异常处理器的起始 pc
	CatchType uint16 // 捕获的异常类型：常量池 Class 索引；0 = 捕获所有（finally）
}

// LocalVariableTableEntry 是局部变量表的一项（调试用）。
type LocalVariableTableEntry struct {
	StartPc         uint16
	Length          uint16 // 作用域长度（字节码条数）
	NameIndex       uint16
	DescriptorIndex uint16
	Slot            uint16 // 局部变量槽位号
}

// LocalVariableTypeTableEntry 是局部变量类型表的一项（泛型擦除前的类型）。
type LocalVariableTypeTableEntry struct {
	StartPc        uint16
	Length         uint16
	NameIndex      uint16
	SignatureIndex uint16
	Slot           uint16
}

// Name 返回属性名。
func (a *CodeAttribute) Name() string { return "Code" }

// Length 返回属性内容长度（u4，由 readAttributeInfo 回填）。
func (a *CodeAttribute) Length() uint32 { return a.attrLength }

// readInfo 读取 Code 属性内容。
//
// 注意：嵌套在 Code 末尾的属性表用 readAttributes 读（那里可能出现
// LineNumberTable / LocalVariableTable / StackMapTable 等）。
func (a *CodeAttribute) readInfo(r *ClassReader) error {
	maxStack, err := r.ReadUint16()
	if err != nil {
		return err
	}
	maxLocals, err := r.ReadUint16()
	if err != nil {
		return err
	}
	codeLength, err := r.ReadUint32()
	if err != nil {
		return err
	}
	code, err := r.ReadBytes(codeLength)
	if err != nil {
		return err
	}

	exceptionCount, err := r.ReadUint16()
	if err != nil {
		return err
	}
	exceptionTable := make([]ExceptionTableEntry, exceptionCount)
	for i := 0; i < int(exceptionCount); i++ {
		startPc, err := r.ReadUint16()
		if err != nil {
			return err
		}
		endPc, err := r.ReadUint16()
		if err != nil {
			return err
		}
		handlerPc, err := r.ReadUint16()
		if err != nil {
			return err
		}
		catchType, err := r.ReadUint16()
		if err != nil {
			return err
		}
		exceptionTable[i] = ExceptionTableEntry{
			StartPc: startPc, EndPc: endPc, HandlerPc: handlerPc, CatchType: catchType,
		}
	}

	// 末尾的嵌套属性表交给 readAttributes；只把 LineNumberTable 提到
	// 专属字段（M09 用），其余（LocalVariableTable/StackMapTable 等）
	// 按未知属性跳过即可。
	nested, err := readAttributes(r)
	if err != nil {
		return err
	}
	for _, attr := range nested {
		if lnt, ok := attr.(*LineNumberTableAttribute); ok {
			a.LineNumberTable = lnt.LineNumberTable
		}
	}

	a.MaxStack = maxStack
	a.MaxLocals = maxLocals
	a.Code = code
	a.ExceptionTable = exceptionTable
	return nil
}
