package sm3

import (
	"encoding/binary"
	"hash"
)

const (
	Size      = 32
	BlockSize = 64
	chunk     = 64
)

// 预计算好的 Tj <<< j (参考 RustCrypto 预计算常数)
var tRot = [64]uint32{
	0x79cc4519, 0xf3988a32, 0xe7311465, 0xce6228cb, 0x9cc45197, 0x3988a32f, 0x7311465e, 0xe6228cbc,
	0xcc451979, 0x988a32f3, 0x311465e7, 0x6228cbce, 0xc451979c, 0x88a32f39, 0x11465e73, 0x228cbce6,
	0x9d8a7a87, 0x3b14f50f, 0x7629ea1e, 0xec53d43c, 0xd8a7a879, 0xb14f50f3, 0x629ea1e7, 0xc53d43ce,
	0x8a7a879d, 0x14f50f3b, 0x29ea1e76, 0x53d43cec, 0xa7a879d8, 0x4f50f3b1, 0x9ea1e762, 0x3d43cec5,
	0x7a879d8a, 0xf50f3b14, 0xea1e7629, 0xd43cec53, 0xa879d8a7, 0x50f3b14f, 0xa1e7629e, 0x43cec53d,
	0x879d8a7a, 0x0f3b14f5, 0x1e7629ea, 0x3cec53d4, 0x79d8a7a8, 0xf3b14f50, 0xe7629ea1, 0xcec53d43,
	0x9d8a7a87, 0x3b14f50f, 0x7629ea1e, 0xec53d43c, 0xd8a7a879, 0xb14f50f3, 0x629ea1e7, 0xc53d43ce,
	0x8a7a879d, 0x14f50f3b, 0x29ea1e76, 0x53d43cec, 0xa7a879d8, 0x4f50f3b1, 0x9ea1e762, 0x3d43cec5,
}

var iv = [8]uint32{
	0x7380166f, 0x4914b2b9, 0x172442d7, 0xda8a0600,
	0xa96f30bc, 0x163138aa, 0xe38dee4d, 0xb0fb0e4e,
}

type digest struct {
	h   [8]uint32
	x   [chunk]byte
	nx  int
	len uint64
}

func (d *digest) Reset() {
	d.h = iv
	d.nx = 0
	d.len = 0
}

func New() hash.Hash {
	d := new(digest)
	d.Reset()
	return d
}

func (d *digest) Size() int      { return Size }
func (d *digest) BlockSize() int { return BlockSize }

func (d *digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	d.len += uint64(nn)
	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == chunk {
			block(d, d.x[:])
			d.nx = 0
		}
		p = p[n:]
	}
	if len(p) >= chunk {
		n := len(p) &^ (chunk - 1)
		block(d, p[:n])
		p = p[n:]
	}
	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}
	return
}

func (d *digest) Sum(in []byte) []byte {
	d0 := *d
	hash := d0.checkSum()
	return append(in, hash[:]...)
}

func (d *digest) checkSum() [Size]byte {
	len := d.len
	var tmp [64]byte
	tmp[0] = 0x80
	if len%64 < 56 {
		d.Write(tmp[0 : 56-len%64])
	} else {
		d.Write(tmp[0 : 64+56-len%64])
	}

	len <<= 3
	binary.BigEndian.PutUint64(tmp[:], len)
	d.Write(tmp[0:8])

	var out [Size]byte
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(out[i*4:], d.h[i])
	}
	return out
}

func P0(x uint32) uint32 { return x ^ (x<<9 | x>>(32-9)) ^ (x<<17 | x>>(32-17)) }
func P1(x uint32) uint32 { return x ^ (x<<15 | x>>(32-15)) ^ (x<<23 | x>>(32-23)) }

func block(d *digest, p []byte) {
	var w [68]uint32
	for len(p) >= chunk {
		a, b, c, dVal, e, f, g, h := d.h[0], d.h[1], d.h[2], d.h[3], d.h[4], d.h[5], d.h[6], d.h[7]

		for j := 0; j < 16; j++ {
			w[j] = binary.BigEndian.Uint32(p[j*4:])
		}
		for j := 16; j < 68; j++ {
			// 消息扩展
			w[j] = P1(w[j-16]^w[j-9]^(w[j-3]<<15|w[j-3]>>(32-15))) ^ (w[j-13]<<7 | w[j-13]>>(32-7)) ^ w[j-6]
		}

		// 0-15 轮: FF1, GG1 (循环展开 8 轮一组)
		for j := 0; j < 16; j += 8 {
			round1(&a, &b, &c, &dVal, &e, &f, &g, &h, j, &w)
			round1(&a, &b, &c, &dVal, &e, &f, &g, &h, j+1, &w)
			round1(&a, &b, &c, &dVal, &e, &f, &g, &h, j+2, &w)
			round1(&a, &b, &c, &dVal, &e, &f, &g, &h, j+3, &w)
			round1(&a, &b, &c, &dVal, &e, &f, &g, &h, j+4, &w)
			round1(&a, &b, &c, &dVal, &e, &f, &g, &h, j+5, &w)
			round1(&a, &b, &c, &dVal, &e, &f, &g, &h, j+6, &w)
			round1(&a, &b, &c, &dVal, &e, &f, &g, &h, j+7, &w)
		}

		// 16-63 轮: FF2, GG2 (循环展开 8 轮一组)
		for j := 16; j < 64; j += 8 {
			round2(&a, &b, &c, &dVal, &e, &f, &g, &h, j, &w)
			round2(&a, &b, &c, &dVal, &e, &f, &g, &h, j+1, &w)
			round2(&a, &b, &c, &dVal, &e, &f, &g, &h, j+2, &w)
			round2(&a, &b, &c, &dVal, &e, &f, &g, &h, j+3, &w)
			round2(&a, &b, &c, &dVal, &e, &f, &g, &h, j+4, &w)
			round2(&a, &b, &c, &dVal, &e, &f, &g, &h, j+5, &w)
			round2(&a, &b, &c, &dVal, &e, &f, &g, &h, j+6, &w)
			round2(&a, &b, &c, &dVal, &e, &f, &g, &h, j+7, &w)
		}

		d.h[0] ^= a
		d.h[1] ^= b
		d.h[2] ^= c
		d.h[3] ^= dVal
		d.h[4] ^= e
		d.h[5] ^= f
		d.h[6] ^= g
		d.h[7] ^= h
		p = p[chunk:]
	}
}

// 辅助内联函数 - 轮处理 1
func round1(a, b, c, d, e, f, g, h *uint32, j int, w *[68]uint32) {
	a12 := (*a<<12 | *a>>(32-12))
	ss1 := a12 + *e + tRot[j]
	ss1 = ss1<<7 | ss1>>(32-7)
	ss2 := ss1 ^ a12
	tt1 := (*a ^ *b ^ *c) + *d + ss2 + (w[j] ^ w[j+4])
	tt2 := (*e ^ *f ^ *g) + *h + ss1 + w[j]
	*d = *c
	*c = *b<<9 | *b>>(32-9)
	*b = *a
	*a = tt1
	*h = *g
	*g = *f<<19 | *f>>(32-19)
	*f = *e
	*e = P0(tt2)
}

// 辅助内联函数 - 轮处理 2
func round2(a, b, c, d, e, f, g, h *uint32, j int, w *[68]uint32) {
	a12 := (*a<<12 | *a>>(32-12))
	ss1 := a12 + *e + tRot[j]
	ss1 = ss1<<7 | ss1>>(32-7)
	ss2 := ss1 ^ a12
	// FF2
	tt1 := ((*a & *b) | (*a & *c) | (*b & *c)) + *d + ss2 + (w[j] ^ w[j+4])
	// GG2: 使用 Rust 版本的快速位运算技巧 (y ^ z) & x ^ z
	tt2 := ((*f ^ *g) & *e ^ *g) + *h + ss1 + w[j]
	*d = *c
	*c = *b<<9 | *b>>(32-9)
	*b = *a
	*a = tt1
	*h = *g
	*g = *f<<19 | *f>>(32-19)
	*f = *e
	*e = P0(tt2)
}

func Sum(data []byte) [Size]byte {
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}
