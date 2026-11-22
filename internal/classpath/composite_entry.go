// CompositeEntry：多段类路径的组合（M02 任务 2，对应书 2.5 节）。
package classpath

// CompositeEntry 依序尝试若干子 Entry，对应 -cp 里 ListSeparator 分隔的路径列表。
type CompositeEntry []Entry // 保持构造时的顺序，顺序即查找优先级

// newCompositeEntry 按分隔符拆分路径列表并逐段构造子 Entry。
//
// 供 NewEntry（-cp 多段）与 newWildcardEntry（展开的 jar 列表，直接拼切片）复用。
func newCompositeEntry(pathList string) CompositeEntry {
	panic("TODO(M02): not implemented")
}

// ReadClass 依序尝试子 Entry，第一个命中即返回（子 Entry 给的命中 Entry 原样上抛）。
//
// 实现要点：子 Entry 的真实 IO 错误立即上抛（不继续尝试后续段）；
// 全部 miss 时返回统一的 ErrClassNotFound 包装错误（可聚合各段路径）。
func (e CompositeEntry) ReadClass(className string) ([]byte, Entry, error) {
	panic("TODO(M02): not implemented")
}

// String 用 filepath.ListSeparator 连接各子 Entry 的 String()。
func (e CompositeEntry) String() string {
	panic("TODO(M02): not implemented")
}
