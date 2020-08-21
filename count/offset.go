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

const OffsetSize = 4

type Offset int32
type Offsets struct {
	Entries []Offset
	Offset  int32
}

func NewOffsets(reader io.ReadSeeker, offset int32, tableLength int32) (Offsets, error) {
	o := Offsets{
		Entries: make([]Offset, tableLength),
		Offset:  offset,
	}

	if _, err := reader.Seek(int64(offset), io.SeekStart); err != nil {
		return Offsets{}, fmt.Errorf("unable to seek to beginning of offsets\n%w", err)
	}

	for i := 0; i < len(o.Entries); i++ {
		if err := binary.Read(reader, binary.LittleEndian, &o.Entries[i]); err != nil {
			return Offsets{}, fmt.Errorf("unable to read offset\n%w", err)
		}
	}

	return o, nil
}

func (o Offsets) Size() int {
	return len(o.Entries) * OffsetSize
}
