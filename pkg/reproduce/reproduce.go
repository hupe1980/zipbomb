package reproduce

import (
	"bytes"
	"compress/flate"
	"fmt"
	"hash/crc32"
	"io"
)

type Options struct {
	GZip bool
}

// Make creates a self reproducing zip bomb.
func Make(w io.Writer, optFns ...func(o *Options)) error {
	opts := Options{
		GZip: false,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	if opts.GZip {
		return makeGz(w)
	}

	return makeZip(w)
}

func makeGz(w io.Writer) error {
	head := []byte{
		0x1f,                   // ID1
		0x8b,                   // ID2
		0x08,                   // Compression Method 0x08 => deflate
		0x08,                   // Flags only fname is set
		0x00, 0x00, 0x00, 0x00, // MTIME
		0x00,                                              // eXtra FLags
		0x00,                                              // Operating System
		'r', 'e', 'c', 'u', 'r', 's', 'i', 'v', 'e', 0x00, // file name, zero-terminated
	}

	zhead, err := deflate(head, true, false)
	if err != nil {
		return err
	}

	ztail := make([]byte, 5+8)
	ztail[0] = 1
	ztail[1] = 8
	ztail[2] = 0
	ztail[3] = ^byte(8)
	ztail[4] = ^byte(0)
	tail := ztail[5:]
	// CRC32
	tail[0] = 0xaa
	tail[1] = 0xbb
	tail[2] = 0xcc
	tail[3] = 0xdd

	whole, err := makeGeneric(zhead, head, ztail, tail, nil)
	if err != nil {
		return err
	}

	n := len(whole)
	// ISIZE
	tail[4] = byte(n)
	tail[5] = byte(n >> 8)
	tail[6] = byte(n >> 16)
	tail[7] = byte(n >> 24)

	whole, err = makeGeneric(zhead, head, ztail, tail, tail[0:4])
	if err != nil {
		return err
	}

	if _, err := w.Write(whole); err != nil {
		return err
	}

	return nil
}

func makeZip(w io.Writer) error {
	csize := 0
	uncsize := 0
	sufpos := 0
	zhead := []byte{
		0x00, 37, 0, ^byte(37), 0xFF, // 37-byte literal
		0x50, 0x4b, 0x03, 0x04, // ZHeader
		0x14,       // extvers
		0x00,       // extos
		0x00, 0x00, // flags
		0x08, 0x00, // method deflate
		0x08, 0x03, // modtime
		0x64, 0x3c, // moddate
		0xaa, 0xbb, 0xcc, 0xdd, // crc
		byte(csize), byte(csize >> 8), 0, 0, // csize
		byte(uncsize), byte(uncsize >> 8), 0, 0, // uncsize
		0x07, 0x00, // file name length
		0x00, 0x00, // extra field length
		'r', '/', 'r', '.', 'z', 'i', 'p', // file name
	}
	head := zhead[5:]
	headsize := head[14:26]
	tail := []byte{
		0x50, 0x4b, 0x01, 0x02, // ZCHeader
		0x14,       // madevers
		0x00,       // madeos
		0x14,       // extvers
		0x00,       // extos
		0x00, 0x00, // flags
		0x08, 0x00, // meth
		0x08, 0x03, // modtime
		0x64, 0x3c, // moddate
		0xaa, 0xbb, 0xcc, 0xdd, // crc
		byte(csize), byte(csize >> 8), 0, 0, // csize
		byte(uncsize), byte(uncsize >> 8), 0, 0, // uncsize
		0x07, 0x00, // flen
		0x00, 0x00, // xlen
		0x00, 0x00, // fclen
		0x00, 0x00, // disk start
		0x00, 0x00, // iattr
		0x00, 0x00, 0x00, 0x00, // eattr
		0x00, 0x00, 0x00, 0x00, // off
		'r', '/', 'r', '.', 'z', 'i', 'p', // file name
		0x50, 0x4b, 0x05, 0x06, // ZECHeader
		0x00, 0x00, // dn
		0x00, 0x00, // ds
		0x01, 0x00, // de
		0x01, 0x00, // entries
		53, 0x00, 0x00, 0x00, // size
		byte(sufpos), byte(sufpos >> 8), 0x00, 0x00, // off
		0x00, 0x00, // zclen
	}

	var (
		b    wbuf
		zero [12]byte
	)

	b.writeBits(0, 1, false)
	b.writeBits(1, 2, false)
	b.writeBits(0x50+48, 8, true)
	b.writeBits(0x4b+48, 8, true)
	b.writeBits(0x01+48, 8, true)
	b.writeBits(0x02+48, 8, true)
	b.writeBits(0x14+48, 8, true)
	b.writeBits(0x00+48, 8, true)
	b.writeBits(270-256, 7, true)
	b.writeBits(1, 2, false)
	b.writeBits(16, 5, true)
	b.writeBits(367-256-1, 7, false)
	b.writeBits(267-256, 7, true)
	b.writeBits(1, 1, false)
	b.writeBits(0, 5, true)
	b.writeBits('r'+48, 8, true) // file name
	b.writeBits('/'+48, 8, true)
	b.writeBits('r'+48, 8, true)
	b.writeBits('.'+48, 8, true)
	b.writeBits('z'+48, 8, true)
	b.writeBits('i'+48, 8, true)
	b.writeBits('p'+48, 8, true)
	b.writeBits(0x50+48, 8, true)
	b.writeBits(0x4b+48, 8, true)
	b.writeBits(0x05+48, 8, true)
	b.writeBits(0x06+48, 8, true)
	b.writeBits(4-2, 7, true)
	b.writeBits(7, 5, true)
	b.writeBits(3, 2, false)
	b.writeBits(0x01+48, 8, true)
	b.writeBits(3-2, 7, true)
	b.writeBits(2-1, 5, true)
	b.writeBits(53+48, 8, true) // size
	b.writeBits(0x00+48, 8, true)
	b.writeBits(0x00+48, 8, true)
	b.writeBits(0x00+48, 8, true)
	b.writeBits(0, 7, true)
	b.writeBits(1, 1, false)
	b.writeBits(0, 2, false)
	b.flushBits()
	b.bytes.WriteByte(6)
	b.bytes.WriteByte(0)
	b.bytes.WriteByte(^byte(6))
	b.bytes.WriteByte(^byte(0))
	tailsufOffset := b.bytes.Len()
	b.bytes.Write(zero[0:6])
	ztail := b.bytes.Bytes()
	tailsuf := ztail[tailsufOffset : tailsufOffset+4]

	whole, err := makeGeneric(zhead, head, ztail, tail, nil)
	if err != nil {
		return err
	}

	csize = len(whole) - len(head) - len(tail)
	uncsize = len(whole)
	headsize[4+0] = byte(csize)
	headsize[4+1] = byte(csize >> 8)
	headsize[8+0] = byte(uncsize)
	headsize[8+1] = byte(uncsize >> 8)
	tail[20] = byte(csize)
	tail[21] = byte(csize >> 8)
	tail[24] = byte(uncsize)
	tail[25] = byte(uncsize >> 8)
	sufpos = len(head) + csize
	tailsuf[0+0] = byte(sufpos)
	tailsuf[0+1] = byte(sufpos >> 8)
	tail[len(tail)-6+0] = byte(sufpos)
	tail[len(tail)-6+1] = byte(sufpos >> 8)

	whole, err = makeGeneric(zhead, head, ztail, tail, headsize[0:4])
	if err != nil {
		return err
	}

	if _, err := w.Write(whole); err != nil {
		return err
	}

	return nil
}

func makeGeneric(zhead, head, ztail, tail, crc []byte) ([]byte, error) {
	const unit = 5

	var b wbuf

	b.bytes.Write(zhead)

	// LITn+1 zhead LITn+1
	b.lit(len(zhead) + unit)
	b.bytes.Write(zhead)
	b.lit(len(zhead) + unit)

	// REPn+1
	b.rep(len(zhead) + unit)

	// LIT1 REPn+1
	b.lit(unit)
	b.rep(len(zhead) + unit)

	// LIT1 LIT1
	b.lit(unit)
	b.lit(unit)

	// LIT4 REPn+1 LIT1 LIT1 LIT4
	b.lit(4 * unit)
	b.rep(len(zhead) + unit)
	b.lit(unit)
	b.lit(unit)
	b.lit(4 * unit)

	// REP4
	b.rep(4 * unit)

	// LIT4 REP4 LIT4 REP4 LIT4
	b.lit(4 * unit)
	b.rep(4 * unit)
	b.lit(4 * unit)
	b.rep(4 * unit)
	b.lit(4 * unit)

	// REP4
	b.rep(4 * unit)

	// LIT4 REP4 NOP NOP LITm+1
	b.lit(4 * unit)
	b.rep(4 * unit)
	b.lit(0)
	b.lit(0)
	b.lit(len(ztail) + 2*unit)

	// REP4
	b.rep(4 * unit)

	// NOP NOP LITm+1 REPm+1 suffix
	b.lit(0)
	b.lit(0)
	b.lit(len(ztail) + 2*unit)
	b.rep(len(ztail) + 2*unit)
	b.lit(0)
	b.bytes.Write(ztail)

	// REPm+1
	b.rep(len(ztail) + 2*unit)

	// suffix
	b.lit(0)
	b.bytes.Write(ztail)
	out := b.bytes.Bytes()

	var whole []byte
	// double-check
	{
		r := flate.NewReader(bytes.NewBuffer(out))

		var b1 bytes.Buffer
		// nolint gosec
		if _, err := io.Copy(&b1, r); err != nil {
			return nil, err
		}

		if err := r.Close(); err != nil {
			return nil, err
		}

		var b2 bytes.Buffer
		b2.Write(head)
		b2.Write(out)
		b2.Write(tail)

		if !bytes.Equal(b1.Bytes(), b2.Bytes()) {
			return nil, fmt.Errorf("failed double-check")
		}

		whole = b1.Bytes()
	}

	if crc != nil {
		n := bytes.Count(whole, crc)
		embed := make([]int, n)
		off := 0

		for i := 0; i < n; i++ {
			j := bytes.Index(whole[off:], crc)
			off += j
			embed[i] = off
			off += 4
		}

		crc0 := uint32(0)
		crcbase := crc32.ChecksumIEEE(whole[0:embed[0]])

		for {
			if crc0&0xfffff == 0 {
				//PROGRESS
				fmt.Printf("%#f%%\r", 100*float64(crc0)/0xffffffff)
			}

			for _, i := range embed {
				whole[i+0] = byte(crc0)
				whole[i+1] = byte(crc0 >> 8)
				whole[i+2] = byte(crc0 >> 16)
				whole[i+3] = byte(crc0 >> 24)
			}

			crc1 := crc32.Update(crcbase, crc32.IEEETable, whole[embed[0]:])
			if crc0 == crc1 {
				break
			}

			crc0++
		}

		fmt.Printf("SUCCESS   \n")
	}

	// double double-check
	{
		r := flate.NewReader(bytes.NewBuffer(whole[len(head) : len(head)+len(out)]))

		var b1 bytes.Buffer
		// nolint gosec
		if _, err := io.Copy(&b1, r); err != nil {
			return nil, err
		}

		if err := r.Close(); err != nil {
			return nil, err
		}

		if !bytes.Equal(b1.Bytes(), whole) {
			return nil, fmt.Errorf("failed double double-check")
		}

		whole = b1.Bytes()
	}

	return whole, nil
}

type wbuf struct {
	bytes bytes.Buffer
	bit   uint32
	nbit  uint
	final uint32
}

func (b *wbuf) writeBits(bit uint32, nbit uint, rev bool) {
	if rev {
		br := uint32(0)

		for i := uint(0); i < nbit; i++ {
			if bit&(1<<i) != 0 {
				br |= 1 << (nbit - 1 - i)
			}
		}

		bit = br
	}

	b.bit |= bit << b.nbit

	b.nbit += nbit

	for b.nbit >= 8 {
		b.bytes.WriteByte(byte(b.bit))
		b.bit >>= 8
		b.nbit -= 8
	}
}

func (b *wbuf) flushBits() {
	if b.nbit > 0 {
		b.bytes.WriteByte(byte(b.bit))
		b.nbit = 0
		b.bit = 0
	}
}

func (b *wbuf) lit(n int) {
	b.writeBits(b.final, 1, false)
	b.writeBits(0, 2, false)
	b.flushBits()

	b1 := byte(n)

	b2 := byte(n >> 8)

	b.bytes.WriteByte(b1)
	b.bytes.WriteByte(b2)
	b.bytes.WriteByte(^b1)
	b.bytes.WriteByte(^b2)
}

func (b *wbuf) rep(n int) {
	b.writeBits(b.final, 1, false)
	b.writeBits(1, 2, false)

	steal := uint(0)

	switch {
	case 9 <= n && n <= 12:
		b.writeBits(uint32(254+n/2)-256, 7, true)
		b.writeBits(6, 5, true)
		b.writeBits(uint32(n-8-1), 2, false)
		b.writeBits(uint32(254+n-n/2)-256, 7, true)
		b.writeBits(6, 5, true)
		b.writeBits(uint32(n-8-1), 2, false)
	case 13 <= n && n <= 16:
		b.writeBits(uint32(254+n/2)-256, 7, true)
		b.writeBits(7, 5, true)
		b.writeBits(uint32(n-12-1), 2, false)
		b.writeBits(uint32(254+n-n/2)-256, 7, true)
		b.writeBits(7, 5, true)
		b.writeBits(uint32(n-12-1), 2, false)
	case 17 <= n && n <= 20:
		b.writeBits(uint32(254+n/2)-256, 7, true)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
		b.writeBits(uint32(254+n-n/2)-256, 7, true)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
	case n == 21:
		b.writeBits(uint32(254+10)-256, 7, true)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
		b.writeBits(uint32(265)-256, 7, true)
		b.writeBits(0, 1, true)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)

		steal = 1
	case 22 <= n && n <= 24:
		b.writeBits(uint32(265+(n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n/2-11)&1, 1, false)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
		b.writeBits(uint32(265+(n-n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n-n/2-11)&1, 1, false)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)

		steal = 2
	case 25 <= n && n <= 32:
		b.writeBits(uint32(265+(n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n/2-11)&1, 1, false)
		b.writeBits(9, 5, true)
		b.writeBits(uint32(n-24-1), 3, false)
		b.writeBits(uint32(265+(n-n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n-n/2-11)&1, 1, false)
		b.writeBits(9, 5, true)
		b.writeBits(uint32(n-24-1), 3, false)

		steal = 2
	case 33 <= n && n <= 36:
		b.writeBits(uint32(265+(n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n/2-11)&1, 1, false)
		b.writeBits(10, 5, true)
		b.writeBits(uint32(n-32-1), 4, false)
		b.writeBits(uint32(265+(n-n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n-n/2-11)&1, 1, false)
		b.writeBits(10, 5, true)
		b.writeBits(uint32(n-32-1), 4, false)

		steal = 4
	case 37 <= n && n <= 48:
		b.writeBits(uint32(265+(18-11)>>1)-256, 7, true)
		b.writeBits(uint32(18-11)&1, 1, false)
		b.writeBits(10, 5, true)
		b.writeBits(uint32(n-32-1), 4, false)
		b.writeBits(uint32(269+(n-18-19)>>2)-256, 7, true)
		b.writeBits(uint32(n-18-19)&3, 2, false)
		b.writeBits(10, 5, true)
		b.writeBits(uint32(n-32-1), 4, false)

		steal = 5
	case 49 <= n && n <= 64:
		b.writeBits(uint32(254+10)-256, 7, true)
		b.writeBits(11, 5, true)
		b.writeBits(uint32(n-48-1), 4, false)
		b.writeBits(uint32(273+(n-10-35)>>3)-256, 7, true)
		b.writeBits(uint32(n-10-35)&7, 3, false)
		b.writeBits(11, 5, true)
		b.writeBits(uint32(n-48-1), 4, false)

		steal = 5
	default:
		panic("cannot encode REP")
	}

	b.writeBits(0, 7-steal, true)
}

var inflateO, inflateB int

func deflate(data []byte, litNext bool, final bool) ([]byte, error) {
	var buf bytes.Buffer

	w, err := flate.NewWriter(&buf, 9)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	z := buf.Bytes()

	if final {
		return z, nil
	}

	b1 := bytes.NewBuffer(z)

	var b2 bytes.Buffer

	r := flate.NewReader(b1)

	// nolint gosec
	if _, err := io.Copy(&b2, r); err != nil {
		return nil, err
	}

	if err := r.Close(); err != nil {
		return nil, err
	}

	if inflateB == 0 {
		return z[0:inflateO], nil
	}

	z[inflateO] ^= 1 << uint(inflateB)

	if litNext {
		if inflateB >= 6 && len(z) == inflateO+1+5 && z[inflateO+1] == 0 && z[inflateO+2] == 0 && z[inflateO+3] == 0 && z[inflateO+4] == 0xff && z[inflateO+5] == 0xff {
			return z[0 : inflateO+1], nil
		}

		if inflateB <= 5 && z[inflateO] == 0 {
			return z[0:inflateO], nil
		}
	}

	return z, nil
}
