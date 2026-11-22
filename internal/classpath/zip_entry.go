// ZipEntry：.jar / .zip 类路径条目（M02 任务 2，对应书 2.4 节）。
package classpath

import (
	"archive/zip"
	"path/filepath"
)

// ZipEntry 从 jar/zip 包内读取 class 文件。
//
// M02 设计决策（一步到位，书到 M14 才做此优化）：首次读取时打开 zip 并缓存
// central directory 的条目索引，避免每次 ReadClass 都重新解析（rt.jar 约两万条目，
// 查找频繁）。minijvm 单线程，懒加载无需加锁；*zip.ReadCloser 打开后常驻，
// 进程生命周期内不关闭（玩具 JVM，由 OS 回收）。
type ZipEntry struct {
	zipName string // jar/zip 文件绝对路径

	// reader 与 entries 均为懒加载：首次 ReadClass 时填充（见 ensureOpen）。
	reader  *zip.ReadCloser      // 打开的 zip；nil 表示尚未打开
	entries map[string]*zip.File // 条目名 → zip.File 索引，如 "java/lang/Object.class"
}

// newZipEntry 构造 zip Entry，只记录路径，不立刻打开文件。
//
// zip 打不开（不存在/损坏）在 ReadClass → ensureOpen 时才暴露；
// 属真实 IO 错误，不包装 ErrClassNotFound，原样返回。
func newZipEntry(path string) *ZipEntry {
	absPath, err := filepath.Abs(path)
	if err != nil {
		// 仅在 path 是相对路径且连工作目录都取不到时发生，极罕见。
		// 签名不返回 error，按书的方式 panic（进程早死好过静默错路径）。
		panic(err)
	}
	return &ZipEntry{zipName: absPath}
}

// ensureOpen 懒加载：首次调用时 zip.OpenReader 并把 reader.File 建成 entries 索引；
// 之后为空操作。ReadClass 开头先调用它。
func (e *ZipEntry) ensureOpen() error {
	if e.reader != nil {
		return nil
	}
	classReader, err := zip.OpenReader(e.zipName)
	if err != nil {
		return err
	}
	entries := make(map[string]*zip.File)
	for _, f := range classReader.File {
		entries[f.Name] = f
	}
	e.entries = entries
	e.reader = classReader
	return nil
}

// ReadClass 读取包内 <className>.class 条目的字节。
//
// 实现要点：
//   - 先 ensureOpen
//   - entries 里查 className+".class"：查不到 → 包装 ErrClassNotFound（附 zipName）
//   - 命中则 f.Open() 后 io.ReadAll（zip.File 每次调用 Open 返回新 reader，
//     同一条目可重复读）
func (e *ZipEntry) ReadClass(className string) ([]byte, Entry, error) {
	panic("TODO(M02): not implemented")
}

// String 返回 jar/zip 文件路径。
func (e *ZipEntry) String() string {
	return e.zipName
}
