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
	"github.com/paketo-buildpacks/libjvm"
	"github.com/sclevine/spec"
)

func testLinkLocalDNS(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Buildpack.Info.Name = "test-name"

		ctx.Buildpack.Path, err = ioutil.TempDir("", "link-local-dns-buildpack")
		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = ioutil.TempDir("", "link-local-dns-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Buildpack.Path)).To(Succeed())
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes link-local DNS", func() {
		Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "link-local-dns"), []byte{}, 0755)).To(Succeed())

		l := libjvm.NewLinkLocalDNS(ctx.Buildpack, NoContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = l.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(filepath.Join(layer.Path, "bin", "link-local-dns")).To(BeARegularFile())
		Expect(layer.Profile["link-local-dns.sh"]).To(Equal(`link-local-dns
`))
	})

	it("marks layer for build", func() {
		Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "link-local-dns"), []byte{}, 0755)).To(Succeed())

		l := libjvm.NewLinkLocalDNS(ctx.Buildpack, BuildContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = l.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Build).To(BeTrue())
		Expect(layer.Cache).To(BeTrue())
	})

	it("marks layer for launch", func() {
		Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "link-local-dns"), []byte{}, 0755)).To(Succeed())

		l := libjvm.NewLinkLocalDNS(ctx.Buildpack, LaunchContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = l.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
	})
}
