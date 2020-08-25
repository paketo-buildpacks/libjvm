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
	"fmt"
	"os"
)

type Image struct {
	MajorVersion  int32
	MinorVersion  int32
	Flags         int32
	ResourceCount int32

	Redirects Redirects
	Offsets   Offsets
	Locations Locations
	Strings   Strings
}

func NewImage(path string) (Image, error) {
	var i Image

	in, err := os.Open(path)
	if err != nil {
		return Image{}, fmt.Errorf("unable to open %s\n%w", path, err)
	}

	h, err := NewHeader(in)
	if err != nil {
		return Image{}, fmt.Errorf("unable to read header\n%w", err)
	}

	i.MajorVersion = h.MajorVersion
	i.MinorVersion = h.MinorVersion
	i.Flags = h.Flags
	i.ResourceCount = h.ResourceCount

	i.Redirects, err = NewRedirects(in, int32(h.Size()), h.TableLength)
	if err != nil {
		return Image{}, fmt.Errorf("unable to create redirects\n%w", err)
	}

	i.Offsets, err = NewOffsets(in, i.Redirects.Offset+int32(i.Redirects.Size()), h.TableLength)
	if err != nil {
		return Image{}, fmt.Errorf("unable to create offsets\n%w", err)
	}

	i.Locations = Locations{Offset: i.Offsets.Offset + int32(i.Offsets.Size()), Size: h.LocationsSize, Reader: in}

	i.Strings = Strings{Offset: i.Locations.Offset + i.Locations.Size, Reader: in}

	return i, nil
}
