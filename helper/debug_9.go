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
	"strconv"
	"strings"

	"github.com/paketo-buildpacks/libpak/bard"
)

type Debug9 struct {
	Logger bard.Logger
}

func (d Debug9) Execute() (map[string]string, error) {
	if _, ok := os.LookupEnv("BPL_DEBUG_ENABLED"); !ok {
		return nil, nil
	}

	var err error

	port := "*:8000" // Java 9+ address format
	if s, ok := os.LookupEnv("BPL_DEBUG_PORT"); ok {
		port = "*:" + s
	}

	suspend := false
	if s, ok := os.LookupEnv("BPL_DEBUG_SUSPEND"); ok {
		suspend, err = strconv.ParseBool(s)
		if err != nil {
			return nil, fmt.Errorf("unable to parse $BPL_DEBUG_SUSPEND\n%w", err)
		}
	}

	s := fmt.Sprintf("Debugging enabled on port %s", port)
	if suspend {
		s = fmt.Sprintf("%s, suspended on start", s)
	}
	d.Logger.Info(s)

	var values []string
	if s, ok := os.LookupEnv("JAVA_TOOL_OPTIONS"); ok {
		values = append(values, s)
	}

	if suspend {
		s = "y"
	} else {
		s = "n"
	}

	values = append(values,
		fmt.Sprintf("-agentlib:jdwp=transport=dt_socket,server=y,address=%s,suspend=%s", port, s))

	return map[string]string{"JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil
}
