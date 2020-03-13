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

	ddns "github.com/miekg/dns"
	"github.com/paketo-buildpacks/libjvm/dns"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

func main() {
	sherpa.Execute(func() error {
		var (
			err error
			l   dns.LinkLocalDNS
			ok  bool
		)

		file := "/etc/resolv.conf"
		l.Config, err = ddns.ClientConfigFromFile(file)
		if err != nil {
			return fmt.Errorf("unable to read DNS client configuration from %s\n%w", file, err)
		}

		if l.SecurityPath, ok = os.LookupEnv("JAVA_SECURITY_PROPERTIES"); !ok {
			return fmt.Errorf("$JAVA_SECURITY_PROPERTIES must be set")
		}

		return l.Execute()
	})
}
