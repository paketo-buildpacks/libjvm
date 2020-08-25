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
	DefaultReservedCodeCache = ReservedCodeCache{Value: 240 * Mibi, Provenance: Default}
	ReservedCodeCacheRE      = regexp.MustCompile(fmt.Sprintf("^-XX:ReservedCodeCacheSize=(%s)$", SizePattern))
)

type ReservedCodeCache Size

func (r ReservedCodeCache) String() string {
	return fmt.Sprintf("-XX:ReservedCodeCacheSize=%s", Size(r))
}

func MatchReservedCodeCache(s string) bool {
	return ReservedCodeCacheRE.MatchString(strings.TrimSpace(s))
}

func ParseReservedCodeCache(s string) (ReservedCodeCache, error) {
	g := ReservedCodeCacheRE.FindStringSubmatch(s)
	if g == nil {
		return ReservedCodeCache{}, fmt.Errorf("%s does not match reserved code cache pattern %s", s, ReservedCodeCacheRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return ReservedCodeCache{}, fmt.Errorf("unable to parse reserved code cache size\n%w", err)
	}

	return ReservedCodeCache(z), nil
}
