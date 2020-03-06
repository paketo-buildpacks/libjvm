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

package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/magiconair/properties"
)

type SecurityProvidersConfigurer struct {
	SourcePath          string
	DestinationPath     string
	AdditionalProviders []string
}

func (s *SecurityProvidersConfigurer) Execute() error {
	if len(s.AdditionalProviders) == 0 {
		return nil
	}

	fmt.Println("Adding Security Providers to JVM")

	p, err := properties.LoadFile(s.SourcePath, properties.UTF8)
	if err != nil {
		return fmt.Errorf("unable to read properties file %s: %w", s.SourcePath, err)
	}
	p = p.FilterStripPrefix("security.provider.")

	keys := p.Keys()
	sort.Slice(keys, func(i, j int) bool {
		a, err := strconv.Atoi(keys[i])
		if err != nil {
			return false
		}

		b, err := strconv.Atoi(keys[j])
		if err != nil {
			return false
		}

		return a < b
	})

	var providers []string
	for _, k := range keys {
		providers = append(providers, p.MustGet(k))
	}

	r := regexp.MustCompile(`(?:([\d]+)\|)?([\w.]+)`)
	for _, a := range s.AdditionalProviders {
		matches := r.FindStringSubmatch(a)

		if matches[1] == "" {
			providers = append(providers, matches[2])
			continue
		}

		i, err := strconv.Atoi(matches[1])
		if err != nil {
			return fmt.Errorf("index %s is not a number: %w", matches[1], err)
		}

		providers = append(providers, "")
		copy(providers[i:], providers[i-1:])
		providers[i-1] = matches[2]
	}

	file := filepath.Dir(s.DestinationPath)
	if err := os.MkdirAll(file, 0755); err != nil {
		return fmt.Errorf("unable to create directory %s: %w", file, err)
	}

	out, err := os.OpenFile(s.DestinationPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", s.DestinationPath, err)
	}
	defer out.Close()

	if _, err := out.WriteString("\n"); err != nil {
		return fmt.Errorf("unable to write to %s: %w", s.DestinationPath, err)
	}

	for i, p := range providers {
		if _, err := out.WriteString(fmt.Sprintf("security.provider.%d=%s\n", i+1, p)); err != nil {
			return fmt.Errorf("unable to write to %s: %w", s.DestinationPath, err)
		}
	}

	return nil
}
