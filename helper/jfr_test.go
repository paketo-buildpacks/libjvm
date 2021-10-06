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
	"github.com/paketo-buildpacks/libjvm/helper"
	"github.com/sclevine/spec"
)

func testJFR(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		jfr = helper.JFR{}
	)

	it("returns if $BPL_JAVA_FLIGHT_RECORDER_ENABLED is not set", func() {
		Expect(jfr.Execute()).To(BeNil())
	})
	context("$BPL_JAVA_FLIGHT_RECORDER_ENABLED", func() {
		it.Before(func() {
			Expect(os.Setenv("BPL_JAVA_FLIGHT_RECORDER_ENABLED", "true")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BPL_JAVA_FLIGHT_RECORDER_ENABLED")).To(Succeed())
		})

		it("contributes base JFR configuration", func() {
			Expect(jfr.Execute()).To(Equal(map[string]string{
				"JAVA_TOOL_OPTIONS": "-XX:StartFlightRecording=",
			}))
		})

		context("$BPL_JFR_ARGS is set", func() {
			it("contributes all arguments to JFR configuration", func() {
				Expect(os.Setenv("BPL_JFR_ARGS", "filename=/tmp/test.jfr,name=file,delay=60s,dumponexit=true,duration=10s,maxage=1d,maxsize=1024m,path-to-gc-roots=true,settings=true")).To(Succeed())
				Expect(jfr.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:StartFlightRecording=filename=/tmp/test.jfr,name=file,delay=60s,dumponexit=true,duration=10s,maxage=1d,maxsize=1024m,path-to-gc-roots=true,settings=true"}))
			})
			it("returns an error if a JFR argument is empty", func() {
				Expect(os.Setenv("BPL_JFR_ARGS", "filename=/tmp/test.jfr,name=")).To(Succeed())
				_, err := jfr.Execute()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid Flight Recorder argument: name="))
			})
			it("returns an error if JFR arguments are not parsable", func() {
				Expect(os.Setenv("BPL_JFR_ARGS", ",filename=/tmp/test.jfr,")).To(Succeed())
				_, err := jfr.Execute()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("unable to parse Flight Recorder arguments: ,filename=/tmp/test.jfr,"))
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
