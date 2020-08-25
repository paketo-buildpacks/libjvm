/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package count

import (
	"encoding/binary"
	"fmt"
	"io"
)

type StringReader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type Strings struct {
	Offset int32
	Size   int32
	Reader StringReader
}

func (s Strings) Get(offset int32) (string, error) {
	if _, err := s.Reader.Seek(int64(s.Offset)+int64(offset), io.SeekStart); err != nil {
		return "", fmt.Errorf("unable to seek to beginning of string\n%w", err)
	}

	length := 0
	for {
		var c byte
		if err := binary.Read(s.Reader, binary.LittleEndian, &c); err != nil {
			return "", fmt.Errorf("unable to determine string length\n%w", err)
		}

		if c == 0 {
			break
		}

		if (c & 0xC0) != 0x80 {
			length++
		}
	}

	b := make([]byte, length)
	if n, err := s.Reader.ReadAt(b, int64(s.Offset)+int64(offset)); err != nil {
		return "", fmt.Errorf("unable to read string\n%w", err)
	} else if n != length {
		return "", fmt.Errorf("unable to read full string")
	}

	return string(b), nil
}
