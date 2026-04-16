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

var (
	iv = [8]uint32{
		0x7380166f, 0x4914b2b9, 0x172442d7, 0xda8a0600,
		0xa96f30bc, 0x163138aa, 0xe38dee4d, 0xb0fb0e4e,
	}
)

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
		a, b, c, e, f, g := d.h[0], d.h[1], d.h[2], d.h[4], d.h[5], d.h[6]
		d3, d7 := d.h[3], d.h[7]

		for j := 0; j < 16; j++ {
			w[j] = binary.BigEndian.Uint32(p[j*4:])
		}
		for j := 16; j < 68; j++ {
			w[j] = P1(w[j-16]^w[j-9]^(w[j-3]<<15|w[j-3]>>(32-15))) ^ (w[j-13]<<7 | w[j-13]>>(32-7)) ^ w[j-6]
		}

		for j := 0; j < 16; j++ {
			t := uint32(0x79cc4519)
			ss1 := (a<<12 | a>>(32-12)) + e + (t<<uint(j) | t>>(32-uint(j)))
			ss1 = ss1<<7 | ss1>>(32-7)
			ss2 := ss1 ^ (a<<12 | a>>(32-12))
			tt1 := (a ^ b ^ c) + d3 + ss2 + (w[j] ^ w[j+4])
			tt2 := (e ^ f ^ g) + d7 + ss1 + w[j]
			d3, c, b, a = c, b<<9|b>>(32-9), a, tt1
			d7, g, f, e = g, f<<19|f>>(32-19), e, P0(tt2)
		}
		for j := 16; j < 64; j++ {
			t := uint32(0x7a879d8a)
			r := uint(j % 32)
			ss1 := (a<<12 | a>>(32-12)) + e + (t<<r | t>>(32-r))
			ss1 = ss1<<7 | ss1>>(32-7)
			ss2 := ss1 ^ (a<<12 | a>>(32-12))
			tt1 := ((a & b) | (a & c) | (b & c)) + d3 + ss2 + (w[j] ^ w[j+4])
			tt2 := ((e & f) | (^e & g)) + d7 + ss1 + w[j]
			d3, c, b, a = c, b<<9|b>>(32-9), a, tt1
			d7, g, f, e = g, f<<19|f>>(32-19), e, P0(tt2)
		}

		d.h[0] ^= a
		d.h[1] ^= b
		d.h[2] ^= c
		d.h[3] ^= d3
		d.h[4] ^= e
		d.h[5] ^= f
		d.h[6] ^= g
		d.h[7] ^= d7
		p = p[chunk:]
	}
}

func Sum(data []byte) [Size]byte {
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}
