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

	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type Debug8 struct {
	Logger bard.Logger
}

func (d Debug8) Execute() (map[string]string, error) {

	if val := sherpa.ResolveBool("BPL_DEBUG_ENABLED"); !val {
		return nil, nil
	}

	port := sherpa.GetEnvWithDefault("BPL_DEBUG_PORT", "8000")

	suspend := sherpa.ResolveBool("BPL_DEBUG_SUSPEND")

	s := fmt.Sprintf("Debugging enabled on port %s", port)
	if suspend {
		s = fmt.Sprintf("%s, suspended on start", s)
	}
	d.Logger.Info(s)

	if suspend {
		s = "y"
	} else {
		s = "n"
	}

	opts := sherpa.AppendToEnvVar("JAVA_TOOL_OPTIONS", " ", fmt.Sprintf("-agentlib:jdwp=transport=dt_socket,server=y,address=%s,suspend=%s", port, s))

	return map[string]string{"JAVA_TOOL_OPTIONS": opts}, nil
}
