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
	DefaultStack = Stack{Value: 1 * Mebi, Provenance: Default}
	StackRE      = regexp.MustCompile(fmt.Sprintf("^-Xss(%s)$", SizePattern))
)

type Stack Size

func (s Stack) String() string {
	return fmt.Sprintf("-Xss%s", Size(s))
}

func MatchStack(s string) bool {
	return StackRE.MatchString(strings.TrimSpace(s))
}

func ParseStack(s string) (Stack, error) {
	g := StackRE.FindStringSubmatch(s)
	if g == nil {
		return Stack{}, fmt.Errorf("%s does not match stack pattern %s", s, StackRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return Stack{}, fmt.Errorf("unable to parse stack size\n%w", err)
	}

	return Stack(z), nil
}
