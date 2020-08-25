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

var MetaspaceRE = regexp.MustCompile(fmt.Sprintf("^-XX:MaxMetaspaceSize=(%s)$", SizePattern))

type Metaspace Size

func (m Metaspace) String() string {
	return fmt.Sprintf("-XX:MaxMetaspaceSize=%s", Size(m))
}

func MatchMetaspace(s string) bool {
	return MetaspaceRE.MatchString(strings.TrimSpace(s))
}

func ParseMetaspace(s string) (*Metaspace, error) {
	g := MetaspaceRE.FindStringSubmatch(s)
	if g == nil {
		return nil, fmt.Errorf("%s does not match metaspace pattern %s", s, MetaspaceRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse metaspace size\n%w", err)
	}

	m := Metaspace(z)
	return &m, nil
}
