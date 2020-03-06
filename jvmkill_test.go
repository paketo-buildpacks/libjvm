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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketoio/libjvm"
	"github.com/paketoio/libpak"
	"github.com/sclevine/spec"
)

func testJVMKill(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "jvmkill-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it("contributes JVMKill", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-jvmkill.so",
			SHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j := libjvm.NewJVMKill(dep, dc, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(filepath.Join(layer.Path, "stub-jvmkill.so")).To(BeARegularFile())
		Expect(layer.SharedEnvironment["JAVA_OPTS.append"]).To(Equal(fmt.Sprintf(" -agentpath:%s/stub-jvmkill.so=printHeapHistogram=1", layer.Path)))
	})
}
