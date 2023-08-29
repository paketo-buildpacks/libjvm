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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm/v2/helper"
	"github.com/paketo-buildpacks/libpak/v2/log"
	"github.com/sclevine/spec"
)

func testJFR(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		jfr = helper.JFR{Logger: log.NewPaketoLogger(io.Discard)}
	)

	it("returns if $BPL_JFR_ENABLED is not set", func() {
		Expect(jfr.Execute()).To(BeNil())
	})
	context("$BPL_JFR_ENABLED", func() {
		it.Before(func() {
			Expect(os.Setenv("BPL_JFR_ENABLED", "true")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BPL_JFR_ENABLED")).To(Succeed())
		})

		it("contributes base JFR configuration", func() {
			Expect(jfr.Execute()).To(Equal(map[string]string{
				"JAVA_TOOL_OPTIONS": fmt.Sprintf("-XX:StartFlightRecording=dumponexit=true,filename=%s", filepath.Join(os.TempDir(), "recording.jfr")),
			}))
		})

		context("$BPL_JFR_ARGS is set", func() {
			it("contributes all arguments to JFR configuration", func() {
				Expect(os.Setenv("BPL_JFR_ARGS", "filename=/tmp/test.jfr,name=file,delay=60s,dumponexit=true,duration=10s,maxage=1d,maxsize=1024m,path-to-gc-roots=true,settings=true")).To(Succeed())
				Expect(jfr.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:StartFlightRecording=filename=/tmp/test.jfr,name=file,delay=60s,dumponexit=true,duration=10s,maxage=1d,maxsize=1024m,path-to-gc-roots=true,settings=true"}))
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
				Expect(os.Setenv("BPL_JFR_ARGS", "filename=/tmp/test.jfr,name=file")).To(Succeed())
				Expect(jfr.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "test-java-tool-options -XX:StartFlightRecording=filename=/tmp/test.jfr,name=file",
				}))
			})
		})

	})
}
