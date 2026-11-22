// Classpath：三段式类路径（M02 任务 3，对应书 2.7 节）。
package classpath

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
//	user = NewEntry(cpOption)，cpOption 为空时用 "."
//
// jre 定位顺序（findJreLib）：
//  1. jreOption 非空 → 直接采用
//  2. 为空 → 依次探测 JAVA_HOME/jre/lib、JAVA_HOME/lib、常见安装路径
//  3. 都失败 → 打印明确错误（提示显式传 -Xjre E:\tools\jdk8\jre\lib）并 os.Exit(1)
//
// M02 简化：jre 定位失败直接退出进程；M14 引入 ClassLoader 体系时再重构为返回 error。
func Parse(jreOption, cpOption string) *Classpath {
	panic("TODO(M02): not implemented")
}

// findJreLib 按 Parse 注释所述顺序定位 jre 的 lib 目录，返回绝对路径。
// 各候选用本地探测（目录存在且含 rt.jar 或至少存在）判定；找不到返回空串，
// 由调用方负责报错退出（保持 Parse 的签名与里程碑一致）。
func findJreLib(jreOption string) string {
	panic("TODO(M02): not implemented")
}

// ReadClass 按固定顺序 boot → ext → user 查找类。
//
// className 形如 "java/lang/Object"；任一段命中即返回（命中 Entry 指明来源，
// 如 rt.jar）；三段都 miss 返回 ErrClassNotFound 包装错误。
// 这保证了 java.lang.Object 等系统类永远优先取自 rt.jar，而非用户 -cp。
func (cp *Classpath) ReadClass(className string) ([]byte, Entry, error) {
	panic("TODO(M02): not implemented")
}
