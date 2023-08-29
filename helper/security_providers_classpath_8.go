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
	"strings"

	"github.com/paketo-buildpacks/libpak/v2/log"
)

type SecurityProvidersClasspath8 struct {
	Logger log.Logger
}

func (s SecurityProvidersClasspath8) Execute() (map[string]string, error) {
	p, ok := os.LookupEnv("SECURITY_PROVIDERS_CLASSPATH")
	if !ok {
		return nil, nil
	}

	e, ok := os.LookupEnv("BPI_JVM_EXT_DIR")
	if !ok {
		return nil, fmt.Errorf("$BPI_JVM_EXT_DIR must be set")
	}

	s.Logger.Body("Adding $SECURITY_PROVIDERS_CLASSPATH to java.ext.dirs")

	var values []string
	if s, ok := os.LookupEnv("JAVA_TOOL_OPTIONS"); ok {
		values = append(values, s)
	}

	dirs := []string{e}
	for _, f := range strings.Split(p, ":") {
		dirs = append(dirs, filepath.Dir(f))
	}
	values = append(values, fmt.Sprintf("-Djava.ext.dirs=%s", strings.Join(dirs, ":")))

	return map[string]string{"JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil
}
