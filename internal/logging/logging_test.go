package logging

import (
	"io"
	"strings"
	"testing"
)

// 用例 1：未知级别必须报错。
func TestSetupRejectsUnknownLevel(t *testing.T) {
	err := SetupTo("bogus", io.Discard)
	if err == nil {
		t.Fatalf("SetupTo(\"bogus\") = nil error, want error")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("错误信息未包含非法值 %q：%v", "bogus", err)
	}
}

// 用例 2：级别过滤 —— warn 级别下 Info 被丢弃，Warn/Error 保留。
func TestLevelFiltering(t *testing.T) {
	var buf strings.Builder
	if err := SetupTo("warn", &buf); err != nil {
		t.Fatalf("SetupTo(warn) = %v", err)
	}

	Info("dropped-info")
	Warn("kept-warn", "k", 1)
	Error("kept-error")

	out := buf.String()
	if strings.Contains(out, "dropped-info") {
		t.Errorf("warn 级别下 Info 不应输出，实际：\n%s", out)
	}
	for _, want := range []string{"WARN", "kept-warn", "ERROR", "kept-error", "k=1"} {
		if !strings.Contains(out, want) {
			t.Errorf("输出缺少 %q，实际：\n%s", want, out)
		}
	}
}

// 用例 3：IsDebug 门控 —— debug 级别开启为 true，error 级别下为 false。
func TestIsDebugGate(t *testing.T) {
	if err := SetupTo("debug", io.Discard); err != nil {
		t.Fatalf("SetupTo(debug) = %v", err)
	}
	if !IsDebug() {
		t.Errorf("debug 级别下 IsDebug() = false, want true")
	}

	if err := SetupTo("error", io.Discard); err != nil {
		t.Fatalf("SetupTo(error) = %v", err)
	}
	if IsDebug() {
		t.Errorf("error 级别下 IsDebug() = true, want false")
	}
}

// 用例 4：单行格式 —— <时间> <级别> <消息> <key=value>。
func TestLineFormat(t *testing.T) {
	var buf strings.Builder
	if err := SetupTo("debug", &buf); err != nil {
		t.Fatalf("SetupTo(debug) = %v", err)
	}

	Debug("loading class", "class", "java/lang/Object")

	line := buf.String()
	if line == "" || line[len(line)-1] != '\n' {
		t.Fatalf("日志必须以换行结尾，实际：\n%q", line)
	}
	if strings.Count(line, "\n") != 1 {
		t.Fatalf("日志必须单行输出，实际 %d 行：\n%s", strings.Count(line, "\n"), line)
	}
	for _, want := range []string{"DEBUG", "loading class", "class=java/lang/Object"} {
		if !strings.Contains(line, want) {
			t.Errorf("输出缺少 %q，实际：\n%s", want, line)
		}
	}
}

// 用例 5：大小写不敏感 + 空串回退 info。
func TestLevelAliases(t *testing.T) {
	cases := map[string]bool{
		"DEBUG": true, "Debug": true, "debug": true,
		"INFO": true, "": true,
		"WARN": true, "warning": true,
		"ERROR": true,
	}
	for in, ok := range cases {
		err := SetupTo(in, io.Discard)
		if ok && err != nil {
			t.Errorf("SetupTo(%q) = %v, want nil", in, err)
		}
		if !ok && err == nil {
			t.Errorf("SetupTo(%q) = nil, want error", in)
		}
	}
}
