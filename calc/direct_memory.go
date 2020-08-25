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

var (
	DefaultDirectMemory = DirectMemory{Value: 10 * Mibi, Provenance: Default}
	DirectMemoryRE      = regexp.MustCompile(fmt.Sprintf("^-XX:MaxDirectMemorySize=(%s)$", SizePattern))
)

type DirectMemory Size

func (d DirectMemory) String() string {
	return fmt.Sprintf("-XX:MaxDirectMemorySize=%s", Size(d))
}

func MatchDirectMemory(s string) bool {
	return DirectMemoryRE.MatchString(strings.TrimSpace(s))
}

func ParseDirectMemory(s string) (DirectMemory, error) {
	g := DirectMemoryRE.FindStringSubmatch(s)
	if g == nil {
		return DirectMemory{}, fmt.Errorf("%s does not match direct memory pattern %s", s, DirectMemoryRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return DirectMemory{}, fmt.Errorf("unable to parse direct memory size\n%w", err)
	}

	return DirectMemory(z), nil
}
