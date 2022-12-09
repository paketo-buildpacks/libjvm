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

	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
)

func testResolveMetadataVersion(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		logger     bard.Logger
		jvmVersion libjvm.JVMVersion
	)

	it.Before(func() {
		logger = bard.NewLogger(ioutil.Discard)
		jvmVersion = libjvm.JVMVersion{Logger: logger}
	})

	it.After(func() {
	})

	context("resolving required version from buildplan", func() {

		it("version cannot with no entries", func() {
			version, err := jvmVersion.ResolveMetadataVersion()
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(BeEmpty())
		})

		it("version cannot be resolved without metadata", func() {
			plan1 := libcnb.BuildpackPlanEntry{Name: "jdk"}
			plan2 := libcnb.BuildpackPlanEntry{Name: "jre"}

			version, err := jvmVersion.ResolveMetadataVersion(plan1, plan2)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(BeEmpty())
		})

		it("version can be resolved with single metadata", func() {
			plan1 := libcnb.BuildpackPlanEntry{Name: "jdk"}
			plan2 := libcnb.BuildpackPlanEntry{Name: "jre", Metadata: map[string]interface{}{"version": "17"}}

			version, err := jvmVersion.ResolveMetadataVersion(plan1, plan2)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})

		it("version can be resolved with multiple metadata requiring the same version", func() {
			plan1 := libcnb.BuildpackPlanEntry{Name: "jdk", Metadata: map[string]interface{}{"version": "17"}}
			plan2 := libcnb.BuildpackPlanEntry{Name: "jre", Metadata: map[string]interface{}{"version": "17"}}

			version, err := jvmVersion.ResolveMetadataVersion(plan1, plan2)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})

		it("error with multiple metadata requiring different versions", func() {
			plan1 := libcnb.BuildpackPlanEntry{Name: "jdk", Metadata: map[string]interface{}{"version": "17"}}
			plan2 := libcnb.BuildpackPlanEntry{Name: "jre", Metadata: map[string]interface{}{"version": "18"}}

			_, err := jvmVersion.ResolveMetadataVersion(plan1, plan2)
			Expect(err).To(MatchError(SatisfyAll(ContainSubstring("18"), ContainSubstring("17"))))
		})

	})

}

func testJVMVersion(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		appPath    string
		logger     bard.Logger
		buildpack  libcnb.Buildpack
		jvmVersion libjvm.JVMVersion
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
		jvmVersion = libjvm.JVMVersion{Logger: logger}
	})

	it.After(func() {
		Expect(os.RemoveAll(appPath)).To(Succeed())
	})

	context("no specific version requested by the user", func() {

		it("detecting JVM version from default", func() {
			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("1.1.1"))
		})

		it("prefer JVM version requested from metadata", func() {
			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "2")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("2"))
		})

		it("from manifest via Build-Jdk-Spec", func() {
			Expect(prepareAppWithEntry(appPath, "Build-Jdk: 1.8")).ToNot(HaveOccurred())

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("8"))
		})

		it("prefers required metadata version over manifest", func() {
			Expect(prepareAppWithEntry(appPath, "Build-Jdk: 1.8")).ToNot(HaveOccurred())

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "18")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("18"))
		})

	})

	context("BP_JVM_VERSION=17", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_JVM_VERSION", "17")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_JVM_VERSION")).To(Succeed())
		})

		it("from environment variable", func() {
			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})

		it("metadata should be ignored", func() {
			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "11")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})

		it("prefers environment variable over manifest", func() {
			Expect(prepareAppWithEntry(appPath, "Build-Jdk: 1.8")).ToNot(HaveOccurred())

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})

	})

	context("detecting JVM version from .sdkmanrc", func() {
		var sdkmanrcFile string

		it.Before(func() {
		})

		it("from .sdkmanrc file", func() {
			sdkmanrcFile = filepath.Join(appPath, ".sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=17.0.2-tem`), 0644)).To(Succeed())

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})

		it("prefers required metadata version over .sdkmanrc", func() {
			sdkmanrcFile = filepath.Join(appPath, ".sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=17.0.2-tem`), 0644)).To(Succeed())

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "18")
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("18"))
		})

		it("picks first from .sdkmanrc file if there are multiple", func() {
			sdkmanrcFile = filepath.Join(appPath, ".sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=17.0.2-tem
java=11.0.2-tem`), 0644)).To(Succeed())

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr, "")
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
