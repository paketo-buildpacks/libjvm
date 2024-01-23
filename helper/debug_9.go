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
	"github.com/paketo-buildpacks/libpak/sherpa"
	"io"
	"os"
	"strings"

	"github.com/paketo-buildpacks/libpak/bard"
)

const DefaultIPv6CheckPath = "/sys/module/ipv6/parameters/disable"

type Debug9 struct {
	Logger              bard.Logger
	CustomIPv6CheckPath string
}

func (d Debug9) Execute() (map[string]string, error) {

	if val := sherpa.ResolveBool("BPL_DEBUG_ENABLED"); !val {
		return nil, nil
	}

	opts := sherpa.GetEnvWithDefault("JAVA_TOOL_OPTIONS", "")
	debugAlreadyExists := strings.Contains(opts, "-agentlib:jdwp=")

	if debugAlreadyExists {
		d.Logger.Info("Java agent 'jdwp' already configured")
		return nil, nil
	}

	port := sherpa.GetEnvWithDefault("BPL_DEBUG_PORT", "8000")
	var host = "*"
	var iPv6CheckPath string
	if d.CustomIPv6CheckPath != "" {
		iPv6CheckPath = d.CustomIPv6CheckPath
	} else {
		iPv6CheckPath = DefaultIPv6CheckPath
	}
	if !IPv6Enabled(iPv6CheckPath) {
		d.Logger.Infof("IPv6 does not seem to be enabled in the container, configuring debug agent with 0.0.0.0\n")
		host = "0.0.0.0"
	}

	address := host + ":" + port

	suspend := sherpa.ResolveBool("BPL_DEBUG_SUSPEND")

	s := fmt.Sprintf("Debugging enabled on address %s", address)
	if suspend {
		s = fmt.Sprintf("%s, suspended on start", s)
	}
	d.Logger.Info(s)

	if suspend {
		s = "y"
	} else {
		s = "n"
	}

	opts = sherpa.AppendToEnvVar("JAVA_TOOL_OPTIONS", " ", fmt.Sprintf("-agentlib:jdwp=transport=dt_socket,server=y,address=%s,suspend=%s", address, s))

	return map[string]string{"JAVA_TOOL_OPTIONS": opts}, nil
}

func IPv6Enabled(iPv6CheckPath string) bool {
	in, err := os.Open(iPv6CheckPath)

	if err != nil {
		return false
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)

	b, err := io.ReadAll(in)
	value := string(b[0:1])

	if err != nil || value == "1" {
		return false
	} else {
		return true
	}
}
