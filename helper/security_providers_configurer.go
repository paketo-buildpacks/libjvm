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

package helper

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/paketo-buildpacks/libpak/bard"
	"golang.org/x/sys/unix"
)

type SecurityProvidersConfigurer struct {
	Logger bard.Logger
}

func (s SecurityProvidersConfigurer) Execute() (map[string]string, error) {
	a, ok := os.LookupEnv("SECURITY_PROVIDERS")
	if !ok {
		return nil, nil
	}

	e, ok := os.LookupEnv("BPI_JVM_SECURITY_PROVIDERS")
	if !ok {
		return nil, fmt.Errorf("$BPI_JVM_SECURITY_PROVIDERS must be set")
	}

	file, ok := os.LookupEnv("JAVA_SECURITY_PROPERTIES")
	if !ok {
		return nil, fmt.Errorf("$JAVA_SECURITY_PROPERTIES must be set")
	}
	if unix.Access(file, unix.W_OK) != nil {
		s.Logger.Debugf("WARNING: Unable to add additional security providers because %s is read-only", file)
		return nil, nil
	}

	s.Logger.Debug("Adding Security Providers to JVM")

	providers := make([]string, 0)
	r := regexp.MustCompile(`(?:([\d]+)\|)?([\w.]+)`)
	for _, s := range append(strings.Split(e, " "), strings.Split(a, " ")...) {
		if matches := r.FindStringSubmatch(s); matches != nil {
			if matches[1] == "" {
				providers = append(providers, matches[2])
				continue
			}

			i, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, fmt.Errorf("index %s is not a number\n%w", matches[1], err)
			}
			i--

			for {
				if len(providers) > i {
					break
				}
				providers = append(providers, "")
			}

			if providers[i] != "" {
				providers = append(providers, "")
				copy(providers[i+1:], providers[i:])
			}

			providers[i] = matches[2]
		}
	}

	j := 0
	for {
		if j >= len(providers) {
			break
		}

		if providers[j] != "" {
			j++
			continue
		}

		copy(providers[j:], providers[j+1:])
		providers = providers[:len(providers)-1]
	}

	out, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %s\n%w", file, err)
	}
	defer out.Close()

	if _, err := out.WriteString("\n"); err != nil {
		return nil, fmt.Errorf("unable to write to %s\n%w", file, err)
	}

	for i, p := range providers {
		if _, err := out.WriteString(fmt.Sprintf("security.provider.%d=%s\n", i+1, p)); err != nil {
			return nil, fmt.Errorf("unable to write to %s\n%w", file, err)
		}
	}

	return nil, nil
}
