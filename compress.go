package lznt1

// https://github.com/you0708/lznt1/blob/master/lznt1.py
// 简单的翻译为golang

import (
	"bytes"
	"encoding/binary"
)

// 防止数组越界
func min(a, b int) int {
	if a > b {
		return b
	}

	return a
}

// 在源slice中查找目标slice，返回的结果为 偏移 与所包含目标的最大长度
func find(src []byte, target []byte, maxLen int) (int, int) {
	resultOffset := 0
	resultLength := 0

	for i := 1; i < maxLen; i++ {
		offset := bytes.LastIndex(src, target[:i])
		if offset == -1 {
			break
		}

		tempOffset := len(src) - offset
		tempLength := i

		if tempOffset == tempLength {
			temp := bytes.Repeat(src[offset:], 0xfff/len(src[offset:])+1)
			for j := 0; j < maxLen+1; j++ {
				offset = bytes.LastIndex(temp, target[:min(j, len(target))])
				if offset == -1 {
					break
				}
				tempLength = j
			}
		}

		if tempLength > resultLength {
			resultOffset = tempOffset
			resultLength = tempLength
		}
	}

	if resultLength < 3 {
		return 0, 0
	}

	return resultOffset, resultLength
}

// 压缩块
func compressChunk(chunk []byte) []byte {
	blob := make([]byte, len(chunk))
	copy(blob, chunk)
	out := new(bytes.Buffer)
	pow2 := 0x10
	lMask3 := 0x1002
	oShift := 12

	for len(blob) > 0 {
		bits := 0
		temp := new(bytes.Buffer)

		// 这里有个坑，底下的取值范围为 0-7 但是循环结束的时候i的值是8,所以要--
		i := 0
		for i = 0; i < 8; i++ {
			bits = bits >> 1
			for pow2 < len(chunk)-len(blob) {
				pow2 = pow2 << 1
				lMask3 = (lMask3 >> 1) + 1
				oShift = oShift - 1
			}

			maxLen := 0
			if len(blob) < lMask3 {
				maxLen = len(blob)
			} else {
				maxLen = lMask3
			}

			offset, length := find(chunk[:len(chunk)-len(blob)], blob, maxLen)

			// try to find more compressed pattern
			_, length2 := find(chunk[:len(chunk)-len(blob)+1], blob[1:], maxLen)
			if length < length2 {
				length = 0
			}

			if length > 0 {
				symbol := ((offset - 1) << oShift) | (length - 3)
				bs := make([]byte, 2)
				binary.LittleEndian.PutUint16(bs, uint16(symbol))
				temp.Write(bs)
				bits = bits | 0x80
				blob = blob[length:]
			} else {
				temp.WriteByte(blob[0])
				blob = blob[1:]
			}

			if len(blob) == 0 {
				break
			}
		}

		if i == 8 {
			i--
		}

		out.WriteByte(byte(bits >> uint32(7 - i)))
		out.Write(temp.Bytes())
	}

	return out.Bytes()
}


// CHUNK_SIZE 0x1000 // to be compatible with all known forms of Windows
func compress(buf []byte, chunkSize int) []byte {
	out := new(bytes.Buffer)
	for len(buf) > 0 {
		chunk := buf[:min(len(buf), chunkSize)]
		compressed := compressChunk(chunk)
		// chunk is compressed
		if len(compressed) < len(chunk) {
			flags := 0xb000
			header := make([]byte, 2)
			binary.LittleEndian.PutUint16(header, uint16(flags|(len(compressed)-1)))
			out.Write(header)
			out.Write(compressed)
		} else {
			flags := 0x3000
			header := make([]byte, 2)
			binary.LittleEndian.PutUint16(header, uint16(flags|(len(chunk)-1)))
			out.Write(header)
			out.Write(chunk)
		}

		buf = buf[min(len(buf), chunkSize):]
	}

	return out.Bytes()
}

func Compress(buf []byte) []byte {
	return compress(buf, 0x1000)
}