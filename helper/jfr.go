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
	"path/filepath"

	"github.com/paketo-buildpacks/libpak/v2/log"
	"github.com/paketo-buildpacks/libpak/v2/sherpa"
)

type JFR struct {
	Logger log.Logger
}

func (j JFR) Execute() (map[string]string, error) {
	if val := sherpa.ResolveBool("BPL_JFR_ENABLED"); !val {
		return nil, nil
	}

	var argList string
	if argList = sherpa.GetEnvWithDefault("BPL_JFR_ARGS", ""); argList == "" {
		argList = fmt.Sprintf("dumponexit=true,filename=%s", filepath.Join(os.TempDir(), "recording.jfr"))
	}
	j.Logger.Body("Enabling Java Flight Recorder with args: %s", argList)

	// minimum flag to enable JFR, with default config args
	jfrConfig := fmt.Sprintf("-XX:StartFlightRecording=%s", argList)

	opts := sherpa.AppendToEnvVar("JAVA_TOOL_OPTIONS", " ", jfrConfig)

	return map[string]string{"JAVA_TOOL_OPTIONS": opts}, nil
}
