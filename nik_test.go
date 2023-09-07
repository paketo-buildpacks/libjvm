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

	"github.com/buildpacks/libcnb/v2"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm/v2"
	"github.com/paketo-buildpacks/libpak/v2"
	"github.com/paketo-buildpacks/libpak/v2/effect"
	"github.com/paketo-buildpacks/libpak/v2/effect/mocks"
	"github.com/paketo-buildpacks/libpak/v2/log"
	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"
)

func testNIK(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cl = libjvm.CertificateLoader{
			CertDirs: []string{filepath.Join("testdata", "certificates")},
			Logger:   log.NewDiscardLogger(),
		}

		ctx      libcnb.BuildContext
		executor *mocks.Executor
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "jdk-layers")
		Expect(err).NotTo(HaveOccurred())

		executor = &mocks.Executor{}
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes JDK without NIK", func() {
		executor.On("Execute", mock.Anything).Return(nil)
		dep := libpak.BuildModuleDependency{
			Version: "11.0.0",
			URI:     "https://localhost/stub-jdk-11.tar.gz",
			SHA256:  "e40a6ddb7d74d78a6d5557380160a174b1273813db1caf9b1f7bcbfe1578e818",
		}
		dc := libpak.DependencyCache{CachePath: "testdata", Logger: log.NewDiscardLogger()}

		n, err := libjvm.NewNIK(dep, nil, dc, cl, "", nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(n.LayerContributor.ExpectedMetadata.(map[string]interface{})["cert-dir"]).To(HaveLen(4))

		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		err = n.Contribute(&layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(executor.Calls).To(HaveLen(0))

		Expect(layer.Build).To(BeTrue())
		Expect(layer.Cache).To(BeTrue())
		Expect(filepath.Join(layer.Path, "fixture-marker")).To(BeARegularFile())
		Expect(layer.BuildEnvironment["JAVA_HOME.override"]).To(Equal(layer.Path))
		Expect(layer.BuildEnvironment["JDK_HOME.override"]).To(Equal(layer.Path))
	})

	it("contributes native image to JDK", func() {
		executor.On("Execute", mock.Anything).Return(nil)

		jdkDep := libpak.BuildModuleDependency{
			Version: "11.0.0",
			URI:     "https://localhost/stub-jdk-11.tar.gz",
			SHA256:  "e40a6ddb7d74d78a6d5557380160a174b1273813db1caf9b1f7bcbfe1578e818",
		}
		niDep := &libpak.BuildModuleDependency{
			URI:    "https://localhost/stub-native-image.jar",
			SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		}
		dc := libpak.DependencyCache{CachePath: "testdata", Logger: log.NewDiscardLogger()}

		n, err := libjvm.NewNIK(jdkDep, niDep, dc, cl, "bin/gu", []string{"install", "--local-file"})
		Expect(err).NotTo(HaveOccurred())

		n.Executor = executor

		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		err = n.Contribute(&layer)
		Expect(err).NotTo(HaveOccurred())

		executor := executor.Calls[0].Arguments[0].(effect.Execution)
		Expect(executor.Command).To(Equal(filepath.Join(layer.Path, "bin", "gu")))
		Expect(executor.Args).To(Equal([]string{"install", "--local-file", filepath.Join("testdata", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "stub-native-image.jar")}))
		Expect(executor.Dir).To(Equal(layer.Path))
	})

	it("updates before Java 9 certificates", func() {
		dep := libpak.BuildModuleDependency{
			Version: "8.0.0",
			URI:     "https://localhost/stub-jdk-8.tar.gz",
			SHA256:  "6860fb9a9a66817ec285fac64c342b678b0810656b1f2413f063911a8bde6447",
		}
		dc := libpak.DependencyCache{CachePath: "testdata", Logger: log.NewDiscardLogger()}

		n, err := libjvm.NewNIK(dep, nil, dc, cl, "", nil)
		Expect(err).NotTo(HaveOccurred())

		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		err = n.Contribute(&layer)
		Expect(err).NotTo(HaveOccurred())

		in, err := os.Open(filepath.Join(layer.Path, "jre", "lib", "security", "cacerts"))
		Expect(err).NotTo(HaveOccurred())
		defer in.Close()

		ks := keystore.New()
		err = ks.Load(in, []byte("changeit"))
		Expect(err).NotTo(HaveOccurred())

		Expect(ks.Aliases()).To(HaveLen(3))
	})

	it("updates after Java 9 certificates", func() {
		dep := libpak.BuildModuleDependency{
			Version: "11.0.0",
			URI:     "https://localhost/stub-jdk-11.tar.gz",
			SHA256:  "e40a6ddb7d74d78a6d5557380160a174b1273813db1caf9b1f7bcbfe1578e818",
		}
		dc := libpak.DependencyCache{CachePath: "testdata", Logger: log.NewDiscardLogger()}

		j, err := libjvm.NewNIK(dep, nil, dc, cl, "", nil)
		Expect(err).NotTo(HaveOccurred())

		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		err = j.Contribute(&layer)
		Expect(err).NotTo(HaveOccurred())

		in, err := os.Open(filepath.Join(layer.Path, "lib", "security", "cacerts"))
		Expect(err).NotTo(HaveOccurred())
		defer in.Close()

		ks := keystore.New()
		err = ks.Load(in, []byte("changeit"))
		Expect(err).NotTo(HaveOccurred())

		Expect(ks.Aliases()).To(HaveLen(3))
	})

}
