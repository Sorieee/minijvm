// Package heap 实现运行时的类与对象模型（自足包，不依赖 rtda）。
//
// 计划内容（对应里程碑 M06–M08、M12–M13，详见 docs/DESIGN.md）：
//   - Slot / Object（字段槽布局）/ Class / Method / Field
//   - 符号引用与惰性解析（ClassRef / MethodRef / FieldRef）
//   - 类加载生命周期（define / link / init）与 <clinit> 防重入
//   - StringPool 字符串池、数组类合成
package heap
