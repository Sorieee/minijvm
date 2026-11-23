// 常量池条目接口与 tag 分派（M03 任务 2，对应书 3.2–3.4 节）。
package classfile

import "fmt"

// 常量池 tag 常量（JVM 规范 §4.4-A）。
//
// 共 17 种：tag 2 / 13 / 14 保留未用。v52（Java 8）实际会出现的只有
// 其中一部分；Dynamic/Module/Package 是 v55+ 才有，解析器支持它们
// 但在 v52 class 文件里不会出现。
const (
	ConstantTagUtf8               uint8 = 1
	ConstantTagInteger            uint8 = 3
	ConstantTagFloat              uint8 = 4
	ConstantTagLong               uint8 = 5
	ConstantTagDouble             uint8 = 6
	ConstantTagClass              uint8 = 7
	ConstantTagString             uint8 = 8
	ConstantTagFieldref           uint8 = 9
	ConstantTagMethodref          uint8 = 10
	ConstantTagInterfaceMethodref uint8 = 11
	ConstantTagNameAndType        uint8 = 12
	ConstantTagMethodHandle       uint8 = 15
	ConstantTagMethodType         uint8 = 16
	ConstantTagDynamic            uint8 = 17 // v55+
	ConstantTagInvokeDynamic      uint8 = 18
	ConstantTagModule             uint8 = 19 // v55+
	ConstantTagPackage            uint8 = 20 // v55+
)

// ConstantInfo 是常量池条目的统一接口。
//
// 接口带未导出方法 readInfo：这是 Go 里「封闭接口」的做法——只有本包
// 能实现它，防止外部误造常量条目（也就保证了 tag ↔ 类型的对应关系
// 不会被人为破坏）。
type ConstantInfo interface {
	// Tag 返回常量池 tag。
	Tag() uint8
	// readInfo 从读取器读出本条目的内容（调用方已读过 tag）。
	readInfo(r *ClassReader) error
}

// readConstantInfo 按 tag 分派到具体常量类型的解析。
//
// tag 已经由调用方读出（readConstantPool 负责），本函数只负责：
//  1. 按 tag 造出对应的空条目
//  2. 调用它的 readInfo 读内容
//
// 未知 tag 返回包装 ErrClassFormat 的错误（坏文件，不是「类找不到」）。
func readConstantInfo(r *ClassReader, tag uint8) (ConstantInfo, error) {
	var info ConstantInfo
	switch tag {
	case ConstantTagUtf8:
		info = &ConstantUtf8Info{}
	case ConstantTagInteger:
		info = &ConstantIntegerInfo{}
	case ConstantTagFloat:
		info = &ConstantFloatInfo{}
	case ConstantTagLong:
		info = &ConstantLongInfo{}
	case ConstantTagDouble:
		info = &ConstantDoubleInfo{}
	case ConstantTagClass:
		info = &ConstantClassInfo{}
	case ConstantTagString:
		info = &ConstantStringInfo{}
	case ConstantTagFieldref:
		info = &ConstantFieldrefInfo{}
	case ConstantTagMethodref:
		info = &ConstantMethodrefInfo{}
	case ConstantTagInterfaceMethodref:
		info = &ConstantInterfaceMethodrefInfo{}
	case ConstantTagNameAndType:
		info = &ConstantNameAndTypeInfo{}
	case ConstantTagMethodHandle:
		info = &ConstantMethodHandleInfo{}
	case ConstantTagMethodType:
		info = &ConstantMethodTypeInfo{}
	case ConstantTagDynamic:
		info = &ConstantDynamicInfo{}
	case ConstantTagInvokeDynamic:
		info = &ConstantInvokeDynamicInfo{}
	case ConstantTagModule:
		info = &ConstantModuleInfo{}
	case ConstantTagPackage:
		info = &ConstantPackageInfo{}
	default:
		return nil, fmt.Errorf("%w: 未知常量池 tag %d（偏移 %d）",
			ErrClassFormat, tag, r.Pos()-1)
	}
	if err := info.readInfo(r); err != nil {
		return nil, err
	}
	return info, nil
}
