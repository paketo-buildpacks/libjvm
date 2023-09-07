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
	"github.com/paketo-buildpacks/libpak/v2/log"
)

func testJavaOpts(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("$JAVA_TOOL_OPTIONS", func() {

		it.Before(func() {
			Expect(os.Setenv("JAVA_TOOL_OPTIONS", "test-java-tool-options")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
		})

		it("return nil if JAVA_OPTS is not set", func() {
			Expect(helper.JavaOpts{Logger: log.NewPaketoLogger(io.Discard)}.Execute()).To(BeNil())
		})

		context("$JAVA_OPTS", func() {
			it.Before(func() {
				Expect(os.Setenv("JAVA_OPTS", "test-java-opts")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("JAVA_OPTS")).To(Succeed())
			})

			it("return $JAVA_TOOL_OPTIONS with $JAVA_OPTS included", func() {
				Expect(helper.JavaOpts{Logger: log.NewPaketoLogger(io.Discard)}.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "test-java-tool-options test-java-opts",
				}))
			})
		})
	})

	context("$JAVA_OPTS", func() {

		it.Before(func() {
			Expect(os.Setenv("JAVA_OPTS", "test-java-opts")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("JAVA_OPTS")).To(Succeed())
		})

		it("return $JAVA_TOOL_OPTIONS with $JAVA_OPTS only", func() {
			Expect(helper.JavaOpts{Logger: log.NewPaketoLogger(io.Discard)}.Execute()).To(Equal(map[string]string{
				"JAVA_TOOL_OPTIONS": "test-java-opts",
			}))
		})
	})

}
