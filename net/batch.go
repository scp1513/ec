package net

// Batch 批量发送
type Batch struct {
	d [][][]byte
	s int
}

// NewBatch 新建Batch
// 参数n<=256
func NewBatch(n int) *Batch {
	if n > 256 {
		n = 256
	}
	return &Batch{d: make([][][]byte, 0, n)}
}

// Add 添加数据
// 超过256个或者总长度超过0xffff之后返回空
func (b *Batch) Add(d ...[]byte) *Batch {
	if len(b.d) >= 256 {
		return nil
	}
	s := b.s + 2
	for i := 0; i < len(d); i++ {
		s += len(d[i])
	}
	if s+1 > 0xffff {
		return nil
	}
	b.s = s
	b.d = append(b.d, d)
	return b
}

// Clear 清除已有缓存
func (b *Batch) Clear() {
	b.d = b.d[0:0]
	b.s = 0
}

func (b *Batch) data() []byte {
	size := b.s + 1
	data := make([]byte, size+3)
	// length
	data[0] = byte(uint16(size) & 0xff)
	data[1] = byte(uint16(size) >> 8)
	// label
	data[2] = 0x33
	// num
	data[3] = byte(len(b.d))
	idx := 5
	for i := 0; i < len(b.d); i++ {
		var l uint16
		var subidx int
		for j := 0; j < len(b.d[i]); j++ {
			subl := copy(data[idx+2+subidx:], b.d[i][j])
			l += uint16(subl)
			subidx += subl
		}

		data[idx] = byte(l & 0x00ff)
		data[idx+1] = byte(l >> 8)
		idx += int(l) + 2
	}
	return data
}
