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

	"github.com/paketo-buildpacks/libpak/sherpa"

	"github.com/paketo-buildpacks/libpak/bard"
)

type NMT struct {
	Logger bard.Logger
}

func (n NMT) Execute() (map[string]string, error) {

	if !ResolveBoolWithDefault("BPL_JAVA_NMT_ENABLED", true) {
		n.Logger.Info("Disabling Java Native Memory Tracking")
		return nil, nil
	}

	level := sherpa.GetEnvWithDefault("BPL_JAVA_NMT_LEVEL", "summary")

	n.Logger.Info("Enabling Java Native Memory Tracking")

	opts := sherpa.AppendToEnvVar("JAVA_TOOL_OPTIONS", " ", fmt.Sprintf("-XX:+UnlockDiagnosticVMOptions -XX:NativeMemoryTracking=%s -XX:+PrintNMTStatistics", level))

	// NMT_LEVEL_1 Required for Java Native Memory Tracking to work due to bug which is not fixed until Java v18 (https://bugs.openjdk.java.net/browse/JDK-8256844)
	// '1' = PID of Java process in the container. Value for NMT level should match that passed to '-XX:NativeMemoryTracking' in the NMT helper.
	return map[string]string{"NMT_LEVEL_1": level, "JAVA_TOOL_OPTIONS": opts}, nil

}

// ResolveBoolWithDefault TODO - replace calling this with libpak's sherpa.ResolveBoolWithDefault once it is implemented
func ResolveBoolWithDefault(name string, defaultVal bool) bool {
	s, ok := os.LookupEnv(name)
	if !ok {
		// not set, use default (in NMT's case, true/enable)
		return defaultVal
	}

	t, err := strconv.ParseBool(s)
	if err != nil {
		// set but contains junk, default to false/disable
		return false
	}

	return t
}
