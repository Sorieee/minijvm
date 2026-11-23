// WildcardEntry 单元测试（M02 验收用例 5；排序与大小写约定见 wildcard_entry.go doc）。
package classpath

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

// closeWildcard 在测试结束前释放内层 ZipEntry 的 reader：
// ZipEntry 按设计进程生命周期内不关，Windows 下句柄不释放
// t.TempDir 就删不掉 jar（cleanups 后注册先执行）。
func closeWildcard(t *testing.T, e WildcardEntry) {
	t.Cleanup(func() {
		for _, sub := range e {
			if z, ok := sub.(*ZipEntry); ok && z.reader != nil {
				z.reader.Close()
			}
		}
	})
}

// 用例 5：目录下 2 个 jar，各自的类都能命中，hit 是对应 jar 的 ZipEntry。
func TestWildcardEntry_TwoJars(t *testing.T) {
	dir := t.TempDir()
	aZip := filepath.Join(dir, "a.zip")
	bJar := filepath.Join(dir, "b.jar")
	writeTestJar(t, aZip, map[string][]byte{"A.class": {0xAA}})
	writeTestJar(t, bJar, map[string][]byte{"B.class": {0xBB}})
	e := newWildcardEntry(filepath.Join(dir, "*"))
	closeWildcard(t, e)

	data, hit, err := e.ReadClass("A")
	if err != nil {
		t.Fatalf("ReadClass(A): %v", err)
	}
	if len(data) != 1 || data[0] != 0xAA {
		t.Errorf("ReadClass(A) = % X, want AA", data)
	}
	if z, ok := hit.(*ZipEntry); !ok || z.String() != aZip {
		t.Errorf("hit = %v, want a.zip 的 ZipEntry", hit)
	}

	data, hit, err = e.ReadClass("B")
	if err != nil {
		t.Fatalf("ReadClass(B): %v", err)
	}
	if len(data) != 1 || data[0] != 0xBB {
		t.Errorf("ReadClass(B) = % X, want BB", data)
	}
	if z, ok := hit.(*ZipEntry); !ok || z.String() != bJar {
		t.Errorf("hit = %v, want b.jar 的 ZipEntry", hit)
	}
}

// 展开顺序按文件名排序（与书的差异，见 doc）：同名类取排在前面的 jar 的字节。
// 目录里放 a.zip 与 b.jar：按「jar 先、zip 后」的展开会命中 b.jar（02），
// 按文件名排序则 a.zip 在前，应命中 01。
func TestWildcardEntry_SortedExpansion(t *testing.T) {
	dir := t.TempDir()
	writeTestJar(t, filepath.Join(dir, "a.zip"), map[string][]byte{"Shared.class": {0x01}})
	writeTestJar(t, filepath.Join(dir, "b.jar"), map[string][]byte{"Shared.class": {0x02}})
	e := newWildcardEntry(filepath.Join(dir, "*"))
	closeWildcard(t, e)

	data, _, err := e.ReadClass("Shared")
	if err != nil {
		t.Fatalf("ReadClass(Shared): %v", err)
	}
	if len(data) != 1 || data[0] != 0x01 {
		t.Errorf("ReadClass(Shared) = % X, want 01（a.zip 按文件名排在 b.jar 前）", data)
	}
}

// 后缀大小写：Windows 文件名不区分大小写，.JAR/.Zip 也要展开
// （Glob 的 pattern 匹配是字节级大小写敏感的，须全量列目录后 ToLower 判断）。
func TestWildcardEntry_ExtensionCase(t *testing.T) {
	dir := t.TempDir()
	writeTestJar(t, filepath.Join(dir, "UPPER.JAR"), map[string][]byte{"U.class": cafebabe})
	writeTestJar(t, filepath.Join(dir, "Mixed.Zip"), map[string][]byte{"M.class": {0x55}})
	e := newWildcardEntry(filepath.Join(dir, "*"))
	closeWildcard(t, e)

	if data, _, err := e.ReadClass("U"); err != nil || !bytes.Equal(data, cafebabe) {
		t.Errorf("ReadClass(U) = % X, %v; want CA FE BA BE", data, err)
	}
	if data, _, err := e.ReadClass("M"); err != nil || len(data) != 1 || data[0] != 0x55 {
		t.Errorf("ReadClass(M) = % X, %v; want 55", data, err)
	}
}

// 目录不存在 = 空 Entry 列表，构造不报错；ReadClass 自然返回 ErrClassNotFound。
func TestWildcardEntry_MissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir")
	e := newWildcardEntry(filepath.Join(missing, "*"))

	_, hit, err := e.ReadClass("Hello")
	if !errors.Is(err, ErrClassNotFound) {
		t.Fatalf("want ErrClassNotFound, got %v", err)
	}
	if hit != nil {
		t.Errorf("错误时 hit = %v, want nil", hit)
	}
}

// String 委托内层 Composite：展开后的 jar 列表按排序用 ListSeparator 连接。
func TestWildcardEntry_String(t *testing.T) {
	dir := t.TempDir()
	aZip := filepath.Join(dir, "a.zip")
	bJar := filepath.Join(dir, "b.jar")
	writeTestJar(t, aZip, map[string][]byte{"A.class": {0xAA}})
	writeTestJar(t, bJar, map[string][]byte{"B.class": {0xBB}})
	e := newWildcardEntry(filepath.Join(dir, "*"))
	closeWildcard(t, e)

	want := aZip + string(filepath.ListSeparator) + bJar // 排序后 a 在前
	if got := e.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
