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
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm/v2/helper"
	"github.com/paketo-buildpacks/libpak/v2/bard"
)

func testSecurityProvidersClasspath9(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("return nil if SECURITY_PROVIDERS_CLASSPATH is not set", func() {
		Expect(helper.SecurityProvidersClasspath9{Logger: bard.NewLogger(io.Discard)}.Execute()).To(BeNil())
	})

	context("$SECURITY_PROVIDERS_CLASSPATH", func() {
		it.Before(func() {
			Expect(os.Setenv("SECURITY_PROVIDERS_CLASSPATH", "test-security-providers-classpath")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("SECURITY_PROVIDERS_CLASSPATH")).To(Succeed())
		})

		it("return $CLASSPATH with $SECURITY_PROVIDERS_CLASSPATH only", func() {
			Expect(helper.SecurityProvidersClasspath9{Logger: bard.NewLogger(io.Discard)}.Execute()).To(Equal(map[string]string{
				"CLASSPATH": "test-security-providers-classpath",
			}))
		})

		context("$CLASSPATH", func() {

			it.Before(func() {
				Expect(os.Setenv("CLASSPATH", "test-classpath")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("CLASSPATH")).To(Succeed())
			})

			it("return $CLASSPATH with $SECURITY_PROVIDERS_CLASSPATH included", func() {
				Expect(helper.SecurityProvidersClasspath9{Logger: bard.NewLogger(io.Discard)}.Execute()).To(Equal(map[string]string{
					"CLASSPATH": "test-classpath:test-security-providers-classpath",
				}))
			})
		})
	})

}
