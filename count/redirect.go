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

const RedirectSize = 4

type Redirect int32
type Redirects struct {
	Entries []Redirect
	Offset  int32
}

func NewRedirects(reader io.ReadSeeker, offset int32, tableLength int32) (Redirects, error) {
	r := Redirects{
		Entries: make([]Redirect, tableLength),
		Offset:  offset,
	}

	if _, err := reader.Seek(int64(offset), io.SeekStart); err != nil {
		return Redirects{}, fmt.Errorf("unable to seek to beginning of redirects\n%w", err)
	}

	for i := 0; i < len(r.Entries); i++ {
		if err := binary.Read(reader, binary.LittleEndian, &r.Entries[i]); err != nil {
			return Redirects{}, fmt.Errorf("unable to read redirect\n%w", err)
		}
	}

	return r, nil
}

func (r Redirects) Size() int {
	return len(r.Entries) * RedirectSize
}
