// ClassReader：class 文件的大端读取器（M03 任务 1，对应书 3.1 节）。
package classfile

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// ErrClassFormat 表示 class 文件格式非法。
//
// 约定（与 classpath 包的 ErrClassNotFound 同风格）：所有格式错误都用
// fmt.Errorf + %w 包装本哨兵，调用方用 errors.Is(err, ErrClassFormat)
// 判定；错误信息里带出错的字节偏移，便于对着 javap 输出定位。
var ErrClassFormat = errors.New("class format error")

// ClassReader 是 class 文件的顺序读取器。
//
// data 是整个 class 文件的字节；pos 是当前读取位置（所有读取都是
// 从 pos 顺序前进，不回退）。
//
// cp 是常量池引用：解析成员与属性时需要按索引回查 Utf8 等条目，
// 由 Parse 在常量池解析完后回填（书里的做法——读取器持有常量池，
// 避免把 cp 一路穿透给每个子解析函数）。
type ClassReader struct {
	data []byte
	pos  int
	cp   ConstantPool // 解析成员/属性阶段才有效；常量池自身解析时尚未填充
}

// NewClassReader 构造读取器（不校验 magic，magic 校验在 Parse 里做）。
func NewClassReader(data []byte) *ClassReader {
	return &ClassReader{data: data}
}

// ReadUint8 读取 1 字节（u1）。
func (r *ClassReader) ReadUint8() (uint8, error) {
	b, err := r.ReadBytes(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// ReadUint16 读取 2 字节无符号整数（u2）。
func (r *ClassReader) ReadUint16() (uint16, error) {
	b, err := r.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

// ReadUint32 读取 4 字节无符号整数（u4）。
func (r *ClassReader) ReadUint32() (uint32, error) {
	b, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

// ReadUint64 读取 8 字节无符号整数（u8）。
//
// 注意：class 文件里没有 u8 字段，但常量池的 long/double 是 8 字节，
// 且 heap 层将来可能用到（如 java.util.Date 的时间戳），故一并提供。
func (r *ClassReader) ReadUint64() (uint64, error) {
	b, err := r.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

// ReadBytes 读取 n 个字节。
//
// 返回的切片与底层数据共享内存（不复制），调用方只读使用；需要独立
// 副本时自行 copy（例如测试里改字节做「magic 错误」用例）。
func (r *ClassReader) ReadBytes(n uint32) ([]byte, error) {
	if uint64(r.pos)+uint64(n) > uint64(len(r.data)) {
		return nil, fmt.Errorf("%w: 偏移 %d 处需要 %d 字节，但只剩 %d 字节",
			ErrClassFormat, r.pos, n, len(r.data)-r.pos)
	}
	b := r.data[r.pos : r.pos+int(n)]
	r.pos += int(n)
	return b, nil
}

// Skip 跳过 n 个字节（未知属性按长度跳过时用）。
func (r *ClassReader) Skip(n uint32) error {
	_, err := r.ReadBytes(n)
	return err
}

// Pos 返回当前读取偏移（错误信息里用于定位）。
func (r *ClassReader) Pos() int { return r.pos }

// Remaining 返回剩余未读字节数。
func (r *ClassReader) Remaining() int { return len(r.data) - r.pos }
