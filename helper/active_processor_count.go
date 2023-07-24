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
	"runtime"
	"strings"

	"github.com/mattn/go-shellwords"
	"github.com/paketo-buildpacks/libpak/bard"
)

type ActiveProcessorCount struct {
	Logger bard.Logger
}

func (a ActiveProcessorCount) Execute() (map[string]string, error) {
	var values []string
	s, ok := os.LookupEnv("JAVA_TOOL_OPTIONS")
	if ok {
		values = append(values, s)
	}

	if p, err := shellwords.Parse(s); err != nil {
		return nil, fmt.Errorf("unable to parse $JAVA_TOOL_OPTIONS\n%w", err)
	} else {
		for _, s := range p {
			if strings.HasPrefix(s, "-XX:ActiveProcessorCount=") {
				return nil, nil
			}
		}
	}

	a.Logger.Debugf("Setting Active Processor Count to %d", runtime.NumCPU())

	values = append(values, fmt.Sprintf("-XX:ActiveProcessorCount=%d", runtime.NumCPU()))

	return map[string]string{"JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil
}
