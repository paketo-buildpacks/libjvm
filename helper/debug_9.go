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
	Logger         bard.Logger
	ListenerHolder Debug9ListenerHolder
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
	// if not set, it means no test set that up; it's the real runtime Listener that will be used
	if d.ListenerHolder == nil {
		d.ListenerHolder = RealListenerHolder{}
	}

	listener, err := d.ListenerHolder.Listen("tcp", fmt.Sprintf("[::]:%s", port))
	defer listener.Close()
	if err != nil {
		d.Logger.Infof("Port %s is not available or cannot be opened on [::]; configuring debug agent with 0.0.0.0\n", port)
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

type Debug9ListenerHolder interface {
	Listen(network, address string) (net.Listener, error)
}

type RealListenerHolder struct {
}

func (RealListenerHolder) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}
