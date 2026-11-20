// Package rtda 实现线程私有的运行时数据区。
//
// 计划内容（对应里程碑 M04，详见 docs/milestones/M04-rtda.md）：
//   - Slot 槽（long/double 占两槽；M06 起引用 heap 包类型）
//   - LocalVars 局部变量表 / OperandStack 操作数栈
//   - Frame 帧（链表）/ JvmStack 栈（-Xss 限深）/ Thread 线程
//
// 依赖方向：rtda → heap（单向，见 docs/DESIGN.md §3）。
package rtda
