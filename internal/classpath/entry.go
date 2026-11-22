// Entry 抽象与构造分派（M02 任务 1，对应书 2.1–2.2 节）。
package classpath

import "errors"

// ErrClassNotFound 是「类未找到」的统一哨兵错误。
//
// 约定：各 Entry 实现读不到类时，返回用它包装的错误（fmt.Errorf + %w），
// 携带 className 与路径上下文，调用方用 errors.Is(err, ErrClassNotFound) 判定；
// 真实的 IO 错误（权限不足、zip 损坏等）不包装，原样向上传递。
var ErrClassNotFound = errors.New("class not found")

// Entry 表示类路径中的一段，知道如何按类全限定名定位并读取 class 字节。
//
// 四种实现：DirEntry / ZipEntry / CompositeEntry / WildcardEntry。
type Entry interface {
	// ReadClass 读取类字节。
	//
	// className 是全限定名，形如 "java/lang/Object"（斜杠分隔、
	// 不含 .class 后缀）；调用方负责把主类名的 '.' 换成 '/'。
	// 返回：类字节、实际命中的 Entry（用于 Found class / -verbose:class 输出）、错误。
	ReadClass(className string) ([]byte, Entry, error)
	// String 返回该 Entry 的路径表示（目录绝对路径、jar 路径、
	// 或 ListSeparator 连接的组合路径）。
	String() string
}

// NewEntry 解析 -cp / -classpath 的值，构造类路径入口。
//
// 实现要点：按 filepath.ListSeparator（Windows 为 ';'）拆分路径列表，
// 每段交给 newSingleEntry 分派；多段包成 *CompositeEntry，单段直接返回该实现。
// 空字符串按 java 语义处理为当前目录 "."（也可在此处归一化，见 Classpath.Parse）。
func NewEntry(pathList string) Entry {
	panic("TODO(M02): not implemented")
}

// newSingleEntry 把单段路径分派到具体实现：
//
//	以 '*' 结尾（含 "<dir>/*" / "<dir>\*"）→ newWildcardEntry
//	以 .jar / .zip 结尾（ToLowerCase 后判断，兼容 .JAR）→ newZipEntry
//	其余 → newDirEntry
func newSingleEntry(path string) Entry {
	panic("TODO(M02): not implemented")
}
