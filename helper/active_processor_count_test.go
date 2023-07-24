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
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm/helper"
	"github.com/paketo-buildpacks/libpak/bard"
)

func testActiveProcessorCount(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("configures active processor count", func() {
		Expect(helper.ActiveProcessorCount{Logger: bard.NewLogger(io.Discard)}.Execute()).
			To(Equal(map[string]string{"JAVA_TOOL_OPTIONS": fmt.Sprintf("-XX:ActiveProcessorCount=%d", runtime.NumCPU())}))
	})

	context("$JAVA_TOOL_OPTIONS", func() {

		it.Before(func() {
			Expect(os.Setenv("JAVA_TOOL_OPTIONS", "test-java-tool-options")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
		})

		it("configures active processor count", func() {
			Expect(helper.ActiveProcessorCount{Logger: bard.NewLogger(io.Discard)}.Execute()).
				To(Equal(map[string]string{"JAVA_TOOL_OPTIONS": fmt.Sprintf("test-java-tool-options -XX:ActiveProcessorCount=%d", runtime.NumCPU())}))
		})

	})

	context("-XX:ActiveProcessorCount", func() {
		it.Before(func() {
			Expect(os.Setenv("JAVA_TOOL_OPTIONS", "-XX:ActiveProcessorCount=0")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
		})

		it("does not override active processor count", func() {
			Expect(helper.ActiveProcessorCount{Logger: bard.NewLogger(io.Discard)}.Execute()).To(BeNil())
		})
	})
}
