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
	"strings"

	"github.com/paketo-buildpacks/libpak/sherpa"

	"github.com/paketo-buildpacks/libpak/bard"
)

type JFR struct {
	Logger bard.Logger
}

func (j JFR) Execute() (map[string]string, error) {
	if val, ok := os.LookupEnv("BPL_JAVA_FLIGHT_RECORDER_ENABLED"); !ok || val != "true" {
		return nil, nil
	}
	// list of valid JFR config arguments, to test user-provided args against keys
	validArgs := map[string]string{"delay": "", "dumponexit": "", "filename": "", "name": "", "duration": "",
		"maxage": "", "maxsize": "", "path-to-gc-roots": "", "settings": ""}

	argList := sherpa.GetEnvWithDefault("BPL_JFR_ARGS", "")

	j.Logger.Infof("Enabling Java Flight Recorder")

	// minimum flag to enable JFR, with default config args
	jfrConfig := "-XX:StartFlightRecording="

	if argList != "" {
		list := strings.Split(argList, ",")
		for _, a := range list {
			if a == "" {
				return nil, fmt.Errorf("unable to parse Flight Recorder arguments: %s", argList)
			}
			kv := strings.Split(a, "=")
			if _, ok := validArgs[kv[0]]; !ok || (kv[0] == "" || kv[1] == "") {
				return nil, fmt.Errorf("invalid Flight Recorder argument: %s=%s", kv[0], kv[1])
			}
		}
		jfrConfig += argList
	}
	opts := sherpa.AppendToEnvVar("JAVA_TOOL_OPTIONS", " ", jfrConfig)

	return map[string]string{"JAVA_TOOL_OPTIONS": opts}, nil
}
