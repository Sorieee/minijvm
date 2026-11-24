// M03 验收测试：用 javap -v -p internal/classfile/testdata/Meta.class
// 记录的期望值对拍 Parse 结果（用例见 docs/milestones/M03-classfile.md）。
//
// 测试夹具用运行时读取而非 go:embed：testdata/ 被 .gitignore 排除，
// embed 会在目录不存在时直接编译失败；这里是运行时读文件，缺夹具时
// 给出明确提示「先跑 ./scripts/build-testdata.sh」。
package classfile

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// metaClassData 读取构建脚本生成的 v52 测试类；不得修改 CLI 的工作目录。
func metaClassData(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "Meta.class"))
	if err != nil {
		t.Fatalf("读取 testdata/Meta.class 失败；请先运行 ./scripts/build-testdata.sh: %v", err)
	}
	return data
}

// mustParse 在 M03 实现期间把未完成的 panic 转成清晰的 RED 失败，
// 实现完成后则正常返回解析结果。
func mustParse(t *testing.T, data []byte) *ClassFile {
	t.Helper()
	var cf *ClassFile
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Parse panic: %v（若仍为 TODO(M03) panic 属预期 RED）", r)
			}
		}()
		cf, err = Parse(data)
	}()
	if err != nil {
		t.Fatalf("Parse 返回错误: %v", err)
	}
	if cf == nil {
		t.Fatal("Parse 返回 nil ClassFile")
	}
	return cf
}

func mustParseMeta(t *testing.T) *ClassFile {
	t.Helper()
	return mustParse(t, metaClassData(t))
}

// parseWantError 校验坏 class 文件返回包装 ErrClassFormat 的错误。
func parseWantError(t *testing.T, data []byte) error {
	t.Helper()
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Parse panic: %v（若仍为 TODO(M03) panic 属预期 RED）", r)
			}
		}()
		_, err = Parse(data)
	}()
	if err == nil {
		t.Fatal("Parse 返回 nil error，want ErrClassFormat")
	}
	if !errors.Is(err, ErrClassFormat) {
		t.Errorf("Parse error = %v, want errors.Is(ErrClassFormat)", err)
	}
	return err
}

func findField(cf *ClassFile, name string) *MemberInfo {
	for i := range cf.Fields {
		if cf.Fields[i].Name == name {
			return &cf.Fields[i]
		}
	}
	return nil
}

func findMethod(cf *ClassFile, name, descriptor string) *MemberInfo {
	for i := range cf.Methods {
		if cf.Methods[i].Name == name && cf.Methods[i].Descriptor == descriptor {
			return &cf.Methods[i]
		}
	}
	return nil
}

func findAttribute(attrs []AttributeInfo, name string) AttributeInfo {
	for _, a := range attrs {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// 用例 1：Parse 成功，顶层结构与 javap 输出的计数/索引一致。
func TestParseMetaStructure(t *testing.T) {
	cf := mustParseMeta(t)

	if cf.MinorVersion != 0 || cf.MajorVersion != 52 {
		t.Errorf("version = %d.%d, want 0.52", cf.MinorVersion, cf.MajorVersion)
	}
	if cf.AccessFlags != 0x0021 { // ACC_PUBLIC | ACC_SUPER
		t.Errorf("AccessFlags = %#x, want 0x21", cf.AccessFlags)
	}
	if cf.ThisClass != 8 || cf.SuperClass != 2 {
		t.Errorf("this/super = %d/%d, want 8/2 (javap #8 / #2)", cf.ThisClass, cf.SuperClass)
	}
	if len(cf.Interfaces) != 2 || cf.Interfaces[0] != 46 || cf.Interfaces[1] != 48 {
		t.Errorf("Interfaces = %v, want [46 48]", cf.Interfaces)
	}
	if len(cf.Fields) != 6 || len(cf.Methods) != 8 || len(cf.Attributes) != 2 {
		t.Errorf("fields/methods/attrs = %d/%d/%d, want 6/8/2",
			len(cf.Fields), len(cf.Methods), len(cf.Attributes))
	}
}

// 用例 2：magic 被破坏必须报 ClassFormatError。
func TestParseRejectsBadMagic(t *testing.T) {
	data := metaClassData(t)
	bad := append([]byte(nil), data...)
	bad[0] ^= 0xFF // 0xCA → 0x35
	parseWantError(t, bad)
}

// 附加：major > 52 的 class 不在本项目支持范围内。
func TestParseRejectsUnsupportedVersion(t *testing.T) {
	data := metaClassData(t)
	bad := append([]byte(nil), data...)
	bad[6], bad[7] = 0x00, 53
	parseWantError(t, bad)
}

// 用例 4 + 9：类名、父类名、接口名（内部形式，斜杠分隔）。
func TestParseMetaNames(t *testing.T) {
	cf := mustParseMeta(t)

	if got := cf.ClassName(); got != "Meta" {
		t.Errorf("ClassName() = %q, want %q", got, "Meta")
	}
	if got := cf.SuperClassName(); got != "java/lang/Object" {
		t.Errorf("SuperClassName() = %q, want %q", got, "java/lang/Object")
	}
	want := []string{"java/lang/Runnable", "java/lang/Comparable"}
	if got := cf.InterfaceNames(); !reflect.DeepEqual(got, want) {
		t.Errorf("InterfaceNames() = %v, want %v", got, want)
	}
}

// 用例 5：Long/Double 后的槽位编号与 javap 的 #n 对齐（第二个槽是 nil）。
func TestConstantPoolSlotIndexing(t *testing.T) {
	cf := mustParseMeta(t)
	cp := cf.ConstantPool

	if len(cp) != 81 { // javap: #1..#80，下标 0 留空
		t.Fatalf("len(ConstantPool) = %d, want 81", len(cp))
	}
	longInfo, ok := cp.GetConstantInfo(55).(*ConstantLongInfo)
	if !ok || longInfo.Value != 1234567890123 {
		t.Errorf("cp[55] = %#v, want Long(1234567890123)", cp.GetConstantInfo(55))
	}
	if got := cp.GetConstantInfo(56); got != nil {
		t.Errorf("cp[56] = %#v, want nil（long 的第二槽）", got)
	}
	doubleInfo, ok := cp.GetConstantInfo(59).(*ConstantDoubleInfo)
	if !ok || doubleInfo.Value != 3.14159 {
		t.Errorf("cp[59] = %#v, want Double(3.14159)", cp.GetConstantInfo(59))
	}
	if got := cp.GetConstantInfo(60); got != nil {
		t.Errorf("cp[60] = %#v, want nil（double 的第二槽）", got)
	}
	// 若按「第 n 个条目」而不是槽位索引，下面几项必错。
	if got := cp.GetUtf8(57); got != "DOUBLE_CONST" {
		t.Errorf("cp.GetUtf8(57) = %q, want DOUBLE_CONST", got)
	}
	if got := cp.GetClassName(8); got != "Meta" {
		t.Errorf("cp.GetClassName(8) = %q, want Meta", got)
	}
	if got := cp.GetUtf8(80); got != "Meta.java" {
		t.Errorf("cp.GetUtf8(80) = %q, want Meta.java", got)
	}
}

// 用例 7：static final 字段的 ConstantValue 属性。
func TestConstantValueAttributes(t *testing.T) {
	cf := mustParseMeta(t)
	cp := cf.ConstantPool

	cases := []struct {
		name string
		want any
	}{
		{"INT_CONST", int32(42)},
		{"LONG_CONST", int64(1234567890123)},
		{"DOUBLE_CONST", float64(3.14159)},
	}
	for _, c := range cases {
		field := findField(cf, c.name)
		if field == nil {
			t.Fatalf("未找到字段 %s", c.name)
		}
		attr := field.ConstantValueAttribute()
		if attr == nil {
			t.Errorf("%s 没有 ConstantValue 属性", c.name)
			continue
		}
		if got := cp.GetConstantValue(attr.ConstantValueIndex); got != c.want {
			t.Errorf("%s ConstantValue = %#v, want %#v", c.name, got, c.want)
		}
	}

	// String 常量指向 CONSTANT_String，其内容经 StringIndex 取 Utf8。
	field := findField(cf, "STRING_CONST")
	if field == nil {
		t.Fatal("未找到字段 STRING_CONST")
	}
	attr := field.ConstantValueAttribute()
	if attr == nil {
		t.Fatal("STRING_CONST 没有 ConstantValue 属性")
	}
	str, ok := cp.GetConstantInfo(attr.ConstantValueIndex).(*ConstantStringInfo)
	if !ok {
		t.Fatalf("STRING_CONST cp[%d] = %#v, want ConstantStringInfo",
			attr.ConstantValueIndex, cp.GetConstantInfo(attr.ConstantValueIndex))
	}
	if got := cp.GetUtf8(str.StringIndex); got != "hello" {
		t.Errorf("STRING_CONST = %q, want hello", got)
	}

	// 非 static final 字段不应有 ConstantValue。
	if a := findField(cf, "instanceField").ConstantValueAttribute(); a != nil {
		t.Errorf("instanceField 不应有 ConstantValue 属性，得到 %#v", a)
	}
}

// 用例 6：Code 属性（maxStack/maxLocals/code/LineNumberTable）。
func TestCodeAttributeStaticMethod(t *testing.T) {
	cf := mustParseMeta(t)
	method := findMethod(cf, "staticMethod", "(II)I")
	if method == nil {
		t.Fatal("未找到方法 staticMethod(II)I")
	}
	code := method.CodeAttribute()
	if code == nil {
		t.Fatal("staticMethod 没有 Code 属性")
	}
	if code.MaxStack != 2 || code.MaxLocals != 2 {
		t.Errorf("maxStack/maxLocals = %d/%d, want 2/2", code.MaxStack, code.MaxLocals)
	}
	// javap: iload_0(1a) iload_1(1b) iadd(60) ireturn(ac)
	wantCode := []byte{0x1a, 0x1b, 0x60, 0xac}
	if !reflect.DeepEqual(code.Code, wantCode) {
		t.Errorf("Code = % x, want % x", code.Code, wantCode)
	}
	wantLines := []LineNumberTableEntry{{StartPc: 0, LineNumber: 56}}
	if !reflect.DeepEqual(code.LineNumberTable, wantLines) {
		t.Errorf("LineNumberTable = %v, want %v", code.LineNumberTable, wantLines)
	}
	if len(code.ExceptionTable) != 0 {
		t.Errorf("ExceptionTable = %v, want 空（throws 子句不是 try/catch）", code.ExceptionTable)
	}
}

// Code 属性的边界：main 与构造器的栈深度/字节码长度/行号表。
func TestCodeAttributeMainAndConstructor(t *testing.T) {
	cf := mustParseMeta(t)

	main := findMethod(cf, "main", "([Ljava/lang/String;)V")
	if main == nil {
		t.Fatal("未找到 main([Ljava/lang/String;)V")
	}
	mainCode := main.CodeAttribute()
	if mainCode == nil {
		t.Fatal("main 没有 Code 属性")
	}
	if mainCode.MaxStack != 2 || mainCode.MaxLocals != 1 || len(mainCode.Code) != 9 {
		t.Errorf("main code: stack=%d locals=%d len=%d, want 2/1/9",
			mainCode.MaxStack, mainCode.MaxLocals, len(mainCode.Code))
	}
	wantMainLines := []LineNumberTableEntry{{StartPc: 0, LineNumber: 64}, {StartPc: 8, LineNumber: 65}}
	if !reflect.DeepEqual(mainCode.LineNumberTable, wantMainLines) {
		t.Errorf("main LineNumberTable = %v, want %v", mainCode.LineNumberTable, wantMainLines)
	}

	init := findMethod(cf, "<init>", "(Ljava/lang/String;)V")
	if init == nil {
		t.Fatal("未找到 <init>(Ljava/lang/String;)V")
	}
	initCode := init.CodeAttribute()
	if initCode == nil {
		t.Fatal("<init> 没有 Code 属性")
	}
	if initCode.MaxStack != 2 || initCode.MaxLocals != 2 || len(initCode.Code) != 15 {
		t.Errorf("<init> code: stack=%d locals=%d len=%d, want 2/2/15",
			initCode.MaxStack, initCode.MaxLocals, len(initCode.Code))
	}
	wantInitLines := []LineNumberTableEntry{
		{StartPc: 0, LineNumber: 30}, {StartPc: 4, LineNumber: 31},
		{StartPc: 9, LineNumber: 32}, {StartPc: 14, LineNumber: 33},
	}
	if !reflect.DeepEqual(initCode.LineNumberTable, wantInitLines) {
		t.Errorf("<init> LineNumberTable = %v, want %v", initCode.LineNumberTable, wantInitLines)
	}
}

// 用例 8：Exceptions 属性（throws java.io.IOException）。
func TestExceptionsAttribute(t *testing.T) {
	cf := mustParseMeta(t)

	method := findMethod(cf, "throwIOException", "()V")
	if method == nil {
		t.Fatal("未找到 throwIOException()V")
	}
	attr := method.ExceptionsAttribute()
	if attr == nil {
		t.Fatal("throwIOException 没有 Exceptions 属性")
	}
	if len(attr.ExceptionIndexTable) != 1 {
		t.Fatalf("ExceptionIndexTable = %v, want 1 项", attr.ExceptionIndexTable)
	}
	if got := cf.ConstantPool.GetClassName(attr.ExceptionIndexTable[0]); got != "java/io/IOException" {
		t.Errorf("异常类名 = %q, want java/io/IOException", got)
	}
	if a := findMethod(cf, "run", "()V").ExceptionsAttribute(); a != nil {
		t.Errorf("run() 不应有 Exceptions 属性，得到 %#v", a)
	}
}

// 用例 4（附加）：SourceFile 与未知属性 Signature 跳过。
func TestSourceFileAndUnknownAttributeSkipped(t *testing.T) {
	cf := mustParseMeta(t)

	attr := findAttribute(cf.Attributes, "SourceFile")
	if attr == nil {
		t.Fatal("类级属性里没有 SourceFile")
	}
	source, ok := attr.(*SourceFileAttribute)
	if !ok {
		t.Fatalf("SourceFile 属性类型 = %T, want *SourceFileAttribute", attr)
	}
	if got := cf.ConstantPool.GetUtf8(source.SourceFileIndex); got != "Meta.java" {
		t.Errorf("SourceFile = %q, want Meta.java", got)
	}

	// Signature 属性（泛型）不在 M03 支持列表，必须按 u4 长度安全跳过。
	unknown := findAttribute(cf.Attributes, "Signature")
	if unknown == nil {
		t.Fatal("类级属性里没有 Signature")
	}
	if _, ok := unknown.(*UnknownAttribute); !ok {
		t.Errorf("Signature 属性类型 = %T, want *UnknownAttribute", unknown)
	}
	if unknown.Length() != 2 { // #78 = Utf8 索引，2 字节
		t.Errorf("Signature.Length() = %d, want 2", unknown.Length())
	}
}
