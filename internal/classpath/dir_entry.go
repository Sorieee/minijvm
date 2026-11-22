// DirEntry：类路径上的普通目录（M02 任务 2，对应书 2.3 节）。
package classpath

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// DirEntry 从本机目录读取 class 文件。
type DirEntry struct {
	absPath string // 构造时用 filepath.Abs 规范化的目录绝对路径，作为 String() 输出
}

// newDirEntry 构造目录 Entry。
//
// 注意：目录不存在不在此报错——留到 ReadClass 时表现为「类找不到」，
// 与 java 命令行为一致（classpath 上允许列不存在的目录）。
func newDirEntry(path string) *DirEntry {
	absDir, err := filepath.Abs(path)
	if err != nil {
		// 仅在 path 是相对路径且连工作目录都取不到时发生，极罕见。
		// 签名不返回 error，按书的方式 panic（进程早死好过静默错路径）。
		panic(err)
	}
	return &DirEntry{absPath: absDir}
}

// ReadClass 读取 <absPath>/<className>.class 的全部字节。
//
// 实现要点：
//   - filepath.Join 拼路径（自动处理 Windows '\' 与 className 中的 '/'）
//   - os.ReadFile 读取；文件不存在（os.IsNotExist / errors.Is(err, fs.ErrNotExist)）
//     时返回 fmt.Errorf 包装的 ErrClassNotFound，附 className 与目录
//   - 命中时第二个返回值是 e 自身（Entry 接口形态）
func (e *DirEntry) ReadClass(className string) ([]byte, Entry, error) {
	s := filepath.Join(e.absPath, className+".class")
	data, err := os.ReadFile(s)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// 文件不存在 = 类未找到：包装哨兵错误，让上层继续尝试下一段。
			return nil, nil, fmt.Errorf("%w: %s in %s", ErrClassNotFound, className, e.absPath)
		}
		// 权限等真实 IO 错误：原样上抛，不伪装成「类未找到」。
		return nil, nil, err
	}
	return data, e, nil
}

// String 返回目录绝对路径。
func (e *DirEntry) String() string {
	return e.absPath
}
