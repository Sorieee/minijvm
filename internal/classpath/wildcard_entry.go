// WildcardEntry：通配符目录展开（M02 任务 2，对应书 2.6 节）。
package classpath

// WildcardEntry 表示以 '*' 结尾的通配符目录（如 jre 的 lib/*）。
//
// 语义与 java 一致：只匹配 baseDir 下（不递归子目录）的所有 .jar / .zip，
// 该目录下的 .class 文件与子目录不参与匹配。
//
// 底层复用 CompositeEntry；与书的差异：展开时按文件名排序，
// 保证查找顺序确定（多次运行一致，测试可复现）。
type WildcardEntry CompositeEntry // baseDir 下排序后的全部 jar/zip

// newWildcardEntry 构造通配符 Entry。
//
// 实现要点：
//   - path 形如 "<dir>/*" 或 "<dir>\*"：去掉结尾的 '*'（以及多余的路径分隔符）
//     得到 baseDir
//   - 用 filepath.Glob 匹配 baseDir 下的 *.jar 与 *.zip；Windows 文件名
//     不区分大小写，需兜住 .JAR/.Zip（Glob 大小写敏感，可全量列目录后
//     strings.ToLower 判断后缀）
//   - 每个命中文件构造 *ZipEntry，按文件名排序后组成 WildcardEntry
//   - 目录不存在 / 一个 jar 都没有 → 空 Entry 列表（不报错，
//     ReadClass 自然返回类未找到）
func newWildcardEntry(path string) WildcardEntry {
	panic("TODO(M02): not implemented")
}

// ReadClass 直接委托内层 CompositeEntry。
func (e WildcardEntry) ReadClass(className string) ([]byte, Entry, error) {
	panic("TODO(M02): not implemented")
}

// String 直接委托内层 CompositeEntry。
func (e WildcardEntry) String() string {
	panic("TODO(M02): not implemented")
}
