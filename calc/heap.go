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

package calc

import (
	"fmt"
	"regexp"
	"strings"
)

var HeapRE = regexp.MustCompile(fmt.Sprintf("^-Xmx(%s)$", SizePattern))

type Heap Size

func (h Heap) String() string {
	return fmt.Sprintf("-Xmx%s", Size(h))
}

func MatchHeap(s string) bool {
	return HeapRE.MatchString(strings.TrimSpace(s))
}

func ParseHeap(s string) (*Heap, error) {
	g := HeapRE.FindStringSubmatch(s)
	if g == nil {
		return nil, fmt.Errorf("%s does not match heap pattern %s", s, HeapRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse heap size\n%w", err)
	}

	h := Heap(z)
	return &h, nil
}
