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

package libjvm_test

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/libpak/v2/log"

	"github.com/buildpacks/libcnb/v2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm/v2"
)

func testGenerate(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		ctx    libcnb.GenerateContext
		path   string
	)

	it.Before(func() {
		ctx = libcnb.GenerateContext{}

		var err error
		path, err = ioutil.TempDir("", "outputDir")
		Expect(err).NotTo(HaveOccurred())
		ctx.OutputDirectory = path
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("does not invoke generate if no JDK/JRE", func() {
		ctx.Extension.API = "0.9"
		ctx.Extension.Metadata = map[string]interface{}{}
		ctx.StackID = "test-stack-id"

		invoked := false
		_, err := libjvm.NewGenerate(log.NewPaketoLogger(io.Discard), func(ctx libjvm.GenerateContentContext) (libjvm.GenerateContentResult, error) {
			invoked = true
			return libjvm.GenerateContentResult{
				ExtendConfig:    libjvm.ExtendConfig{Build: libjvm.ExtendImageConfig{Args: []libjvm.ExtendImageConfigArg{}}},
				BuildDockerfile: strings.NewReader("buildDockerfileContent"),
				RunDockerfile:   strings.NewReader("buildDockerfileContent"),
				GenerateResult:  libcnb.NewGenerateResult(),
			}, nil
		}).Generate(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(invoked).To(Equal(false))
	})

	it("invokes if JDK requested", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jdk"})
		ctx.Extension.API = "0.9"
		ctx.StackID = "test-stack-id"

		invoked := false
		_, err := libjvm.NewGenerate(log.NewPaketoLogger(io.Discard), func(ctx libjvm.GenerateContentContext) (libjvm.GenerateContentResult, error) {
			invoked = true
			return libjvm.GenerateContentResult{
				ExtendConfig:    libjvm.ExtendConfig{Build: libjvm.ExtendImageConfig{Args: []libjvm.ExtendImageConfigArg{}}},
				BuildDockerfile: strings.NewReader("buildDockerfileContent"),
				RunDockerfile:   strings.NewReader("buildDockerfileContent"),
				GenerateResult:  libcnb.NewGenerateResult(),
			}, nil
		}).Generate(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(invoked).To(Equal(true))
	})

	it("invokes if JRE requested", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Extension.API = "0.9"
		ctx.StackID = "test-stack-id"

		invoked := false
		_, err := libjvm.NewGenerate(log.NewPaketoLogger(io.Discard), func(ctx libjvm.GenerateContentContext) (libjvm.GenerateContentResult, error) {
			invoked = true
			return libjvm.GenerateContentResult{
				ExtendConfig:    libjvm.ExtendConfig{Build: libjvm.ExtendImageConfig{Args: []libjvm.ExtendImageConfigArg{}}},
				BuildDockerfile: strings.NewReader("buildDockerfileContent"),
				RunDockerfile:   strings.NewReader("buildDockerfileContent"),
				GenerateResult:  libcnb.NewGenerateResult(),
			}, nil
		}).Generate(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(invoked).To(Equal(true))
	})

}
