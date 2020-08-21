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
	AttributeEnd byte = iota
	AttributeModule
	AttributeParent
	AttributeBase
	AttributeExtension
	AttributeOffset
	AttributeCompressed
	AttributeUncompressed
)

type Location struct {
	ModuleOffset     int32
	ParentOffset     int32
	BaseOffset       int32
	ExtensionOffset  int32
	ContentOffset    int32
	CompressedSize   int32
	UncompressedSize int32
}

func (l Location) Extension(strings Strings) (string, error) {
	return strings.Get(l.ExtensionOffset)
}

func (l Location) FullName(strings Strings) (string, error) {
	f := ""

	if l.ModuleOffset != 0 {
		s, err := strings.Get(l.ModuleOffset)
		if err != nil {
			return "", fmt.Errorf("unable to get module name\n%w", err)
		}
		f = fmt.Sprintf("%s/modules/%s/", f, s)
	}

	if l.ParentOffset != 0 {
		s, err := strings.Get(l.ParentOffset)
		if err != nil {
			return "", fmt.Errorf("unable to get parent name\n%w", err)
		}
		f = fmt.Sprintf("%s/", s)
	}

	s, err := strings.Get(l.BaseOffset)
	if err != nil {
		return "", fmt.Errorf("unable to get base name\n%w", err)
	}
	f = fmt.Sprintf("%s%s", f, s)

	if l.ExtensionOffset != 0 {
		s, err := strings.Get(l.ExtensionOffset)
		if err != nil {
			return "", fmt.Errorf("unable to get extension\n%w", err)
		}
		f = fmt.Sprintf("%s.%s", f, s)
	}

	return f, nil
}

type Locations struct {
	Offset int32
	Size   int32
	Reader io.ReadSeeker
}

func (l Locations) Get(offset Offset) (Location, error) {
	o := Location{}

	if _, err := l.Reader.Seek(int64(l.Offset)+int64(offset), io.SeekStart); err != nil {
		return Location{}, fmt.Errorf("unable to seek to beginning of location\n%w", err)
	}

	for {
		var raw byte
		if err := binary.Read(l.Reader, binary.LittleEndian, &raw); err != nil {
			return Location{}, fmt.Errorf("unable to read attribute\n%w", err)
		}

		data := raw & 0xFF
		kind := raw >> 3

		if kind == AttributeEnd {
			break
		}

		length := (data & 0x7) + 1
		value := int32(0)

		for i := 0; i < int(length); i++ {
			value <<= 8

			if err := binary.Read(l.Reader, binary.LittleEndian, &raw); err != nil {
				return Location{}, fmt.Errorf("unable to read attribute\n%w", err)
			}

			value |= int32(raw & 0xFF)
		}

		switch kind {
		case AttributeModule:
			o.ModuleOffset = value
		case AttributeParent:
			o.ParentOffset = value
		case AttributeBase:
			o.BaseOffset = value
		case AttributeExtension:
			o.ExtensionOffset = value
		case AttributeOffset:
			o.ContentOffset = value
		case AttributeCompressed:
			o.CompressedSize = value
		case AttributeUncompressed:
			o.UncompressedSize = value
		default:
			return Location{}, fmt.Errorf("unknown attribute type%d", kind)
		}
	}

	return o, nil
}
