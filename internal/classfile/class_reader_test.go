// ClassReader 单元测试（M03 任务 1）。
package classfile

import (
	"errors"
	"testing"
)

// 大端读取：u1/u2/u4/u8 必须按 JVM 规范顺序消费字节。
func TestClassReaderReadBigEndian(t *testing.T) {
	data := []byte{
		0x01,
		0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
	}
	r := NewClassReader(data)

	if got, err := r.ReadUint8(); err != nil || got != 0x01 {
		t.Fatalf("ReadUint8 = %#x, %v; want 0x01, nil", got, err)
	}
	if got, err := r.ReadUint16(); err != nil || got != 0x0203 {
		t.Fatalf("ReadUint16 = %#x, %v; want 0x0203, nil", got, err)
	}
	if got, err := r.ReadUint32(); err != nil || got != 0x04050607 {
		t.Fatalf("ReadUint32 = %#x, %v; want 0x04050607, nil", got, err)
	}
	if got, err := r.ReadUint64(); err != nil || got != 0x08090A0B0C0D0E0F {
		t.Fatalf("ReadUint64 = %#x, %v; want 0x08090A0B0C0D0E0F, nil", got, err)
	}
	if r.Pos() != len(data) || r.Remaining() != 0 {
		t.Errorf("Pos/Remaining = %d/%d, want %d/0", r.Pos(), r.Remaining(), len(data))
	}
}

// ReadBytes 顺序前进；Skip 消费但不返回字节。
func TestClassReaderReadBytesAndSkip(t *testing.T) {
	r := NewClassReader([]byte{1, 2, 3, 4, 5})

	b, err := r.ReadBytes(2)
	if err != nil || len(b) != 2 || b[0] != 1 || b[1] != 2 {
		t.Fatalf("ReadBytes(2) = %v, %v; want [1 2]", b, err)
	}
	if err := r.Skip(2); err != nil {
		t.Fatalf("Skip(2): %v", err)
	}
	if got, err := r.ReadUint8(); err != nil || got != 5 {
		t.Fatalf("ReadUint8 = %d, %v; want 5", got, err)
	}
	if r.Remaining() != 0 {
		t.Errorf("Remaining = %d, want 0", r.Remaining())
	}
}

// 越界读必须返回包装 ErrClassFormat 的错误，错误信息带当前偏移。
func TestClassReaderOutOfBounds(t *testing.T) {
	r := NewClassReader([]byte{1, 2, 3})

	_, err := r.ReadBytes(4)
	if !errors.Is(err, ErrClassFormat) {
		t.Fatalf("ReadBytes(4) error = %v, want ErrClassFormat", err)
	}
	if r.Pos() != 0 {
		t.Errorf("Pos = %d, want 0（失败读取不应前进）", r.Pos())
	}

	// 已消费 2 字节后再读 2 字节（只剩 1 字节）。
	if err := r.Skip(2); err != nil {
		t.Fatalf("Skip(2): %v", err)
	}
	_, err = r.ReadUint16()
	if !errors.Is(err, ErrClassFormat) {
		t.Fatalf("ReadUint16 error = %v, want ErrClassFormat", err)
	}
}
