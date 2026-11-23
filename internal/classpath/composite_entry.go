// CompositeEntry：多段类路径的组合（M02 任务 2，对应书 2.5 节）。
package classpath

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// CompositeEntry 依序尝试若干子 Entry，对应 -cp 里 ListSeparator 分隔的路径列表。
type CompositeEntry []Entry // 保持构造时的顺序，顺序即查找优先级

func newCompositeEntryByStrings(paths []string) CompositeEntry {
	var entries []Entry
	for _, path := range paths {
		entries = append(entries, newSingleEntry(path))
	}
	return entries
}

// ReadClass 依序尝试子 Entry，第一个命中即返回（子 Entry 给的命中 Entry 原样上抛）。
//
// 实现要点：子 Entry 的真实 IO 错误立即上抛（不继续尝试后续段）；
// 全部 miss 时返回统一的 ErrClassNotFound 包装错误（聚合 className 与各段路径）。
func (e CompositeEntry) ReadClass(className string) ([]byte, Entry, error) {
	for _, sub := range e {
		data, hit, err := sub.ReadClass(className)
		if err == nil {
			return data, hit, nil
		}
		if !errors.Is(err, ErrClassNotFound) {
			// 真实 IO 错误（jar 损坏、权限等）立即上抛，不伪装成类未找到。
			return nil, nil, err
		}
		// 类未命中：尝试下一段。
	}
	return nil, nil, fmt.Errorf("%w: %s in %s", ErrClassNotFound, className, e.String())
}

// String 用 filepath.ListSeparator 连接各子 Entry 的 String()。
func (e CompositeEntry) String() string {
	strs := []string{}
	for _, entry := range e {
		strs = append(strs, entry.String())
	}
	return strings.Join(strs, string(filepath.ListSeparator))
}
