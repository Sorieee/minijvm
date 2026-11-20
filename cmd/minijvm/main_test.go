package main

// M01 集成验收用例（docs/milestones/M01-cmdline.md 用例 5、6 及用例 4 的退出码部分）。
// TestMain 先把当前包编译成真实二进制，再以子进程方式运行——验证退出码与用户可见输出。

import (
	"os"
	"os/exec"
	"path/filepath"
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
