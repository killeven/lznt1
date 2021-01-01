package lznt1

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func decompressChunk(chunk []byte) []byte {
	out := new(bytes.Buffer)

	for len(chunk) > 0 {
		flags := chunk[0]
		chunk = chunk[1:]

		for i := 0; i < 8; i++ {
			if ((flags >> i) & 1) == 0 {
				out.WriteByte(chunk[0])
				chunk = chunk[1:]
			} else {
				flag := binary.LittleEndian.Uint16(chunk[0:2])
				pos := out.Len() - 1
				lMask := 0xfff
				oShift := 12

				for pos >= 0x10 {
					lMask = lMask >> 1
					oShift = oShift - 1
					pos = pos >> 1
				}

				length := (uint32(flag) & uint32(lMask)) + 3
				offset := (uint32(flag) >> oShift) + 1
				index := out.Len() - int(offset)

				if length >= offset {
					temp := bytes.Repeat(out.Bytes()[index:], 0xfff/len(out.Bytes()[index:])+1)
					out.Write(temp[:length])
				} else {
					out.Write(out.Bytes()[index: index + int(length)])
				}
				chunk = chunk[2:]
			}

			if len(chunk) == 0 {
				break
			}
		}
	}

	return out.Bytes()
}

func Decompress(buf []byte, checkLength bool) ([]byte, error) {
	out := new(bytes.Buffer)

	for len(buf) != 0 {
		header := binary.LittleEndian.Uint16(buf[:2])
		length := (header & 0xfff) + 1
		if checkLength && (int(length) > len(buf[2:])) {
			return nil, errors.New("invalid chunk length")
		} else {
			chunk := buf[2 : 2+length]
			if (header & 0x8000) > 0 {
				out.Write(decompressChunk(chunk))
			} else {
				out.Write(chunk)
			}
		}

		buf = buf[2+length:]
	}

	return out.Bytes(), nil
}
