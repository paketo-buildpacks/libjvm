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

	"github.com/paketo-buildpacks/libpak/sherpa"
	"github.com/spf13/pflag"

	"github.com/paketo-buildpacks/libjvm"
)

func main() {
	sherpa.Execute(func() error {
		c := libjvm.CertificateLoader{Logger: os.Stdout}

		flagSet := pflag.NewFlagSet("OpenSSL Certificate Loader", pflag.ExitOnError)
		flagSet.StringVar(&c.CACertificatesPath, "ca-certificates", "", "path to OpenSSL CA Certificates file")
		flagSet.StringVar(&c.KeyStorePassword, "keystore-password", "", "password for the Java cacerts keystore")
		flagSet.StringVar(&c.KeyStorePath, "keystore-path", "", "path to Java cacerts keystore")

		if err := flagSet.Parse(os.Args[1:]); err != nil {
			return fmt.Errorf("unable to parse flags\n%w", err)
		}

		return c.Load()
	})
}
