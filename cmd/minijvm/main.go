// minijvm 命令行入口。
//
// 当前状态：框架骨架（M00）。里程碑推进路线见 docs/PLAN.md。
package main

import (
	"fmt"
	"os"

	"minijvm/internal"
)

func main() {
	cmd := internal.ParseCmd()

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

	// TODO(M02): 接入 classpath 定位主类字节。
	// TODO(M05): 接入解释器执行主类 main 方法。
	fmt.Printf("classpath: %q\n", cmd.CpOption)
	fmt.Printf("xjre:      %q\n", cmd.XjreOption)
	fmt.Printf("xss:       %q\n", cmd.XssOption)
	fmt.Printf("class:     %s\n", cmd.Class)
	fmt.Printf("args:      %v\n", cmd.Args)
	fmt.Println("(framework only: class locating lands in M02)")
}
