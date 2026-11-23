package main

// M01 集成验收用例（docs/milestones/M01-cmdline.md 用例 5、6 及用例 4 的退出码部分）。
// TestMain 先把当前包编译成真实二进制，再以子进程方式运行——验证退出码与用户可见输出。

import (
	"archive/zip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var exePath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "minijvm-integration-*")
	if err != nil {
		panic(err)
	}
	exePath = filepath.Join(dir, "minijvm-test.exe")
	build := exec.Command("go", "build", "-o", exePath, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		os.RemoveAll(dir)
		panic("build minijvm binary for integration tests: " + err.Error())
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// runMinijvm 运行编译出的二进制，返回合并输出与退出码。
func runMinijvm(t *testing.T, args ...string) (string, int) {
	t.Helper()
	out, err := exec.Command(exePath, args...).CombinedOutput()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("run minijvm %v: %v", args, err)
	}
	return string(out), code
}

// 用例 5：minijvm -version
func TestBinaryVersion(t *testing.T) {
	out, code := runMinijvm(t, "-version")
	if code != 0 {
		t.Fatalf("exit = %d, want 0; output:\n%s", code, out)
	}
	if !strings.Contains(out, "minijvm version 0.0.1") {
		t.Fatalf("want stdout to contain version line, got:\n%s", out)
	}
}

// 用例 6：无参数 → 打印用法，退出码 ≠ 0
func TestBinaryNoArgsPrintsUsage(t *testing.T) {
	out, code := runMinijvm(t)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero; output:\n%s", out)
	}
	if !strings.Contains(out, "Usage") {
		t.Fatalf("want output to contain usage, got:\n%s", out)
	}
}

// 用例 4（集成部分）：-Xssabc → 友好报错（含 "invalid -Xss"），退出码 ≠ 0
func TestBinaryInvalidXssExitsNonZero(t *testing.T) {
	out, code := runMinijvm(t, "-Xssabc", "Foo")
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero for invalid -Xss; output:\n%s", out)
	}
	if !strings.Contains(out, "invalid -Xss") {
		t.Fatalf(`want friendly message containing "invalid -Xss", got:`+"\n%s", out)
	}
}

// M02 集成验收用例（docs/milestones/M02-classpath.md 用例 8）。
// 测试数据自造（伪造 jre/lib/rt.jar + 用户类目录），不依赖本机 JDK。

// writeJar 造一个最小 jar（条目名 → 内容）。
func writeJar(t *testing.T, path string, entries map[string][]byte) {
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

// setupFakeJre 造 <root>/jre/lib/rt.jar（findJreLib 以 rt.jar 判定 jre 有效）。
func setupFakeJre(t *testing.T, root string) string {
	t.Helper()
	lib := filepath.Join(root, "jre", "lib")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJar(t, filepath.Join(lib, "rt.jar"),
		map[string][]byte{"java/lang/Object.class": {0xCA, 0xFE, 0xBA, 0xBE}})
	return lib
}

// 用例 8：-Xjre + -cp 定位主类 → 打印 Found class + 命中 Entry + 字节长度，退出码 0。
func TestBinaryFindsMainClass(t *testing.T) {
	root := t.TempDir()
	jre := setupFakeJre(t, root)

	classes := filepath.Join(root, "classes")
	if err := os.MkdirAll(filepath.Join(classes, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	// 主类名含包：验证 '.' → '/' 转换。
	classBytes := []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34}
	if err := os.WriteFile(filepath.Join(classes, "demo", "Hello.class"), classBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runMinijvm(t, "-Xjre", jre, "-cp", classes, "demo.Hello")
	if code != 0 {
		t.Fatalf("exit = %d, want 0; output:\n%s", code, out)
	}
	if !strings.Contains(out, "Found class: demo.Hello") {
		t.Errorf("输出缺少 Found class 行，实际：\n%s", out)
	}
	if !strings.Contains(out, "demo/Hello") && !strings.Contains(out, classes) {
		t.Errorf("输出未指明命中来源，实际：\n%s", out)
	}
	if !strings.Contains(out, strconv.Itoa(len(classBytes))+" bytes") {
		t.Errorf("输出缺少字节长度 %d bytes，实际：\n%s", len(classBytes), out)
	}
}

// 类找不到 → 友好错误 + 退出码非 0（错误里带类名，不吐原始 panic/堆栈）。
func TestBinaryClassNotFound(t *testing.T) {
	root := t.TempDir()
	jre := setupFakeJre(t, root)
	classes := filepath.Join(root, "classes")
	if err := os.MkdirAll(classes, 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runMinijvm(t, "-Xjre", jre, "-cp", classes, "no.Such")
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero; output:\n%s", out)
	}
	if !strings.Contains(out, "class not found") || !strings.Contains(out, "no.Such") {
		t.Fatalf("want friendly class-not-found error mentioning the class, got:\n%s", out)
	}
}
