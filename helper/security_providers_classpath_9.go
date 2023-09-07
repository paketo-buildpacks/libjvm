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
	"os"
	"strings"

	"github.com/paketo-buildpacks/libpak/v2/log"
)

type SecurityProvidersClasspath9 struct {
	Logger log.Logger
}

func (s SecurityProvidersClasspath9) Execute() (map[string]string, error) {
	p, ok := os.LookupEnv("SECURITY_PROVIDERS_CLASSPATH")
	if !ok {
		return nil, nil
	}

	s.Logger.Body("Adding $SECURITY_PROVIDERS_CLASSPATH to $CLASSPATH")

	var values []string
	if s, ok := os.LookupEnv("CLASSPATH"); ok {
		values = append(values, s)
	}

	values = append(values, p)

	return map[string]string{"CLASSPATH": strings.Join(values, ":")}, nil
}
