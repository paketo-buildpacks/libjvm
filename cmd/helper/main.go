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

	"github.com/miekg/dns"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"

	"github.com/paketo-buildpacks/libjvm"
	"github.com/paketo-buildpacks/libjvm/helper"
)

func main() {
	sherpa.Execute(func() error {
		var (
			err error

			l = bard.NewLogger(os.Stdout)

			cl = libjvm.NewCertificateLoader()

			a  = helper.ActiveProcessorCount{Logger: l}
			c  = helper.SecurityProvidersConfigurer{Logger: l}
			d  = helper.LinkLocalDNS{Logger: l}
			j  = helper.JavaOpts{Logger: l}
			jh = helper.JVMHeapDump{Logger: l}
			m  = helper.MemoryCalculator{
				Logger:            l,
				MemoryLimitPathV1: helper.DefaultMemoryLimitPathV1,
				MemoryLimitPathV2: helper.DefaultMemoryLimitPathV2,
				MemoryInfoPath:    helper.DefaultMemoryInfoPath,
			}
			o  = helper.OpenSSLCertificateLoader{CertificateLoader: cl, Logger: l}
			s8 = helper.SecurityProvidersClasspath8{Logger: l}
			s9 = helper.SecurityProvidersClasspath9{Logger: l}
			d8 = helper.Debug8{Logger: l}
			d9 = helper.Debug9{Logger: l}
			jm = helper.JMX{Logger: l}
			n  = helper.NMT{Logger: l}
			jf = helper.JFR{Logger: l}
		)

		file := "/etc/resolv.conf"
		d.Config, err = dns.ClientConfigFromFile(file)
		if err != nil {
			return fmt.Errorf("unable to read DNS client configuration from %s\n%w", file, err)
		}

		return sherpa.Helpers(map[string]sherpa.ExecD{
			"active-processor-count":         a,
			"java-opts":                      j,
			"jvm-heap":                       jh,
			"link-local-dns":                 d,
			"memory-calculator":              m,
			"openssl-certificate-loader":     o,
			"security-providers-classpath-8": s8,
			"security-providers-classpath-9": s9,
			"security-providers-configurer":  c,
			"debug-8":                        d8,
			"debug-9":                        d9,
			"jmx":                            jm,
			"nmt":                            n,
			"jfr":                            jf,
		})
	})
}
