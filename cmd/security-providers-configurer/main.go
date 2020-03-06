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

package main

import (
	"fmt"
	"os"

	"github.com/paketoio/libjvm/provider"
	"github.com/paketoio/libpak/sherpa"
	"github.com/spf13/pflag"
)

func main() {
	sherpa.Execute(func() error {
		var (
			ok bool
			s  provider.SecurityProvidersConfigurer
		)

		flagSet := pflag.NewFlagSet("Security Providers Configurer", pflag.ExitOnError)
		flagSet.StringSliceVar(&s.AdditionalProviders, "additional-providers", []string{}, "additional security providers")
		flagSet.StringVar(&s.SourcePath, "source", "", "path to security.properties file containing existing providers")

		if err := flagSet.Parse(os.Args[1:]); err != nil {
			return fmt.Errorf("unable to parse flags: %w", err)
		}

		if s.DestinationPath, ok = os.LookupEnv("JAVA_SECURITY_PROPERTIES"); !ok {
			return fmt.Errorf("$JAVA_SECURITY_PROPERTIES must be set")
		}

		return s.Execute()
	})
}
