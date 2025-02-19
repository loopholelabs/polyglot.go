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

const (
	defaultSize = 512
)

type Buffer struct {
	b      []byte
	offset int
}

func NewBuffer() *Buffer {
	return &Buffer{
		b:      make([]byte, defaultSize),
		offset: 0,
	}
}

func NewBufferSize(size int) *Buffer {
	return &Buffer{
		b:      make([]byte, size),
		offset: 0,
	}
}

func NewBufferFromBytes(b []byte) *Buffer {
	return &Buffer{
		b:      b,
		offset: 0,
	}
}

func (buf *Buffer) Reset() {
	buf.offset = 0
}

func (buf *Buffer) MoveOffset(offset int) {
	buf.offset += offset
}

func (buf *Buffer) Grow(n int) {
	if cap(buf.b)-buf.offset < n {
		if cap(buf.b) < n {
			buf.b = append(buf.b[:buf.offset], make([]byte, n+cap(buf.b)-buf.offset)...)
		} else {
			buf.b = append(buf.b[:buf.offset], make([]byte, cap(buf.b)*2-buf.offset)...)
		}
		// Grow buffer length to match its new capacity.
		buf.b = buf.b[:cap(buf.b)]
	}
}

func (buf *Buffer) Write(b []byte) int {
	buf.Grow(len(b))
	buf.offset += copy(buf.b[buf.offset:cap(buf.b)], b)
	return len(b)
}

func (buf *Buffer) Bytes() []byte {
	return buf.b[:buf.offset]
}

func (buf *Buffer) Len() int {
	return buf.offset
}

func (buf *Buffer) Cap() int {
	return cap(buf.b)
}
