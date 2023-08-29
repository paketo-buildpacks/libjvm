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
	"io"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm/v2/helper"
	"github.com/paketo-buildpacks/libjvm/v2/internal"
	"github.com/paketo-buildpacks/libpak/v2/log"
)

func testSecurityProvidersConfigurer(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		f, err := ioutil.TempFile("", "security-providers-configurer-path")
		Expect(err).NotTo(HaveOccurred())
		_, err = f.WriteString("test")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())
		path = f.Name()
	})

	it.After(func() {
		Expect(os.RemoveAll(path))
	})

	it("returns if $SECURITY_PROVIDERS not set", func() {
		Expect(helper.SecurityProvidersConfigurer{Logger: log.NewPaketoLogger(io.Discard)}.Execute()).To(BeNil())

		Expect(ioutil.ReadFile(path)).To(Equal([]byte("test")))
	})

	context("$SECURITY_PROVIDERS", func() {
		it.Before(func() {
			Expect(os.Setenv("SECURITY_PROVIDERS", "2|DELTA ECHO 10|FOXTROT")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("SECURITY_PROVIDERS"))
		})

		it("returns error if BPI_SECURITY_PROVIDERS is not set", func() {
			_, err := helper.SecurityProvidersConfigurer{Logger: log.NewPaketoLogger(io.Discard)}.Execute()

			Expect(err).To(MatchError("$BPI_JVM_SECURITY_PROVIDERS must be set"))
		})

		context("$BPI_JVM_SECURITY_PROVIDERS", func() {
			it.Before(func() {
				Expect(os.Setenv("BPI_JVM_SECURITY_PROVIDERS", "3|CHARLIE 1|ALPHA 2|BRAVO")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPI_JVM_SECURITY_PROVIDERS"))
			})

			it("returns error if $JAVA_SECURITY_PROPERTIES is not set", func() {
				_, err := helper.SecurityProvidersConfigurer{Logger: log.NewPaketoLogger(io.Discard)}.Execute()

				Expect(err).To(MatchError("$JAVA_SECURITY_PROPERTIES must be set"))
			})

			context("$JAVA_SECURITY_PROPERTIES", func() {
				it.Before(func() {
					Expect(os.Setenv("JAVA_SECURITY_PROPERTIES", path)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("JAVA_SECURITY_PROPERTIES"))
				})

				it("modifies the security properties file", func() {
					Expect(helper.SecurityProvidersConfigurer{Logger: log.NewPaketoLogger(io.Discard)}.Execute()).To(BeNil())

					Expect(ioutil.ReadFile(path)).To(Equal([]byte(`test
security.provider.1=ALPHA
security.provider.2=DELTA
security.provider.3=BRAVO
security.provider.4=CHARLIE
security.provider.5=ECHO
security.provider.6=FOXTROT
`)))
				})

				if internal.IsRoot() {
					return
				}

				it("warns if the file is read-only", func() {
					Expect(os.Chmod(path, 0555)).To(Succeed())

					Expect(helper.SecurityProvidersConfigurer{Logger: log.NewPaketoLogger(io.Discard)}.Execute()).To(BeNil())

					Expect(ioutil.ReadFile(path)).To(Equal([]byte("test")))
				})
			})
		})
	})
}
