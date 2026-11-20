// Package native 实现本地方法的注册与查找。
//
// 计划内容（对应里程碑 M11，详见 docs/milestones/M11-native.md）：
//   - 注册表：className → "name+descriptor" → NativeMethod(frame)
//   - 第一批实现：Object / Class / System / Float / Double / Throwable /
//     FileOutputStream（println 全真实链路的前置）
package native
