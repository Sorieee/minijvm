// Package internal 提供 minijvm 的命令行解析与全局配置。
package internal

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DefaultLogLevel 是 -Xlog 未指定时的默认日志级别（与 main 的
// -verbose:class 提升判断共用，避免字面量散落）。
const DefaultLogLevel = "info"

// DefaultStackSize 是 -Xss 未指定时的默认栈大小（512KiB，与书一致）。
const DefaultStackSize = 512 * 1024

// Cmd 保存 java 风格的命令行参数（设计见 docs/milestones/M01-cmdline.md）。
type Cmd struct {
	HelpFlag         bool     // -h / -help
	VersionFlag      bool     // -version
	CpOption         string   // -cp / -classpath（分隔符取 filepath.ListSeparator）
	XjreOption       string   // -Xjre <path>，JDK8 的 jre/lib 目录（rt.jar 类库来源）
	XssOption        string   // -Xss <size>，如 1m / 512k / 纯数字字节（原始值）
	XssBytes         int      // -Xss 解析后的字节数；未指定 = DefaultStackSize
	LogLevel         string   // -Xlog <level>，日志级别 debug|info|warn|error（默认 info）
	VerboseClassFlag bool     // -verbose:class
	Class            string   // 主类全限定名（com/foo/Bar）
	Args             []string // 传给 main 的参数
}

// ParseCmd 解析命令行参数（os.Args）。
func ParseCmd() *Cmd {
	return ParseCmdFrom(os.Args[1:])
}

// ParseCmdFrom 解析给定的参数列表，是 ParseCmd 的可测试接缝：
// 行为完全一致（含 normalizeXss 参数规范化）。
//
// 与 java 命令一致：flag 必须出现在主类名之前，主类名之后的参数
// 原样透传给程序 main。参数非法时由 flag（ExitOnError）打印错误并退出。
func ParseCmdFrom(args []string) *Cmd {
	cmd := &Cmd{}

	fs := flag.NewFlagSet("minijvm", flag.ExitOnError)
	fs.BoolVar(&cmd.HelpFlag, "h", false, "print help")
	fs.BoolVar(&cmd.HelpFlag, "help", false, "print help")
	fs.BoolVar(&cmd.VersionFlag, "version", false, "print version and exit")
	fs.StringVar(&cmd.CpOption, "cp", "", "user classpath, entries separated by path list separator")
	fs.StringVar(&cmd.CpOption, "classpath", "", "same as -cp")
	fs.StringVar(&cmd.XjreOption, "Xjre", "", "path to JDK8 jre/lib (runtime class library)")
	fs.StringVar(&cmd.XssOption, "Xss", "", "thread stack size, e.g. 1m / 512k / bytes")
	fs.StringVar(&cmd.LogLevel, "Xlog", DefaultLogLevel, "log level: debug|info|warn|error")
	fs.BoolVar(&cmd.VerboseClassFlag, "verbose:class", false, "log every loaded class")
	fs.Usage = cmd.PrintUsage

	fs.Parse(normalizeXss(args)) // 停在第一个非 flag 参数处
	if rest := fs.Args(); len(rest) > 0 {
		cmd.Class = rest[0]
		cmd.Args = rest[1:] // 主类之后的都是程序参数（java 语义）
	}

	// -Xss 解析：非法值立即友好报错并退出（契约见 M01 里程碑文档）。
	cmd.XssBytes = DefaultStackSize
	if cmd.XssOption != "" {
		n, err := ParseXss(cmd.XssOption)
		if err != nil {
			fmt.Fprintf(os.Stderr, "minijvm: invalid -Xss %q: %v\n", cmd.XssOption, err)
			fmt.Fprintln(os.Stderr, "用法: -Xss <size>，如 1m / 512k / 1048576（字节）")
			os.Exit(1)
		}
		cmd.XssBytes = n
	}
	return cmd
}

// ParseXss 把 -Xss 的值解析成字节数。
//
// 合法域：正整数（字节），或带 m/k 后缀的正整数（后缀大小写均可，
// m = 1MiB，k = 1KiB）。其余（空串、零、负数、未知后缀、非数字）一律报错。
func ParseXss(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("空值")
	}

	mult := 1
	switch s[len(s)-1] {
	case 'm', 'M':
		mult = 1 << 20
		s = s[:len(s)-1]
	case 'k', 'K':
		mult = 1 << 10
		s = s[:len(s)-1]
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%q 不是合法的数字", s)
	}
	if n <= 0 {
		return 0, fmt.Errorf("栈大小必须为正数，得到 %d", n)
	}
	if n > math.MaxInt/mult { // 防乘溢出成负数
		return 0, errors.New("栈大小过大")
	}
	return n * mult, nil
}

// normalizeXss 把 -Xss1m 这类「值紧跟参数名」的写法改写为 -Xss=1m。
// java 命令支持紧凑写法，但 Go 的 flag 包只认 -flag value / -flag=value。
//
// 只处理主类名之前的 flag 区：主类之后的一切都是程序参数，必须原样
// 透传（否则 `Foo -Xss1m` 会被误改写成 `Foo -Xss=1m`）。
func normalizeXss(args []string) []string {
	// 取独立值的 flag：遇到时跳过下一个 token，避免把 flag 值误判成主类名。
	takesValue := map[string]bool{
		"-cp": true, "-classpath": true, "-Xjre": true, "-Xss": true, "-Xlog": true,
	}

	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		// 第一个位置参数即主类名（或 "--" 之后的全部参数），其后原样保留。
		if a == "--" || !strings.HasPrefix(a, "-") || a == "-" {
			out = append(out, args[i:]...)
			return out
		}
		if strings.HasPrefix(a, "-Xss") && a != "-Xss" && !strings.HasPrefix(a, "-Xss=") {
			out = append(out, "-Xss="+strings.TrimPrefix(a, "-Xss"))
			continue
		}
		out = append(out, a)

		name := a
		if idx := strings.IndexByte(a, '='); idx >= 0 {
			name = a[:idx]
		}
		if takesValue[name] && !strings.Contains(a, "=") && i+1 < len(args) {
			i++
			out = append(out, args[i])
		}
	}
	return out
}

// PrintUsage 打印用法说明。
func (cmd *Cmd) PrintUsage() {
	fmt.Fprintf(os.Stderr, `Usage: minijvm [-options] <class> [args...]

Options:
  -h, -help          print this help
  -version           print version and exit
  -cp <path>         user classpath, entries separated by '%c'
  -classpath <path>  same as -cp
  -Xjre <path>       JDK8 jre/lib directory (runtime class library source)
  -Xss <size>        thread stack size, e.g. 1m / 512k / plain bytes
  -Xlog <level>      log level: debug|info|warn|error (default info)
  -verbose:class     log every loaded class
`, filepath.ListSeparator)
}
