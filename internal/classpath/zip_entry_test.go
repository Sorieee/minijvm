// ZipEntry 单元测试（M02 验收用例 3，错误约定见 entry.go）。
package classpath

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestJar 用 zip 库造一个测试 jar：entries 为 条目名 → 内容。
func writeTestJar(t *testing.T, path string, entries map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for name, data := range entries {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

// 用例 3：读 jar 内存在的类 → 返回字节（首 4 字节 CA FE BA BE），
// 命中 Entry 是自身；包路径条目读对内容；同一条目可重复读。
func TestZipEntry_ReadClass(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "test.jar")
	// java/lang/Object.class 故意用与 Hello.class 不同的内容，
	// 验证包路径 '/' 条目名命中且没有读错条目。
	objectBytes := []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x01}
	writeTestJar(t, jar, map[string][]byte{
		"Hello.class":            cafebabe,
		"java/lang/Object.class": objectBytes,
	})

	e := newZipEntry(jar)
	// ZipEntry 按设计进程生命周期内不关 reader；Windows 下句柄不释放
	// t.TempDir 就删不掉 jar，测试结束前手动释放（cleanups 后注册先执行）。
	t.Cleanup(func() {
		if e.reader != nil {
			e.reader.Close()
		}
	})

	// 懒加载：构造时不打开文件，首次 ReadClass 后 reader/entries 才就位。
	if e.reader != nil || e.entries != nil {
		t.Errorf("构造后不应打开 zip: reader=%v entries=%v", e.reader, e.entries)
	}

	for className, want := range map[string][]byte{
		"Hello":            cafebabe,
		"java/lang/Object": objectBytes,
	} {
		data, hit, err := e.ReadClass(className)
		if err != nil {
			t.Fatalf("ReadClass(%q): %v", className, err)
		}
		if !bytes.Equal(data, want) {
			t.Errorf("ReadClass(%q) = % X, want % X", className, data, want)
		}
		if hit != Entry(e) {
			t.Errorf("ReadClass(%q) hit = %v, want the entry itself", className, hit)
		}
	}
	if e.reader == nil || e.entries == nil {
		t.Errorf("首次 ReadClass 后应完成懒加载: reader=%v entries=%v", e.reader, e.entries)
	}

	// zip.File 每次 Open 返回新 reader：同一条目再次读取结果一致。
	data2, _, err := e.ReadClass("Hello")
	if err != nil {
		t.Fatalf("重复 ReadClass(Hello): %v", err)
	}
	if !bytes.Equal(data2, cafebabe) {
		t.Errorf("重复 ReadClass(Hello) = % X, want CA FE BA BE", data2)
	}
}

// 用例 2 扩展：读不存在的类 → ErrClassNotFound（errors.Is 可判定），
// 且按 entry.go 约定包装了 className 与 jar 路径上下文；错误时没有命中的 Entry。
func TestZipEntry_ClassNotFound(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "test.jar")
	writeTestJar(t, jar, map[string][]byte{"Hello.class": cafebabe})
	e := newZipEntry(jar)
	t.Cleanup(func() {
		if e.reader != nil {
			e.reader.Close()
		}
	})

	_, hit, err := e.ReadClass("no/such/Class")
	if !errors.Is(err, ErrClassNotFound) {
		t.Fatalf("want ErrClassNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "no/such/Class") || !strings.Contains(err.Error(), jar) {
		t.Errorf("err = %q, want 上下文含 className %q 与 jar 路径 %q", err, "no/such/Class", jar)
	}
	if hit != nil {
		t.Errorf("错误时 hit = %v, want nil（契约：只返回实际命中的 Entry）", hit)
	}
}

// 真实 IO 错误（jar 不存在 / 内容损坏）不包装 ErrClassNotFound，原样上抛；
// 构造时不打开文件，坏路径到 ReadClass 才暴露。
func TestZipEntry_IoErrorNotWrapped(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.jar")
	e := newZipEntry(missing)
	if got := e.String(); got != missing {
		t.Errorf("String() = %q, want %q", got, missing)
	}

	_, hit, err := e.ReadClass("Hello")
	if err == nil {
		t.Fatal("jar 不存在, want error")
	}
	if errors.Is(err, ErrClassNotFound) {
		t.Errorf("真实 IO 错误不应包装 ErrClassNotFound, got %v", err)
	}
	if hit != nil {
		t.Errorf("错误时 hit = %v, want nil", hit)
	}

	corrupt := filepath.Join(t.TempDir(), "corrupt.jar")
	if err := os.WriteFile(corrupt, []byte("not a zip file"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, hit, err = newZipEntry(corrupt).ReadClass("Hello")
	if err == nil {
		t.Fatal("jar 损坏, want error")
	}
	if errors.Is(err, ErrClassNotFound) {
		t.Errorf("真实 IO 错误不应包装 ErrClassNotFound, got %v", err)
	}
	if hit != nil {
		t.Errorf("错误时 hit = %v, want nil", hit)
	}
}

// String 返回构造时规范化的绝对路径。
func TestZipEntry_String(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "test.jar")
	if got := newZipEntry(jar).String(); got != jar {
		t.Errorf("String() = %q, want %q", got, jar)
	}
	// 相对路径应被 Abs 规范化：结果必须是绝对路径。
	if got := newZipEntry(".").String(); !filepath.IsAbs(got) {
		t.Errorf("String() = %q, want absolute path", got)
	}
}
