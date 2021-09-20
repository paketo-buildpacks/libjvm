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

package helper_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnit(t *testing.T) {
	suite := spec.New("libjvm/helper", spec.Report(report.Terminal{}))
	suite("ActiveProcessorCount", testActiveProcessorCount)
	suite("JavaOpts", testJavaOpts)
	suite("JVMHeapDump", testJVMHeapDump)
	suite("LinkLocalDNS", testLinkLocalDNS)
	suite("MemoryCalculator", testMemoryCalculator)
	suite("OpenSSLCertificateLoader", testOpenSSLCertificateLoader)
	suite("SecurityProvidersClasspath8", testSecurityProvidersClasspath8)
	suite("SecurityProvidersClasspath9", testSecurityProvidersClasspath9)
	suite("SecurityProvidersConfigurer", testSecurityProvidersConfigurer)
	suite("Debug8", testDebug8)
	suite("Debug9", testDebug9)
	suite("JMX", testJMX)
	suite("NMT", testNMT)
	suite.Run(t)
}
