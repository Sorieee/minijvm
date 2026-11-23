// Classpath 三段模型测试（M02 任务 3，对应书 2.7 节）。
//
// 契约：Parse 是对外的命令行入口（jre 定位失败打印错误并退出进程），
// ParseTo 是它的可测试接缝（返回 error，不退出）；行为断言通过
// newTestParse 展开，jre 定位失败路径直接测 ParseTo。
package classpath

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestClass 在 dir 下按类全限定名写一个 .class 文件（自动建子目录）。
func writeTestClass(t *testing.T, dir, className, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(className)+".class")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

// 用例 1：多段 -cp 每段都可达（java 语义：ListSeparator 分隔、依序查找）。
func TestParseMultiSegmentUserClasspath(t *testing.T) {
	root := t.TempDir()
	dirA := filepath.Join(root, "a")
	dirB := filepath.Join(root, "b")
	writeTestClass(t, dirA, "pkg/A", "class-a")
	writeTestClass(t, dirB, "pkg/B", "class-b")

	cp := dirA + string(filepath.ListSeparator) + dirB
	jre := makeFakeJre(t, root)

	parsed := newTestParse(t, jre, cp)

	data, _, err := parsed.ReadClass("pkg/A")
	if err != nil || string(data) != "class-a" {
		t.Fatalf("ReadClass(pkg/A) = (%q, %v), want (\"class-a\", nil)", string(data), err)
	}
	data, _, err = parsed.ReadClass("pkg/B")
	if err != nil || string(data) != "class-b" {
		t.Fatalf("ReadClass(pkg/B) = (%q, %v), want (\"class-b\", nil)", string(data), err)
	}
}

// 用例 2：三段查找顺序 boot → ext → user，boot 优先（系统类永不取自 user -cp）。
func TestParseSearchOrder(t *testing.T) {
	root := t.TempDir()
	jre := makeFakeJre(t, root) // boot 段已含 java/lang/Object.class = "boot"

	// ext: <jre>/ext/*.jar（writeTestJar 不建父目录，ext 需先建）
	if err := os.MkdirAll(filepath.Join(jre, "ext"), 0o755); err != nil {
		t.Fatalf("mkdir ext: %v", err)
	}
	writeTestJar(t, filepath.Join(jre, "ext", "ext.jar"),
		map[string][]byte{
			"javax/ext/Only.class": []byte("ext"),
			// 与 user 段同名，验证 ext 先于 user：
			"pkg/Shadow.class": []byte("ext"),
		})
	// user: 目录，含与 ext 同名的类 + 自己的类 + 与 boot 同名的类
	user := filepath.Join(root, "user")
	writeTestClass(t, user, "pkg/Shadow", "user-should-lose")
	writeTestClass(t, user, "pkg/UserOnly", "user")
	writeTestClass(t, user, "java/lang/Object", "user-should-lose")

	parsed := newTestParse(t, jre, user)

	cases := []struct {
		className string
		want      string
	}{
		{"java/lang/Object", "boot"}, // boot 优先于 user
		{"pkg/Shadow", "ext"},        // ext 优先于 user
		{"javax/ext/Only", "ext"},    // 仅 ext 有
		{"pkg/UserOnly", "user"},     // 仅 user 有
	}
	for _, c := range cases {
		data, _, err := parsed.ReadClass(c.className)
		if err != nil {
			t.Errorf("ReadClass(%s) = err %v, want content %q", c.className, err, c.want)
			continue
		}
		if string(data) != c.want {
			t.Errorf("ReadClass(%s) = %q, want %q", c.className, string(data), c.want)
		}
	}
}

// 用例 3：三段都 miss → ErrClassNotFound 包装错误（errors.Is 可判定），字节为 nil。
func TestParseClassNotFound(t *testing.T) {
	root := t.TempDir()
	jre := makeFakeJre(t, root)
	user := filepath.Join(root, "user")
	writeTestClass(t, user, "pkg/Exists", "x")

	parsed := newTestParse(t, jre, user)

	data, _, err := parsed.ReadClass("no/such/Class")
	if data != nil {
		t.Errorf("ReadClass miss 返回 data = %q, want nil", string(data))
	}
	if !errors.Is(err, ErrClassNotFound) {
		t.Fatalf("ReadClass miss err = %v, want errors.Is(err, ErrClassNotFound)", err)
	}
}

// --- 测试辅助 ---

// newTestParse 是 Parse 的测试包装：构造后注册 zip 句柄释放。
// ZipEntry 按设计进程生命周期内不关 reader，Windows 下不释放会导致
// t.TempDir 清理失败（文件被占用）。
func newTestParse(t *testing.T, jre, cp string) *Classpath {
	t.Helper()
	parsed := Parse(jre, cp)
	t.Cleanup(func() {
		closeEntry(parsed.bootClasspath)
		closeEntry(parsed.extClasspath)
		closeEntry(parsed.userClasspath)
	})
	return parsed
}

// closeEntry 递归关闭组合结构中的全部 ZipEntry（测试清理用）。
func closeEntry(e Entry) {
	switch v := e.(type) {
	case *ZipEntry:
		if v.reader != nil {
			v.reader.Close()
		}
	case CompositeEntry:
		for _, sub := range v {
			closeEntry(sub)
		}
	case WildcardEntry:
		closeEntry(CompositeEntry(v))
	}
}

// makeFakeJre 按 JDK8 真实布局造一个假 jre：<root>/jre/lib/rt.jar
// （boot 段 = <lib>/*），返回 lib 路径（作为 -Xjre 值）。
// rt.jar 是 findJreLib 判定 jre 有效的标志，顺带放一个
// java/lang/Object.class = "boot" 供优先级用例复用。
func makeFakeJre(t *testing.T, root string) string {
	t.Helper()
	lib := filepath.Join(root, "jre", "lib")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", lib, err)
	}
	writeTestJar(t, filepath.Join(lib, "rt.jar"),
		map[string][]byte{"java/lang/Object.class": []byte("boot")})
	return lib
}

// 用例 4：jre 定位失败 → ParseTo 返回 error 并把明确提示写入 errW（不退出进程）。
func TestParseToInvalidJreReportsError(t *testing.T) {
	t.Setenv("JAVA_HOME", "") // 隔离环境，确保探测必然失败

	var errW bytes.Buffer
	cp, err := ParseTo(filepath.Join(t.TempDir(), "no-such-jre"), ".", &errW)
	if err == nil {
		t.Fatalf("ParseTo 无效 jre 应返回 error")
	}
	if cp != nil {
		t.Errorf("失败时 cp = %v, want nil", cp)
	}
	out := errW.String()
	for _, want := range []string{"rt.jar", "-Xjre"} {
		if !strings.Contains(out, want) {
			t.Errorf("错误输出缺少 %q，实际：\n%s", want, out)
		}
	}
}

// 用例 5：-Xjre 有效时直接采用，不探测 JAVA_HOME。
func TestParseToValidJreOptionWins(t *testing.T) {
	jre := makeFakeJre(t, t.TempDir())

	var errW bytes.Buffer
	cp, err := ParseTo(jre, "", &errW)
	if err != nil {
		t.Fatalf("ParseTo(有效 jre) = %v", err)
	}
	// ParseTo 不含释放逻辑（newTestParse 只包 Parse）：手动注册 zip 句柄释放，
	// 否则 Windows 下 t.TempDir 清不掉 rt.jar。
	t.Cleanup(func() {
		closeEntry(cp.bootClasspath)
		closeEntry(cp.extClasspath)
		closeEntry(cp.userClasspath)
	})
	// boot 段来自 <jre>/*：读得到 makeFakeJre 放的 Object。
	data, _, err := cp.ReadClass("java/lang/Object")
	if err != nil || string(data) != "boot" {
		t.Fatalf("ReadClass(java/lang/Object) = (%q, %v), want (\"boot\", nil)", string(data), err)
	}
}

// 用例 6：findJreLib 从 JAVA_HOME/jre/lib 探测成功。
func TestFindJreLibFromJavaHome(t *testing.T) {
	root := t.TempDir()
	lib := makeFakeJre(t, root)
	t.Setenv("JAVA_HOME", root) // root/jre/lib = lib，含 rt.jar

	got, err := findJreLib("")
	if err != nil {
		t.Fatalf("findJreLib = %v", err)
	}
	if got != lib {
		t.Errorf("findJreLib = %q, want %q", got, lib)
	}
}
