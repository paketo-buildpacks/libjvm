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

package dns

import (
	"fmt"
	"net"
	"os"

	"github.com/miekg/dns"
)

type LinkLocalDNS struct {
	Config       *dns.ClientConfig
	SecurityPath string
}

func (l *LinkLocalDNS) Execute() error {
	if !net.ParseIP(l.Config.Servers[0]).IsLinkLocalUnicast() {
		return nil
	}

	fmt.Println("JVM DNS caching disabled in favor of link-local DNS caching")

	f, err := os.OpenFile(l.SecurityPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("unable to open %s: %w", l.SecurityPath, err)
	}
	defer f.Close()

	_, err = f.WriteString(`
networkaddress.cache.ttl=0
networkaddress.cache.negative.ttl=0
`)
	if err != nil {
		return fmt.Errorf("unable to write DNS configuration to %s: %w", l.SecurityPath, err)
	}

	return nil
}
