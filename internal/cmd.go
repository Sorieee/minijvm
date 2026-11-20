// Package internal 提供 minijvm 的命令行解析与全局配置。
package internal

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Cmd 保存 java 风格的命令行参数（设计见 docs/milestones/M01-cmdline.md）。
type Cmd struct {
	HelpFlag         bool     // -h / -help
	VersionFlag      bool     // -version
	CpOption         string   // -cp / -classpath（分隔符取 filepath.ListSeparator）
	XjreOption       string   // -Xjre <path>，JDK8 的 jre/lib 目录（rt.jar 类库来源）
	XssOption        string   // -Xss <size>，如 1m / 512k / 纯数字字节
	VerboseClassFlag bool     // -verbose:class
	Class            string   // 主类全限定名（com/foo/Bar）
	Args             []string // 传给 main 的参数
}

// ParseCmd 解析命令行参数。
//
// 与 java 命令一致：flag 必须出现在主类名之前，主类名之后的参数
// 原样透传给程序 main。参数非法时由 flag（ExitOnError）打印错误并退出。
func ParseCmd() *Cmd {
	cmd := &Cmd{}

	fs := flag.NewFlagSet("minijvm", flag.ExitOnError)
	fs.BoolVar(&cmd.HelpFlag, "h", false, "print help")
	fs.BoolVar(&cmd.HelpFlag, "help", false, "print help")
	fs.BoolVar(&cmd.VersionFlag, "version", false, "print version and exit")
	fs.StringVar(&cmd.CpOption, "cp", "", "user classpath, entries separated by path list separator")
	fs.StringVar(&cmd.CpOption, "classpath", "", "same as -cp")
	fs.StringVar(&cmd.XjreOption, "Xjre", "", "path to JDK8 jre/lib (runtime class library)")
	fs.StringVar(&cmd.XssOption, "Xss", "", "thread stack size, e.g. 1m / 512k / bytes")
	fs.BoolVar(&cmd.VerboseClassFlag, "verbose:class", false, "log every loaded class")
	fs.Usage = cmd.PrintUsage

	args := os.Args[1:]
	for {
		fs.Parse(args) // 停在第一个非 flag 参数处
		rest := fs.Args()
		if len(rest) == 0 {
			return cmd
		}
		if cmd.Class != "" {
			// 主类名之后又出现位置参数：不合法（程序参数应紧跟主类）。
			cmd.PrintUsage()
			os.Exit(1)
		}
		cmd.Class = rest[0]
		cmd.Args = rest[1:]
		return cmd
	}
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
  -verbose:class     log every loaded class
`, filepath.ListSeparator)
}
