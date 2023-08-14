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

	"github.com/paketo-buildpacks/libpak/v2"
	"github.com/paketo-buildpacks/libpak/v2/bard"

	"github.com/buildpacks/libcnb/v2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
)

func testJVMVersion(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect    = NewWithT(t).Expect
		appPath   string
		logger    bard.Logger
		buildpack libcnb.Buildpack
	)

	it.Before(func() {
		var err error

		appPath, err = ioutil.TempDir("", "application")
		Expect(err).NotTo(HaveOccurred())

		buildpack = libcnb.Buildpack{
			Metadata: map[string]interface{}{
				"configurations": []map[string]interface{}{
					{
						"name":    "BP_JVM_VERSION",
						"default": "1.1.1",
					},
				},
			},
		}
		logger = bard.NewLogger(ioutil.Discard)
	})

	it.After(func() {
		Expect(os.RemoveAll(appPath)).To(Succeed())
	})

	it("detecting JVM version from default", func() {
		jvmVersion := libjvm.JVMVersion{Logger: logger}

		bpm, err := libpak.NewBuildModuleMetadata(buildpack.Metadata)
		Expect(err).ToNot(HaveOccurred())

		cr, err := libpak.NewConfigurationResolver(bpm, &logger)
		Expect(err).ToNot(HaveOccurred())
		version, err := jvmVersion.GetJVMVersion(appPath, cr)
		Expect(err).ToNot(HaveOccurred())
		Expect(version).To(Equal("1.1.1"))
	})

	context("detecting JVM version", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_JVM_VERSION", "17")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_JVM_VERSION")).To(Succeed())
		})

		it("from environment variable", func() {
			jvmVersion := libjvm.JVMVersion{Logger: logger}

			bpm, err := libpak.NewBuildModuleMetadata(buildpack.Metadata)
			Expect(err).ToNot(HaveOccurred())

			cr, err := libpak.NewConfigurationResolver(bpm, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})
	})

	context("detecting JVM version", func() {
		it.Before(func() {
			Expect(prepareAppWithEntry(appPath, "Build-Jdk: 1.8")).ToNot(HaveOccurred())
		})

		it("from manifest via Build-Jdk-Spec", func() {
			jvmVersion := libjvm.JVMVersion{Logger: logger}

			bpm, err := libpak.NewBuildModuleMetadata(buildpack.Metadata)
			Expect(err).ToNot(HaveOccurred())

			cr, err := libpak.NewConfigurationResolver(bpm, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("8"))
		})
	})

	context("detecting JVM version", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_JVM_VERSION", "17")).To(Succeed())
			Expect(prepareAppWithEntry(appPath, "Build-Jdk: 1.8")).ToNot(HaveOccurred())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_JVM_VERSION")).To(Succeed())
		})

		it("prefers environment variable over manifest", func() {
			jvmVersion := libjvm.JVMVersion{Logger: logger}

			bpm, err := libpak.NewBuildModuleMetadata(buildpack.Metadata)
			Expect(err).ToNot(HaveOccurred())

			cr, err := libpak.NewConfigurationResolver(bpm, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})
	})

	context("detecting JVM version", func() {
		var sdkmanrcFile string

		it.Before(func() {
			sdkmanrcFile = filepath.Join(appPath, ".sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=17.0.2-tem`), 0644)).To(Succeed())
		})

		it("from .sdkmanrc file", func() {
			jvmVersion := libjvm.JVMVersion{Logger: logger}

			bpm, err := libpak.NewBuildModuleMetadata(buildpack.Metadata)
			Expect(err).ToNot(HaveOccurred())

			cr, err := libpak.NewConfigurationResolver(bpm, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})
	})

	context("detecting JVM version", func() {
		var sdkmanrcFile string

		it.Before(func() {
			sdkmanrcFile = filepath.Join(appPath, ".sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=17.0.2-tem
java=11.0.2-tem`), 0644)).To(Succeed())
		})

		it("picks first from .sdkmanrc file if there are multiple", func() {
			jvmVersion := libjvm.JVMVersion{Logger: logger}

			bpm, err := libpak.NewBuildModuleMetadata(buildpack.Metadata)
			Expect(err).ToNot(HaveOccurred())

			cr, err := libpak.NewConfigurationResolver(bpm, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})
	})
}

func prepareAppWithEntry(appPath, entry string) error {
	err := os.Mkdir(filepath.Join(appPath, "META-INF"), 0744)
	if err != nil {
		return err
	}
	manifest := filepath.Join(appPath, "META-INF", "MANIFEST.MF")
	manifestContent := []byte(entry)
	err = ioutil.WriteFile(manifest, manifestContent, 0644)
	if err != nil {
		return err
	}
	return nil
}
