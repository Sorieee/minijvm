// WildcardEntry：通配符目录展开（M02 任务 2，对应书 2.6 节）。
package classpath

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
//   - path 形如 "<dir>/*" 或 "<dir>\*"：去掉结尾的 '*' 得到 baseDir
//     （尾部残留的分隔符由 Join/ReadDir 自然消化）
//   - os.ReadDir 全量列出 baseDir（不递归，子目录跳过）；文件名 ToLower 后
//     判断 .jar/.zip 后缀，兜住 Windows 下 .JAR/.Zip 等大小写变体
//     （filepath.Glob 的 pattern 匹配是字节级大小写敏感的，不满足此要求）
//   - 命中文件按文件名排序（ReadDir 本身按文件名有序，显式排序把这个
//     契约钉在展开处），逐个构造 ZipEntry 组成 WildcardEntry
//   - 目录不存在/不可读 → 空 Entry 列表（不报错，
//     ReadClass 自然返回类未找到）
func newWildcardEntry(path string) WildcardEntry {
	baseDir := strings.TrimSuffix(path, "*")
	dirEntries, err := os.ReadDir(baseDir)
	if err != nil {
		// 目录不存在是类路径上允许的空段（与 java 一致）：空展开，不报错。
		return WildcardEntry(nil)
	}
	var matches []string
	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		lower := strings.ToLower(de.Name())
		if strings.HasSuffix(lower, ".jar") || strings.HasSuffix(lower, ".zip") {
			matches = append(matches, filepath.Join(baseDir, de.Name()))
		}
	}
	sort.Strings(matches)
	return WildcardEntry(newCompositeEntryByStrings(matches))
}

// ReadClass 直接委托内层 CompositeEntry。
func (e WildcardEntry) ReadClass(className string) ([]byte, Entry, error) {
	return (CompositeEntry(e)).ReadClass(className)
}

// String 直接委托内层 CompositeEntry。
func (e WildcardEntry) String() string {
	return (CompositeEntry(e)).String()
}
