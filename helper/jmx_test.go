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
	"github.com/paketo-buildpacks/libjvm/helper"
	"github.com/paketo-buildpacks/libpak/v2/bard"
	"github.com/sclevine/spec"
)

func testJMX(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		j = helper.JMX{Logger: bard.NewLogger(io.Discard)}
	)

	it("returns if $BPL_JMX_ENABLED is not set", func() {
		Expect(j.Execute()).To(BeNil())
	})

	context("$BPL_JMX_ENABLED", func() {
		it.Before(func() {
			Expect(os.Setenv("BPL_JMX_ENABLED", "true")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BPL_JMX_ENABLED")).To(Succeed())
		})

		it("contributes configuration", func() {
			Expect(j.Execute()).To(Equal(map[string]string{
				"JAVA_TOOL_OPTIONS": "-Djava.rmi.server.hostname=127.0.0.1 -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.port=5000 -Dcom.sun.management.jmxremote.rmi.port=5000",
			}))
		})

		context("$BPL_JMX_PORT", func() {
			it.Before(func() {
				Expect(os.Setenv("BPL_JMX_PORT", "5001")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPL_JMX_PORT")).To(Succeed())
			})

			it("contributes port configuration from $BPL_JMX_PORT", func() {
				Expect(j.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-Djava.rmi.server.hostname=127.0.0.1 -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.port=5001 -Dcom.sun.management.jmxremote.rmi.port=5001",
				}))
			})
		})

		context("$JAVA_TOOL_OPTIONS", func() {
			it.Before(func() {
				Expect(os.Setenv("JAVA_TOOL_OPTIONS", "test-java-tool-options")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
			})

			it("contributes configuration appended to existing $JAVA_TOOL_OPTIONS", func() {
				Expect(j.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "test-java-tool-options -Djava.rmi.server.hostname=127.0.0.1 -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.port=5000 -Dcom.sun.management.jmxremote.rmi.port=5000",
				}))
			})
		})
	})

}
