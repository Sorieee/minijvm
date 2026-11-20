// Package instructions 实现字节码指令集与解释器。
//
// 计划内容（对应里程碑 M05–M07，详见 docs/milestones/M05-instructions.md）：
//   - Instruction 接口（FetchOperands / Execute）+ BytecodeReader
//   - 按类别分包：constants / loads / stores / stack / math / conversions /
//     comparisons / control / references / extended / reserved
//   - opcode 查表工厂 + fetch-decode-execute 解释器循环
//
// 子包设计在 M05 落地；本包当前仅为占位。
package instructions
