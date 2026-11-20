package internal

import "testing"

// M01 验收用例（docs/milestones/M01-cmdline.md）。
// 契约：ParseCmdFrom 是 ParseCmd 的可测试接缝，行为完全一致
//（含 normalizeXss 参数规范化）。

// 用例 1：-version
func TestParseCmdFromVersionFlag(t *testing.T) {
	cmd := ParseCmdFrom([]string{"-version"})
	if !cmd.VersionFlag {
		t.Fatalf("VersionFlag = false, want true")
	}
}

// 用例 2：-cp 与主类、程序参数
func TestParseCmdFromFullOptions(t *testing.T) {
	cmd := ParseCmdFrom([]string{
		"-cp", "a;b;c", "-classpath", "d;e", // 后写覆盖先写，与 java 一致
		"-Xjre", `E:\jdk8\jre\lib`,
		"-Xss1m", // 紧凑写法必须被支持
		"Foo", "bar1", "bar2",
	})
	if cmd.CpOption != "d;e" {
		t.Errorf("CpOption = %q, want %q", cmd.CpOption, "d;e")
	}
	if cmd.XjreOption != `E:\jdk8\jre\lib` {
		t.Errorf("XjreOption = %q, want %q", cmd.XjreOption, `E:\jdk8\jre\lib`)
	}
	if cmd.Class != "Foo" {
		t.Errorf("Class = %q, want %q", cmd.Class, "Foo")
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "bar1" || cmd.Args[1] != "bar2" {
		t.Errorf("Args = %v, want [bar1 bar2]", cmd.Args)
	}
	if cmd.XssBytes != 1048576 {
		t.Errorf("XssBytes = %d, want 1048576 (1m)", cmd.XssBytes)
	}
}

// -Xss 未指定时取默认栈大小（512KiB，与书一致）。
func TestParseCmdFromDefaultStackSize(t *testing.T) {
	cmd := ParseCmdFrom([]string{"Foo"})
	if cmd.XssBytes != DefaultStackSize {
		t.Errorf("XssBytes = %d, want DefaultStackSize = %d", cmd.XssBytes, DefaultStackSize)
	}
	if DefaultStackSize != 512*1024 {
		t.Fatalf("DefaultStackSize = %d, want %d", DefaultStackSize, 512*1024)
	}
}

// java 语义：主类名之后的一切都是程序参数，不再是 jvm 的 flag。
func TestParseCmdFromFlagAfterClassIsProgramArg(t *testing.T) {
	cmd := ParseCmdFrom([]string{"Foo", "-version"})
	if cmd.VersionFlag {
		t.Fatalf("VersionFlag = true; 主类之后的 -version 必须进 Args")
	}
	if cmd.Class != "Foo" {
		t.Errorf("Class = %q, want %q", cmd.Class, "Foo")
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "-version" {
		t.Errorf("Args = %v, want [-version]", cmd.Args)
	}
}

// ---- ParseXss：-Xss 值 → 字节数（M01 用例 3）----

func TestParseXss(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"1m", 1048576},
		{"512k", 524288},
		{"1024", 1024}, // 纯数字 = 字节
		{"1M", 1048576},
		{"2K", 2048},
	}
	for _, c := range cases {
		got, err := ParseXss(c.in)
		if err != nil {
			t.Fatalf("ParseXss(%q) unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("ParseXss(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// ---- ParseXss：非法输入（M01 用例 4 的单元部分）----
// 合法域：正整数字节，或 m/k（大小写）后缀；其余一律报错。

func TestParseXssInvalid(t *testing.T) {
	cases := []string{
		"abc",  // 非数字
		"1x",   // 未知后缀
		"",     // 空
		"-1m",  // 负数
		"0",    // 零（栈大小必须为正）
	}
	for _, in := range cases {
		got, err := ParseXss(in)
		if err == nil {
			t.Errorf("ParseXss(%q) = %d, want error", in, got)
		}
	}
}
