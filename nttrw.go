package nttp

import (
	"fmt"
	"io"
)

const (
	bufBlocks     = 16
	BlockDataSize = BlockSize - 1
)

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

type NTTReadWriter struct {
	rw   io.ReadWriter
	buf  [bufBlocks * (BlockSize + 2)]byte
	nbuf int
	ntt  *NTT
}

// Creates an NTT ReadWriter
func newNTTReadWriter(rw io.ReadWriter) *NTTReadWriter {
	return &NTTReadWriter{
		rw:  rw,
		ntt: NewNTT(),
	}
}

// Read data from underlying ReadWriterCloser
// Currently it should be ensured that len(p) >= BlockDataSize!
func (s *NTTReadWriter) Read(p []byte) (n int, err error) {
	// In the worst case we have to put BlockSize-1 bytes in to the buffer, so p should be big enough
	// TODO: Better handle this case, maybe adding some extra buffer to store partial decoded blocks?
	if len(p) < BlockDataSize {
		return 0, fmt.Errorf("buffer too small")
	}
	decodeFromBuffer := func() {
		ptr := 0 // Pointing to the beginning of next block in s.buf
		for {
			overflow := int(s.buf[ptr])
			blockLen := 1 + overflow + BlockSize
			// In the worst case every decoded block consists of BlockSize-1 bytes, we might not be able to put the data
			// into p if n+BlockSize-1 > len(p)
			if ptr+blockLen > s.nbuf || n+BlockDataSize > len(p) {
				copy(s.buf[:], s.buf[ptr:s.nbuf])
				s.nbuf -= ptr
				break
			}
			data := s.ntt.INTT(s.buf[ptr : ptr+blockLen])
			dataLen := int(data[0])
			copy(p[n:], data[1:1+dataLen])
			ptr += blockLen
			n += dataLen
			if ptr == s.nbuf {
				s.nbuf = 0
				break
			}
		}
	}
	// If there's some whole blocks in buf (with length at least BlockSize + 1), try to decode them and
	// put them in p first
	if s.nbuf > BlockSize {
		decodeFromBuffer()
	}
	if n+BlockDataSize <= len(p) {
		nread, err := s.rw.Read(s.buf[s.nbuf:])
		if err != nil && err != io.EOF {
			return n, err
		}
		s.nbuf += nread
		if s.nbuf > BlockSize {
			decodeFromBuffer()
		}
		if err == io.EOF {
			return n, err
		}
	}
	return
}

// Writes NTTed data into underlying ReadWriterCloser
// len(p) should best be multiples of BlockDataSize so there's no wasted bytes
func (s *NTTReadWriter) Write(p []byte) (n int, err error) {
	block := make([]byte, BlockSize)
	for n < len(p) {
		dataLen := min(BlockDataSize, len(p)-n)
		block[0] = byte(dataLen)
		copy(block[1:], p[n:n+dataLen])
		if _, err = s.rw.Write(s.ntt.NTT(block)); err != nil {
			return
		}
		n += dataLen
	}
	return
}
