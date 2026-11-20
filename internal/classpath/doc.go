// Package classpath 实现类路径体系：
// 在目录、jar 包、通配符目录、组合路径中定位并读取类字节。
//
// 计划内容（对应里程碑 M02，详见 docs/milestones/M02-classpath.md）：
//   - Entry 接口 + DirEntry / ZipEntry / CompositeEntry / WildcardEntry 四种实现
//   - Classpath 三段结构：boot(jre/lib) → ext(jre/lib/ext) → user(-cp)
package classpath
