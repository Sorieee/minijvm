// Package classfile 实现class 文件的解析（JVM 规范 §4）。
//
// 计划内容（对应里程碑 M03，详见 docs/milestones/M03-classfile.md）：
//   - ClassReader 大端读取器（u1/u2/u4/u8）
//   - ConstantPool 常量池（17 种 tag，Long/Double 占双槽）
//   - MemberInfo 字段/方法，AttributeInfo 属性（Code/ConstantValue/Exceptions/LineNumberTable）
//   - ClassFile 顶层结构
package classfile
