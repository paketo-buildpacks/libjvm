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
	"net"
	"os"

	"github.com/miekg/dns"
	"github.com/paketo-buildpacks/libpak/v2/log"
	"golang.org/x/sys/unix"
)

type LinkLocalDNS struct {
	Config *dns.ClientConfig
	Logger log.Logger
}

func (l LinkLocalDNS) Execute() (map[string]string, error) {
	if !net.ParseIP(l.Config.Servers[0]).IsLinkLocalUnicast() {
		return nil, nil
	}

	file, ok := os.LookupEnv("JAVA_SECURITY_PROPERTIES")
	if !ok {
		return nil, fmt.Errorf("$JAVA_SECURITY_PROPERTIES must be set")
	}
	if unix.Access(file, unix.W_OK) != nil {
		l.Logger.Bodyf("WARNING: Unable to disable JVM DNS caching disabled in favor of link-local DNS caching because %s is read-only", file)
		return nil, nil
	}

	l.Logger.Body("JVM DNS caching disabled in favor of link-local DNS caching")

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s\n%w", file, err)
	}
	defer f.Close()

	_, err = f.WriteString(`
networkaddress.cache.ttl=0
networkaddress.cache.negative.ttl=0
`)
	if err != nil {
		return nil, fmt.Errorf("unable to write DNS configuration to %s\n%w", file, err)
	}

	return nil, nil
}
