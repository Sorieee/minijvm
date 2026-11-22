// DirEntry 单元测试（M02 验收用例 1、2）。
package classpath

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// cafebabe 是 class 文件魔数，M03 才真正解析内容，这里只比对字节。
var cafebabe = []byte{0xCA, 0xFE, 0xBA, 0xBE}

// 写一个假 class 文件到 dir 下，返回相对 className。
func writeFakeClass(t *testing.T, dir, className string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(className)+".class")
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, cafebabe, 0o644); err != nil {
		t.Fatal(err)
	}
}

// 用例 1：读存在的类 → 返回字节（首 4 字节 CA FE BA BE），命中 Entry 是自身。
func TestDirEntry_ReadClass(t *testing.T) {
	dir := t.TempDir()
	writeFakeClass(t, dir, "Hello")
	writeFakeClass(t, dir, "java/lang/Object") // 验证包路径 '/' 在 Windows 下的拼接

	e := newDirEntry(dir)

	for _, className := range []string{"Hello", "java/lang/Object"} {
		data, hit, err := e.ReadClass(className)
		if err != nil {
			t.Fatalf("ReadClass(%q): %v", className, err)
		}
		if !bytes.Equal(data, cafebabe) {
			t.Errorf("ReadClass(%q) = % X, want CA FE BA BE", className, data)
		}
		if hit != Entry(e) {
			t.Errorf("ReadClass(%q) hit = %v, want the entry itself", className, hit)
		}
	}
}

// 用例 2：读不存在的类 → ErrClassNotFound（可被 errors.Is 判定）。
func TestDirEntry_ClassNotFound(t *testing.T) {
	e := newDirEntry(t.TempDir())

	_, _, err := e.ReadClass("no/such/Class")
	if !errors.Is(err, ErrClassNotFound) {
		t.Fatalf("want ErrClassNotFound, got %v", err)
	}
}

// String 返回构造时规范化的绝对路径。
func TestDirEntry_String(t *testing.T) {
	dir := t.TempDir()
	if got := newDirEntry(dir).String(); got != dir {
		t.Errorf("String() = %q, want %q", got, dir)
	}
	// 相对路径应被 Abs 规范化：结果必须是绝对路径。
	if got := newDirEntry(".").String(); !filepath.IsAbs(got) {
		t.Errorf("String() = %q, want absolute path", got)
	}
}
