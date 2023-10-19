/*
	Copyright 2023 Loophole Labs

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		   http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package polyglot

import (
	"errors"
	"math"
)

const (
	emptyString  = ""
	VarIntLen16  = 3
	VarIntLen32  = 5
	VarIntLen64  = 10
	continuation = 0x80
)

var (
	InvalidSlice   = errors.New("invalid slice encoding")
	InvalidMap     = errors.New("invalid map encoding")
	InvalidBytes   = errors.New("invalid bytes encoding")
	InvalidString  = errors.New("invalid string encoding")
	InvalidError   = errors.New("invalid error encoding")
	InvalidBool    = errors.New("invalid bool encoding")
	InvalidUint8   = errors.New("invalid uint8 encoding")
	InvalidUint16  = errors.New("invalid uint16 encoding")
	InvalidUint32  = errors.New("invalid uint32 encoding")
	InvalidUint64  = errors.New("invalid uint64 encoding")
	InvalidInt32   = errors.New("invalid int32 encoding")
	InvalidInt64   = errors.New("invalid int64 encoding")
	InvalidFloat32 = errors.New("invalid float32 encoding")
	InvalidFloat64 = errors.New("invalid float64 encoding")
)

func decodeNil(b []byte) ([]byte, bool) {
	if len(b) > 0 {
		if b[0] == NilRawKind {
			return b[1:], true
		}
	}
	return b, false
}

func decodeMap(b []byte, keyKind, valueKind Kind) ([]byte, uint32, error) {
	if len(b) > 2 {
		if b[0] == MapRawKind && b[1] == byte(keyKind) && b[2] == byte(valueKind) {
			var size uint32
			var err error
			b, size, err = decodeStaticUint32(b[3:])
			if err != nil {
				return b, 0, InvalidMap
			}
			return b, size, nil
		}
	}
	return b, 0, InvalidMap
}

func decodeSlice(b []byte, kind Kind) ([]byte, uint32, error) {
	if len(b) > 1 {
		if b[0] == SliceRawKind && b[1] == byte(kind) {
			var size uint32
			var err error
			b, size, err = decodeStaticUint32(b[2:])
			if err != nil {
				return b, 0, InvalidSlice
			}
			return b, size, nil
		}
	}
	return b, 0, InvalidSlice
}

func decodeBytes(b []byte, ret []byte) ([]byte, []byte, error) {
	if len(b) > 0 && b[0] == BytesRawKind {
		var size uint32
		var err error
		b, size, err = decodeStaticUint32(b[1:])
		if err != nil {
			return b, nil, InvalidBytes
		}
		if len(b) > int(size)-1 {
			if len(ret) < int(size) {
				if ret == nil {
					ret = make([]byte, size)
					copy(ret, b[:size])
				} else {
					ret = append(ret[:0], b[:size]...)
				}
			} else {
				copy(ret[0:], b[:size])
			}
			return b[size:], ret, nil
		}
	}
	return b, nil, InvalidBytes
}

func decodeString(b []byte) ([]byte, string, error) {
	if len(b) > 0 {
		if b[0] == StringRawKind {
			var size uint32
			var err error
			b, size, err = decodeStaticUint32(b[1:])
			if err != nil {
				return b, emptyString, InvalidString
			}
			if len(b) > int(size)-1 {
				return b[size:], string(b[:size]), nil
			}
		}
	}
	return b, emptyString, InvalidString
}

func decodeError(b []byte) ([]byte, error, error) {
	if len(b) > 0 {
		if b[0] == ErrorRawKind {
			var val string
			var err error
			b, val, err = decodeString(b[1:])
			if err != nil {
				return b, nil, InvalidError
			}
			return b, Error(val), nil
		}
	}
	return b, nil, InvalidError
}

func decodeBool(b []byte) ([]byte, bool, error) {
	if len(b) > 1 {
		if b[0] == BoolRawKind {
			if b[1] == trueBool {
				return b[2:], true, nil
			} else {
				return b[2:], false, nil
			}
		}
	}
	return b, false, InvalidBool
}

func decodeUint8(b []byte) ([]byte, uint8, error) {
	if len(b) > 1 {
		if b[0] == Uint8RawKind {
			return b[2:], b[1], nil
		}
	}
	return b, 0, InvalidUint8
}

// Variable integer encoding with the same format as binary.varint
// (https://developers.google.com/protocol-buffers/docs/encoding#varints)
func decodeUint16(b []byte) ([]byte, uint16, error) {
	if len(b) > 1 && b[0] == Uint16RawKind {
		var x uint16
		var s uint
		for i := 1; i < VarIntLen16+1; i++ {
			cb := b[i]
			// Check if msb is set signifying a continuation byte
			if cb < continuation {
				if i > VarIntLen16 && cb > 1 {
					return b, 0, InvalidUint32
				}
				// End of varint, add the last bits and advance the buffer
				return b[i+1:], x | uint16(cb)<<s, nil
			}
			x |= uint16(cb&(continuation-1)) << s
			s += 7
		}
	}
	return b, 0, InvalidUint16
}

func decodeUint32(b []byte) ([]byte, uint32, error) {
	if len(b) > 1 && b[0] == Uint32RawKind {
		cb := uint32(b[1])
		if cb < continuation {
			return b[2:], cb, nil
		}

		x := cb & (continuation - 1)
		cb = uint32(b[2])
		if cb < continuation {
			return b[3:], x | (cb << 7), nil
		}

		x |= (cb & (continuation - 1)) << 7
		cb = uint32(b[3])
		if cb < continuation {
			return b[4:], x | (cb << 14), nil
		}

		x |= (cb & (continuation - 1)) << 14
		cb = uint32(b[4])
		if cb < continuation {
			return b[5:], x | (cb << 21), nil
		}

		x |= (cb & (continuation - 1)) << 21
		cb = uint32(b[5])
		if cb < continuation {
			return b[6:], x | (cb << 28), nil
		}

		//count += more
		//cb = uint32(b[6])
		//x |= more * ((cb & (continuation - 1)) << 35)
		//more &= (cb & continuation) >> 7

		//return b, 0, InvalidUint32
		//
		//var x uint32
		//var s uint
		//for i := 1; i < VarIntLen32+1; i++ {
		//	cb := _b[i]
		//	// Check if msb is set signifying a continuation byte
		//	if cb < continuation {
		//		if i > VarIntLen32 && cb > 1 {
		//			return _b, 0, InvalidUint32
		//		}
		//		// End of varint, add the last bits and advance the buffer
		//		return _b[i+1:], x | uint32(cb)<<s, nil
		//	}
		//	// Add the lower 7 bits to the result and continue to the next byte
		//	x |= uint32(cb&(continuation-1)) << s
		//	s += 7
		//}
	}
	return b, 0, InvalidUint32
}

func decodeStaticUint32(b []byte) ([]byte, uint32, error) {
	if len(b) > 4 && b[0] == StaticUint32RawKind {
		return b[5:], uint32(b[4]) | uint32(b[3])<<8 | uint32(b[2])<<16 | uint32(b[1])<<24, nil
	}
	return b, 0, InvalidUint32
}

func decodeUint64(b []byte) ([]byte, uint64, error) {
	if len(b) > 1 && b[0] == Uint64RawKind {
		var x uint64
		var s uint
		for i := 1; i < VarIntLen64+1; i++ {
			cb := b[i]
			// Check if msb is set signifying a continuation byte
			if cb < continuation {
				// Check for overflow
				if i > VarIntLen64 && cb > 1 {
					return b, 0, InvalidUint64
				}
				// End of varint, add the last bits and advance the buffer
				return b[i+1:], x | uint64(cb)<<s, nil
			}
			// Add the lower 7 bits to the result and continue to the next byte
			x |= uint64(cb&(continuation-1)) << s
			s += 7
		}
	}
	return b, 0, InvalidUint64
}

func decodeInt32(b []byte) ([]byte, int32, error) {
	if len(b) > 1 && b[0] == Int32RawKind {
		cb := uint32(b[1])
		if cb < continuation {
			x := int32(cb >> 1)
			if cb&1 != 0 {
				x = -(x + 1)
			}
			return b[2:], x, nil
		}

		x := cb & (continuation - 1)
		cb = uint32(b[2])
		if cb < continuation {
			x |= cb << 7
			if x&1 != 0 {
				return b[3:], -(int32(x>>1) + 1), nil
			}
			return b[3:], int32(x >> 1), nil
		}

		x |= (cb & (continuation - 1)) << 7
		cb = uint32(b[3])
		if cb < continuation {
			x |= cb << 14
			if x&1 != 0 {
				return b[4:], -(int32(x>>1) + 1), nil
			}
			return b[4:], int32(x >> 1), nil
		}

		x |= (cb & (continuation - 1)) << 14
		cb = uint32(b[4])
		if cb < continuation {
			x |= cb << 21
			if x&1 != 0 {
				return b[5:], -(int32(x>>1) + 1), nil
			}
			return b[5:], int32(x >> 1), nil
		}

		x |= (cb & (continuation - 1)) << 21
		cb = uint32(b[5])
		if cb < continuation {
			x |= cb << 28
			if x&1 != 0 {
				return b[6:], -(int32(x>>1) + 1), nil
			}
			return b[6:], int32(x >> 1), nil
		}
	}
	return b, 0, InvalidInt32
}

func decodeInt64(b []byte) ([]byte, int64, error) {
	if len(b) > 1 && b[0] == Int64RawKind {
		var ux uint64
		var s uint
		for i := 1; i < VarIntLen64+1; i++ {
			cb := b[i]
			// Check if msb is set signifying a continuation byte
			if cb < continuation {
				if i > VarIntLen64 && cb > 1 {
					return b, 0, InvalidInt64
				}
				// End of varint, add the last bits
				ux |= uint64(cb) << s
				// Separate value and sign
				x := int64(ux >> 1)
				// If sign bit is set, negate the number
				if ux&1 != 0 {
					x = -(x + 1)
				}
				return b[i+1:], x, nil
			}
			ux |= uint64(cb&(continuation-1)) << s
			s += 7
		}
	}
	return b, 0, InvalidInt64
}

func decodeFloat32(b []byte) ([]byte, float32, error) {
	if len(b) > 4 && b[0] == Float32RawKind {
		return b[5:], math.Float32frombits(uint32(b[4]) | uint32(b[3])<<8 | uint32(b[2])<<16 | uint32(b[1])<<24), nil
	}
	return b, 0, InvalidFloat32
}

func decodeFloat64(b []byte) ([]byte, float64, error) {
	if len(b) > 8 && b[0] == Float64RawKind {
		return b[9:], math.Float64frombits(uint64(b[8]) | uint64(b[7])<<8 | uint64(b[6])<<16 | uint64(b[5])<<24 |
			uint64(b[4])<<32 | uint64(b[3])<<40 | uint64(b[2])<<48 | uint64(b[1])<<56), nil
	}
	return b, 0, InvalidFloat64
}
