// CompositeEntry 单元测试（M02 验收用例 4，错误约定见 entry.go 与本包各文件 doc）。
package classpath

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 用例 4：依次尝试子 Entry，第 2 个命中 → 返回字节，hit 透传为命中的子 Entry。
func TestCompositeEntry_SecondHit(t *testing.T) {
	sub1 := newDirEntry(t.TempDir()) // 空目录，必 miss
	dir2 := t.TempDir()
	writeFakeClass(t, dir2, "Hello")
	sub2 := newDirEntry(dir2)

	c := CompositeEntry{sub1, sub2}
	data, hit, err := c.ReadClass("Hello")
	if err != nil {
		t.Fatalf("ReadClass: %v", err)
	}
	if !bytes.Equal(data, cafebabe) {
		t.Errorf("data = % X, want CA FE BA BE", data)
	}
	if hit != Entry(sub2) {
		t.Errorf("hit = %v, want 第 2 段的 *DirEntry 自身（原样透传）", hit)
	}
}

// 多段混合目录与 jar：jar 段命中时字节与 hit 都来自 ZipEntry。
func TestCompositeEntry_DirAndJar(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "test.jar")
	inJar := []byte{0x11, 0x22, 0x33, 0x44}
	writeTestJar(t, jar, map[string][]byte{"java/lang/Object.class": inJar})
	subDir := newDirEntry(t.TempDir())
	subJar := newZipEntry(jar)
	// Windows 下句柄不释放 t.TempDir 删不掉 jar，测试结束前手动释放。
	t.Cleanup(func() {
		if subJar.reader != nil {
			subJar.reader.Close()
		}
	})

	c := CompositeEntry{subDir, subJar}
	data, hit, err := c.ReadClass("java/lang/Object")
	if err != nil {
		t.Fatalf("ReadClass: %v", err)
	}
	if !bytes.Equal(data, inJar) {
		t.Errorf("data = % X, want % X", data, inJar)
	}
	if hit != Entry(subJar) {
		t.Errorf("hit = %v, want jar 段的 *ZipEntry 自身", hit)
	}
}

// 真实 IO 错误（jar 损坏）必须立即中断并上抛，不能跳过继续找后面的段
// ——否则损坏 jar 会被静默略过，错误分流约定（entry.go）就失效了。
func TestCompositeEntry_IoErrorAborts(t *testing.T) {
	corrupt := filepath.Join(t.TempDir(), "corrupt.jar")
	if err := os.WriteFile(corrupt, []byte("not a zip file"), 0o644); err != nil {
		t.Fatal(err)
	}
	goodDir := t.TempDir()
	writeFakeClass(t, goodDir, "Hello")

	c := CompositeEntry{newZipEntry(corrupt), newDirEntry(goodDir)}
	_, _, err := c.ReadClass("Hello")
	if err == nil {
		t.Fatal("首段损坏 jar 应以真实 IO 错误中断查找, want error")
	}
	if errors.Is(err, ErrClassNotFound) {
		t.Errorf("真实 IO 错误不应包装/吞成 ErrClassNotFound, got %v", err)
	}
}

// 全部 miss → ErrClassNotFound（errors.Is 判定），错误信息聚合 className
// 与各段路径；错误时没有命中的 Entry。
func TestCompositeEntry_AllMiss(t *testing.T) {
	dir1, dir2 := t.TempDir(), t.TempDir()
	c := CompositeEntry{newDirEntry(dir1), newDirEntry(dir2)}

	_, hit, err := c.ReadClass("no/such/Class")
	if !errors.Is(err, ErrClassNotFound) {
		t.Fatalf("want ErrClassNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "no/such/Class") ||
		!strings.Contains(err.Error(), dir1) || !strings.Contains(err.Error(), dir2) {
		t.Errorf("err = %q, want 聚合 className 与各段路径 %q / %q", err, dir1, dir2)
	}
	if hit != nil {
		t.Errorf("错误时 hit = %v, want nil", hit)
	}
}

// 空 composite（如通配符目录没有任何 jar）→ ErrClassNotFound，不 panic。
func TestCompositeEntry_Empty(t *testing.T) {
	var c CompositeEntry
	_, hit, err := c.ReadClass("Hello")
	if !errors.Is(err, ErrClassNotFound) {
		t.Fatalf("want ErrClassNotFound, got %v", err)
	}
	if hit != nil {
		t.Errorf("错误时 hit = %v, want nil", hit)
	}
}

// String 用 ListSeparator 连接各子 Entry 的 String()；空 composite 为空串。
func TestCompositeEntry_String(t *testing.T) {
	dir1, dir2 := t.TempDir(), t.TempDir()
	c := CompositeEntry{newDirEntry(dir1), newDirEntry(dir2)}
	want := dir1 + string(filepath.ListSeparator) + dir2
	if got := c.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	var empty CompositeEntry
	if got := empty.String(); got != "" {
		t.Errorf("empty String() = %q, want \"\"", got)
	}
}
