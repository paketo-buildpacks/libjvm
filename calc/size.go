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
	"strconv"
	"strings"
)

type Provenance uint8

const (
	Unknown Provenance = iota
	Default
	UserConfigured
	Calculated

	Kibi = int64(1_024)
	Mebi = 1_024 * Kibi
	Gibi = 1_024 * Mebi
	Tebi = 1_024 * Gibi

	SizePattern = "([\\d]+)([kmgtKMGT]?)"
)

var SizeRE = regexp.MustCompile(fmt.Sprintf("^%s$", SizePattern))

type Size struct {
	Value      int64
	Provenance Provenance
}

// ParseSize parses a memory size in bytes from the given string. Size my include a K, M, G, or T suffix which indicates
// kibibytes, mebibytes, gibibytes or tebibytes respectively.
func ParseSize(s string) (Size, error) {
	t := strings.TrimSpace(s)

	if !SizeRE.MatchString(t) {
		return Size{}, fmt.Errorf("memory size %q does not match pattern %q", t, SizeRE.String())
	}

	groups := SizeRE.FindStringSubmatch(t)
	size, err := strconv.ParseInt(groups[1], 10, 64)
	if err != nil {
		return Size{}, fmt.Errorf("memory size %q is not an integer", groups[1])
	}

	switch strings.ToLower(groups[2]) {
	case "k":
		size *= Kibi
	case "m":
		size *= Mebi
	case "g":
		size *= Gibi
	case "t":
		size *= Tebi
	}

	return Size{Value: size}, nil
}

func (s Size) String() string {
	b := s.Value / Kibi

	if b == 0 {
		return "0"
	}

	if b%Gibi == 0 {
		return fmt.Sprintf("%dT", b/Gibi)
	}

	if b%Mebi == 0 {
		return fmt.Sprintf("%dG", b/Mebi)
	}

	if b%Kibi == 0 {
		return fmt.Sprintf("%dM", b/Kibi)
	}

	return fmt.Sprintf("%dK", b)
}

// ParseUnit parses a unit string and returns the number of bytes in the given unit. It assumes all units are binary
// units.
func ParseUnit(u string) (int64, error) {
	switch strings.TrimSpace(u) {
	case "kB", "KB", "KiB":
		return Kibi, nil
	case "MB", "MiB":
		return Mebi, nil
	case "GB", "GiB":
		return Gibi, nil
	case "TB", "TiB":
		return Tebi, nil
	case "B", "":
		return int64(1), nil
	default:
		return 0, fmt.Errorf("unrecognized unit %q", u)
	}
}
