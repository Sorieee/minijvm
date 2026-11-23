// minijvm 命令行入口。
//
// 当前状态：M02——类路径三段查找已接通；解释器执行主类 main 留给 M05。
// 里程碑推进路线见 docs/PLAN.md。
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"minijvm/internal"
	"minijvm/internal/classpath"
	"minijvm/internal/logging"
)

func main() {
	cmd := internal.ParseCmd()

	// 日志初始化：-verbose:class 时类加载追踪走 logging.Debug，
	// 级别至少提升到 debug；初始化失败属致命错误（fail fast），直接退出。
	level := cmd.LogLevel
	if cmd.VerboseClassFlag && level == internal.DefaultLogLevel {
		level = "debug"
	}
	if err := logging.Setup(level); err != nil {
		fmt.Fprintf(os.Stderr, "minijvm: %v\n", err)
		os.Exit(1)
	}

	if cmd.VersionFlag {
		fmt.Printf("minijvm version %s\n", internal.VERSION)
		return
	}
	if cmd.HelpFlag || cmd.Class == "" {
		cmd.PrintUsage()
		if cmd.Class == "" {
			os.Exit(1)
		}
		return
	}

	logging.Debug("starting minijvm",
		"version", internal.VERSION,
		"classpath", cmd.CpOption,
		"xjre", cmd.XjreOption,
		"xss", cmd.XssOption,
		"class", cmd.Class,
	)
	logging.Debug("program args", "args", fmt.Sprintf("%v", cmd.Args))

	// TODO(M05): 拿到字节后交给解释器执行主类 main 方法。
	if err := locateMainClass(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "minijvm: %v\n", err)
		os.Exit(1)
	}
}

// locateMainClass 按 boot → ext → user 定位主类字节，找到后打印
// 「Found class: <名> => <命中 Entry> (<n> bytes)」（M02 验收输出）。
//
// 错误处理：类未找到转成带原始类名的友好错误（用户写的是 demo.Hello，
// 内部查找用的是 demo/Hello）；其余错误（jre 缺失、jar 损坏等）原样上抛。
func locateMainClass(cmd *internal.Cmd) error {
	// 错误信息统一由本函数 / main 打印一次：ParseTo 的错误已经随返回值上抛，
	// 这里传 io.Discard 避免同一错误打印两遍。
	cp, err := classpath.ParseTo(cmd.XjreOption, cmd.CpOption, io.Discard)
	if err != nil {
		return err
	}

	data, hit, err := cp.ReadClass(toInternalClassName(cmd.Class))
	if err != nil {
		if errors.Is(err, classpath.ErrClassNotFound) {
			return fmt.Errorf("class not found: %s（-cp=%q）", cmd.Class, cmd.CpOption)
		}
		return err
	}

	fmt.Printf("Found class: %s => %s (%d bytes)\n", cmd.Class, hit.String(), len(data))
	return nil
}

// toInternalClassName 把主类名从点分（demo.Hello）转成内部斜杠形式（demo/Hello）。
func toInternalClassName(class string) string {
	return strings.ReplaceAll(class, ".", "/")
}
