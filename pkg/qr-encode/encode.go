package qr_encode

import (
	"bytes"
	"github.com/psxzz/go-qr/pkg/algorithms"
	"go.uber.org/multierr"
)

func NewEncoder(level CodeLevel) *Encoder {
	return &Encoder{level: level}
}

func (e *Encoder) Encode(text string) ([]byte, error) {
	byteLen := len(text)
	codeVersion, err := e.getVersion(byteLen)
	if err != nil {
		return nil, multierr.Combine(ErrVersionNotFound, err)
	}
	e.version = codeVersion

	currBuff := bytes.NewBuffer(make([]byte, 0))
	e.fillBuffer(currBuff, []byte(text))

	blocks := e.divideIntoBlocks(currBuff)
	correctionBlocks := e.generateCorrectionBlocks(blocks)
	result := e.mergeBlocks(blocks, correctionBlocks)

	return result, nil
}

func (e *Encoder) getVersion(byteLen int) (int, error) {
	bitLen := byteLen*8 + 4 // nolint:gomnd
	versionsArray := versionSize[e.level]
	version, err := algorithms.LowerBound(versionsArray[:], bitLen)
	if err != nil {
		return -1, err
	}
	maxVersionSize := versionSize[e.level][version]
	if version < 9 { // nolint:gomnd
		bitLen += 8
	} else {
		bitLen += 16
	}
	if bitLen > maxVersionSize {
		version++
		if version >= len(versionSize[e.level]) {
			return -1, ErrTooLargeSize
		}
	}
	return version, nil
}

// nolint:gomnd
func (e *Encoder) fillBuffer(buff *bytes.Buffer, data []byte) {
	dataLen := len(data)
	var currByte byte

	if e.version < 9 {
		buff.WriteByte(byte((headerNibble << 4) | (dataLen & Nible)))
		currByte = byte((dataLen >> 4) & Nible)
	} else {
		buff.WriteByte(byte((headerNibble << 4) | (dataLen & Nible)))
		buff.WriteByte(byte((dataLen >> 4) & Byte))
		currByte = byte((dataLen >> 12) & Nible)
	}

	for _, b := range data {
		currByte = (currByte << 4) | (b & byte(Nible))
		buff.WriteByte(currByte)
		currByte = (b >> 4) & byte(Nible)
	}
	currByte <<= 4
	buff.WriteByte(currByte)

	idx := 0
	currByte = FillerBytes[idx]
	for buff.Len()*8 < versionSize[e.level][e.version] {
		buff.WriteByte(currByte)
		idx = (idx + 1) % 2
		currByte = FillerBytes[idx]
	}
}

func (e *Encoder) divideIntoBlocks(buff *bytes.Buffer) [][]byte {
	blocksNum := numberOfBlocks[e.level][e.version]
	blockSize := buff.Len() / blocksNum
	rem := buff.Len() % blocksNum
	result := make([][]byte, blocksNum)

	data := buff.Bytes()
	currIdx := 0
	for i := 0; i < blocksNum-rem; i++ {
		result[i] = data[currIdx : currIdx+blockSize]
		currIdx += blockSize
	}
	for i := 0; i < rem; i++ {
		result[blocksNum-rem+i] = data[currIdx : currIdx+blockSize+1]
		currIdx += blockSize + 1
	}

	return result
}

// nolint:gomnd
func (e *Encoder) generateCorrectionBlocks(dataBlocks [][]byte) [][]byte {
	coefficientsNum := numberOfCorrectionBytes[e.level][e.version]
	coefficients := polynomialCoefficients[coefficientsNum]

	result := make([][]byte, len(dataBlocks))
	for blockIdx, block := range dataBlocks {
		correctionBytesNum := algorithms.MaxInt(len(block), coefficientsNum)
		correctionBytes := make([]byte, 0)
		correctionBytes = append(correctionBytes, block...)
		for i := len(correctionBytes); i < correctionBytesNum+len(block); i++ {
			correctionBytes = append(correctionBytes, 0)
		}

		for i, el := range correctionBytes[:len(block)] {
			if el == 0 {
				continue
			}

			b := int(invGF[el])
			for j := 0; j < coefficientsNum; j++ {
				c := (coefficients[j] + b) % 255
				t := GF[c]
				correctionBytes[i+j+1] ^= t
			}
		}
		result[blockIdx] = correctionBytes[len(correctionBytes)-coefficientsNum:]
	}

	return result
}

func (e *Encoder) mergeBlocks(blocks [][]byte, correctionBlocks [][]byte) []byte {
	result := bytes.NewBuffer(make([]byte, 0))

	maxBlockSize := 0
	for _, block := range blocks {
		maxBlockSize = algorithms.MaxInt(maxBlockSize, len(block))
	}

	currByteIdx := 0
	for currByteIdx < maxBlockSize {
		for i := 0; i < len(blocks); i++ {
			if currByteIdx >= len(blocks[i]) {
				continue
			}
			result.WriteByte(blocks[i][currByteIdx])
		}
		currByteIdx++
	}

	currByteIdx = 0
	for currByteIdx < maxBlockSize {
		for i := 0; i < len(correctionBlocks); i++ {
			if currByteIdx >= len(correctionBlocks[i]) {
				continue
			}
			result.WriteByte(correctionBlocks[i][currByteIdx])
		}
		currByteIdx++
	}

	return result.Bytes()
}