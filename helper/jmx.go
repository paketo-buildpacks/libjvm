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

	"github.com/paketo-buildpacks/libpak/v2/bard"
	"github.com/paketo-buildpacks/libpak/v2/sherpa"
)

type JMX struct {
	Logger bard.Logger
}

func (j JMX) Execute() (map[string]string, error) {
	if val := sherpa.ResolveBool("BPL_JMX_ENABLED"); !val {
		return nil, nil
	}

	port := sherpa.GetEnvWithDefault("BPL_JMX_PORT", "5000")

	j.Logger.Debug("JMX enabled on port %s", port)

	opts := sherpa.AppendToEnvVar("JAVA_TOOL_OPTIONS", " ", "-Djava.rmi.server.hostname=127.0.0.1",
		"-Dcom.sun.management.jmxremote.authenticate=false",
		"-Dcom.sun.management.jmxremote.ssl=false",
		fmt.Sprintf("-Dcom.sun.management.jmxremote.port=%s", port),
		fmt.Sprintf("-Dcom.sun.management.jmxremote.rmi.port=%s", port))

	return map[string]string{"JAVA_TOOL_OPTIONS": opts}, nil
}
