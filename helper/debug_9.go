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
	"net"
	"strings"

	"github.com/paketo-buildpacks/libpak/bard"
)

type Debug9 struct {
	Logger                 bard.Logger
	Debug9ListenerAndError ListenerAndError
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
	if &d.Debug9ListenerAndError.Listener == nil {
		listener, err := net.Listen("tcp", fmt.Sprintf("[::]:%s", port))
		d.Debug9ListenerAndError = ListenerAndError{listener, err}
	}

	var host = "*"
	if !isIPv6PortBindingOK(&d.Debug9ListenerAndError) {
		d.Logger.Debugf("Port %s is not available or cannot be opened on [::]; configuring debug agent with 0.0.0.0\n", port)
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

func isIPv6PortBindingOK(listenerAndError *ListenerAndError) bool {
	defer listenerAndError.Listener.Close()
	return listenerAndError.Err == nil
}

type ListenerAndError struct {
	Listener net.Listener
	Err      error
}
