// Classpath：三段式类路径（M02 任务 3，对应书 2.7 节）。
package classpath

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"minijvm/internal/logging"
)

// Classpath 模拟 java 的三段类路径模型，查找顺序固定：boot → ext → user。
type Classpath struct {
	bootClasspath Entry // jre 启动类库：<Xjre>/*（rt.jar 等全部 jar）
	extClasspath  Entry // 扩展类库：<Xjre>/ext/*；目录缺失时为空组合（不报错）
	userClasspath Entry // -cp / -classpath；未指定时默认 "."（与 java 一致）
}

// Parse 从命令行选项构造三段 classpath。
//
// XjreOption 语义（易错点）：它直接指向 jre 的 **lib 目录** 本身，
// 例如 E:\tools\jdk8\jre\lib——不要再往下拼 lib。三段构造：
//
//	boot = <lib>/*      （newWildcardEntry）
//	ext  = <lib>/ext/*  （目录不存在得空组合，由 WildcardEntry 自然处理）
//	user = NewEntry(cpOption)，多段 ListSeparator 拆分；为空时按 java 语义落当前目录
//
// jre 定位顺序（findJreLib）：
//  1. jreOption 非空且含 rt.jar → 直接采用；非空但无 rt.jar → 记入错误提示
//  2. 为空 → 依次探测 JAVA_HOME/jre/lib、JAVA_HOME/lib
//  3. 都失败 → 打印明确错误（含无效的 -Xjre 路径，提示显式传
//     -Xjre E:\tools\jdk8\jre\lib）并 os.Exit(1)
//
// M02 简化：jre 定位失败直接退出进程；M14 引入 ClassLoader 体系时再重构。
// 可测试行为在 ParseTo（接缝），本函数只是命令行壳。
func Parse(jreOption, cpOption string) *Classpath {
	cp, err := ParseTo(jreOption, cpOption, os.Stderr)
	if err != nil {
		os.Exit(1)
	}
	return cp
}

// ParseTo 是 Parse 的可测试接缝：jre 定位失败时把错误写入 errW 并返回 error，
// 不退出进程；其余行为与 Parse 完全一致（模式同 internal.ParseCmdFrom）。
func ParseTo(jreOption, cpOption string, errW io.Writer) (*Classpath, error) {
	foundPath, err := findJreLib(jreOption)
	if err != nil {
		fmt.Fprintf(errW, "minijvm: %v\n", err)
		return nil, err
	}
	logging.Debug("jre lib located", "path", foundPath)

	cp := &Classpath{
		bootClasspath: newSingleEntry(filepath.Join(foundPath, "*")),
		extClasspath:  newSingleEntry(filepath.Join(foundPath, "ext", "*")),
		userClasspath: NewEntry(cpOption),
	}
	return cp, nil
}

// findJreLib 定位 JDK8 jre 的 lib 目录（判定标志：目录下存在 rt.jar），
// 返回该目录路径；全部候选失败时返回带上下文的错误。
func findJreLib(jreOption string) (string, error) {
	invalidHint := ""
	if jreOption != "" {
		if hasRtJar(jreOption) {
			return jreOption, nil
		}
		// 用户显式传了路径但无效：错误信息必须点名，不能静默换用别的候选。
		invalidHint = fmt.Sprintf("（-Xjre 路径 %q 下未找到 rt.jar）", jreOption)
	}
	if javahome := os.Getenv("JAVA_HOME"); javahome != "" {
		for _, sub := range []string{filepath.Join("jre", "lib"), "lib"} {
			p := filepath.Join(javahome, sub)
			if hasRtJar(p) {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("无法定位 JDK8 jre 的 lib 目录%s；请显式指定 -Xjre <jdk8>/jre/lib", invalidHint)
}

// hasRtJar 报告 dir 下是否存在 rt.jar（JDK8 jre/lib 的特征文件）。
func hasRtJar(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "rt.jar"))
	return err == nil
}

// ReadClass 按固定顺序 boot → ext → user 查找类。
//
// className 形如 "java/lang/Object"（.class 后缀由各 Entry 自行补全）；
// 任一段命中即返回（命中 Entry 指明来源，如 rt.jar）；段间未命中
// （ErrClassNotFound）继续下一段，真实 IO 错误（jar 损坏等）立即上抛，
// 与 CompositeEntry 契约一致。这保证 java.lang.Object 等系统类永远
// 优先取自 rt.jar，而非用户 -cp。
//
// 命中时以 debug 级别记录类加载（-verbose:class 追踪的数据源）。
func (cp *Classpath) ReadClass(className string) ([]byte, Entry, error) {
	var lastErr error
	for _, seg := range []struct {
		section string
		entry   Entry
	}{
		{"boot", cp.bootClasspath},
		{"ext", cp.extClasspath},
		{"user", cp.userClasspath},
	} {
		data, hit, err := seg.entry.ReadClass(className)
		switch {
		case err == nil:
			logging.Debug("loaded class",
				"class", className, "section", seg.section, "from", hit.String())
			return data, hit, nil
		case errors.Is(err, ErrClassNotFound):
			lastErr = err // 本段未命中，继续下一段
		default:
			return nil, nil, err
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("%w: %s（boot/ext/user 均未配置）", ErrClassNotFound, className)
	}
	return nil, nil, lastErr
}
