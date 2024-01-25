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
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/anthonydahanne/libjvm/helper"
)

func testSecurityProvidersClasspath8(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("does nothing if $SECURITY_PROVIDERS_CLASSPATH is no set", func() {
		Expect(helper.SecurityProvidersClasspath8{}.Execute()).To(BeNil())
	})

	context("$SECURITY_PROVIDERS_CLASSPATH", func() {

		it.Before(func() {
			Expect(os.Setenv("SECURITY_PROVIDERS_CLASSPATH", "test-dir-1/test-classpath-1:test-dir-2/test-classpath-2")).
				To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("SECURITY_PROVIDERS_CLASSPATH")).To(Succeed())
		})

		it("returns error if $BPI_JVM_EXT_DIR is not set", func() {
			_, err := helper.SecurityProvidersClasspath8{}.Execute()

			Expect(err).To(MatchError("$BPI_JVM_EXT_DIR must be set"))
		})

		context("$BPI_JVM_EXT_DIR", func() {

			it.Before(func() {
				Expect(os.Setenv("BPI_JVM_EXT_DIR", "test-bpi-jvm-ext-dir")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPI_JVM_EXT_DIR")).To(Succeed())
			})

			it("return $JAVA_TOOL_OPTIONS with $SECURITY_PROVIDERS_CLASSPATH only", func() {
				Expect(helper.SecurityProvidersClasspath8{}.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-Djava.ext.dirs=test-bpi-jvm-ext-dir:test-dir-1:test-dir-2",
				}))
			})

			context("$JAVA_TOOL_OPTIONS", func() {
				it.Before(func() {
					Expect(os.Setenv("JAVA_TOOL_OPTIONS", "test-java-tool-options")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
				})

				it("return $JAVA_TOOL_OPTIONS with $SECURITY_PROVIDERS_CLASSPATH included", func() {
					Expect(helper.SecurityProvidersClasspath8{}.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "test-java-tool-options -Djava.ext.dirs=test-bpi-jvm-ext-dir:test-dir-1:test-dir-2",
					}))
				})
			})
		})

	})

}
