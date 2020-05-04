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

package provider_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm/provider"
	"github.com/sclevine/spec"
)

func testSecurityProvidersConfigurer(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		s           provider.SecurityProvidersConfigurer
		source      string
		destination string
	)

	it.Before(func() {
		f, err := ioutil.TempFile("", "security-providers-configurer-source")
		Expect(err).NotTo(HaveOccurred())
		_, err = f.WriteString(`security.provider.1=ALPHA
security.provider.2=BRAVO
security.provider.3=CHARLIE
`)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())
		source = f.Name()

		f, err = ioutil.TempFile("", "security-providers-configurer-destination")
		Expect(err).NotTo(HaveOccurred())
		_, err = f.WriteString("test")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())
		destination = f.Name()

		s = provider.SecurityProvidersConfigurer{JRESourcePath: source, DestinationPath: destination}
	})

	it.After(func() {
		Expect(os.RemoveAll(source)).To(Succeed())
		Expect(os.RemoveAll(destination)).To(Succeed())
	})

	it("does not modify file if no additions", func() {
		Expect(s.Execute()).To(Succeed())

		Expect(ioutil.ReadFile(destination)).To(Equal([]byte("test")))
	})

	it("modifies files if additions", func() {
		s.AdditionalProviders = []string{"", "2|DELTA", "ECHO"}

		Expect(s.Execute()).To(Succeed())

		Expect(ioutil.ReadFile(destination)).To(Equal([]byte(`test
security.provider.1=ALPHA
security.provider.2=DELTA
security.provider.3=BRAVO
security.provider.4=CHARLIE
security.provider.5=ECHO
`)))
	})
}
