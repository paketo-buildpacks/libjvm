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

const (
	HeaderSlots = 7
	HeaderSize  = 4
)

type Header struct {
	Magic         int32
	MajorVersion  int32
	MinorVersion  int32
	Flags         int32
	ResourceCount int32
	TableLength   int32
	LocationsSize int32
	StringsSize   int32
}

func NewHeader(reader io.Reader) (Header, error) {
	var h Header

	if err := binary.Read(reader, binary.LittleEndian, &h.Magic); err != nil {
		return Header{}, fmt.Errorf("unable to read magic\n%w", err)
	}

	var version int32
	if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
		return Header{}, fmt.Errorf("unable to read version\n%w", err)
	}
	h.MajorVersion = version >> 16
	h.MinorVersion = version & 0xFFFF

	if err := binary.Read(reader, binary.LittleEndian, &h.Flags); err != nil {
		return Header{}, fmt.Errorf("unable to read flags\n%w", err)
	}

	if err := binary.Read(reader, binary.LittleEndian, &h.ResourceCount); err != nil {
		return Header{}, fmt.Errorf("unable to read resource count\n%w", err)
	}

	if err := binary.Read(reader, binary.LittleEndian, &h.TableLength); err != nil {
		return Header{}, fmt.Errorf("unable to read table length\n%w", err)
	}

	if err := binary.Read(reader, binary.LittleEndian, &h.LocationsSize); err != nil {
		return Header{}, fmt.Errorf("unable to read locations size\n%w", err)
	}

	if err := binary.Read(reader, binary.LittleEndian, &h.StringsSize); err != nil {
		return Header{}, fmt.Errorf("unable to read strings size\n%w", err)
	}

	return h, nil
}

func (Header) Size() uint32 {
	return HeaderSlots * HeaderSize
}
