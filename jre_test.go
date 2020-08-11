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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
)

func testJRE(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "jre-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes JRE", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-jre.tar.gz",
			SHA256: "b0cb4e1d28229bc92c19831f931863008b0193075a8d35d85240116c372a9c36",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j := libjvm.NewJRE(dep, dc, NoContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(filepath.Join(layer.Path, "fixture-marker")).To(BeARegularFile())
	})

	it("marks layer for build", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-jre.tar.gz",
			SHA256: "b0cb4e1d28229bc92c19831f931863008b0193075a8d35d85240116c372a9c36",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j := libjvm.NewJRE(dep, dc, BuildContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Build).To(BeTrue())
		Expect(layer.Cache).To(BeTrue())
		Expect(layer.BuildEnvironment["JAVA_HOME.default"]).To(Equal(layer.Path))
	})

	it("marks layer for launch", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-jre.tar.gz",
			SHA256: "b0cb4e1d28229bc92c19831f931863008b0193075a8d35d85240116c372a9c36",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j := libjvm.NewJRE(dep, dc, LaunchContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(layer.LaunchEnvironment["JAVA_HOME.override"]).To(Equal(layer.Path))
		Expect(layer.LaunchEnvironment["MALLOC_ARENA_MAX.override"]).To(Equal("2"))
		Expect(layer.Profile["active-processor-count.sh"]).To(Equal(`JAVA_TOOL_OPTIONS="${JAVA_TOOL_OPTIONS} -XX:ActiveProcessorCount=$(nproc)" || exit $?
export JAVA_TOOL_OPTIONS
`))
		Expect(layer.Profile["java-tool-options.sh"]).To(Equal(`[[ -z "${JAVA_OPTS+x}" ]] && return

export JAVA_TOOL_OPTIONS="${JAVA_OPTS} ${JAVA_TOOL_OPTIONS}"
`))
	})
}
