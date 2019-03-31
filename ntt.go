package nttp

const (
	BlockSizeLog  = 8
	BlockSize     = 1 << BlockSizeLog
	Prime         = 257
	PrimitiveRoot = 3
	MaxByte       = 1 << 8
)

// A naive exponentiation implementation
func quickPower(x, y int) int {
	ret := 1
	for ; y > 0; y >>= 1 {
		if y&1 == 1 {
			ret = ret * x % Prime
		}
		x = x * x % Prime
	}
	return ret
}

// Use a struct to store some runtime constants
type NTT struct {
	rader        [BlockSize]int
	temp         [BlockSize]int // Use a temp array to avoid allocation
	omega        [BlockSizeLog]int
	omegaInv     [BlockSizeLog]int
	blockSizeInv int
}

func (ntt *NTT) NTT(x []byte) []byte {
	// This avoids extra memory allocation
	y := ntt.temp[:]
	for i := 0; i < BlockSize; i++ {
		y[ntt.rader[i]] = int(x[i])
	}
	for k := 0; k < BlockSizeLog; k++ {
		half := 1 << uint(k)
		omega := ntt.omega[k]
		span := half << 1
		for l := 0; l < BlockSize; l += span {
			w, mid := 1, l+half
			for i := 0; i < half; i++ {
				t := w * y[mid+i] % Prime
				y[mid+i] = (y[l+i] - t) % Prime
				y[l+i] = (y[l+i] + t) % Prime
				w = w * omega % Prime
			}
		}
	}
	overflow := 0
	for i := 0; i < BlockSize; i++ {
		if y[i] < 0 {
			y[i] += Prime
		}
		if y[i] >= MaxByte {
			overflow += 1
		}
	}
	ret := make([]byte, 1+overflow+BlockSize)
	ret[0] = byte(overflow)
	pos := 1
	for i := 0; i < BlockSize; i++ {
		if y[i] >= MaxByte {
			y[i] -= MaxByte
			ret[pos] = byte(i)
			pos += 1
		}
		ret[1+overflow+i] = byte(y[i])
	}
	return ret
}

func (ntt *NTT) INTT(x []byte) []byte {
	y := ntt.temp[:]
	overflow := int(x[0])
	for i := 0; i < BlockSize; i++ {
		y[ntt.rader[i]] = int(x[1+overflow+i])
	}
	for i := 1; i <= overflow; i++ {
		y[ntt.rader[x[i]]] += MaxByte
	}
	for k := 0; k < BlockSizeLog; k++ {
		half := 1 << uint(k)
		omega := ntt.omegaInv[k]
		span := half << 1
		for l := 0; l < BlockSize; l += span {
			w, mid := 1, l+half
			for i := 0; i < half; i++ {
				t := w * y[mid+i] % Prime
				y[mid+i] = (y[l+i] - t) % Prime
				y[l+i] = (y[l+i] + t) % Prime
				w = w * omega % Prime
			}
		}
	}
	ret := make([]byte, BlockSize)
	for i := 0; i < BlockSize; i++ {
		if y[i] < 0 {
			y[i] += Prime
		}
		ret[i] = byte(y[i] * ntt.blockSizeInv % Prime)
	}
	return ret
}

func NewNTT() *NTT {
	ntt := new(NTT)
	ntt.rader[0] = 0
	ntt.blockSizeInv = quickPower(BlockSize, Prime-2)
	for i := 1; i < BlockSize; i++ {
		ntt.rader[i] = (ntt.rader[i>>1] >> 1) | ((i & 1) << (BlockSizeLog - 1))
	}
	for i := 0; i < BlockSizeLog; i++ {
		ntt.omega[i] = quickPower(PrimitiveRoot, (Prime-1)/(2<<uint(i)))
		ntt.omegaInv[i] = quickPower(ntt.omega[i], Prime-2)
	}
	return ntt
}
