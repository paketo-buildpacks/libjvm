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
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
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

	it("detecting JVM version from default", func() {
		jvmVersion := libjvm.JVMVersion{Logger: logger}

		cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
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

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})
	})

	context("detecting JVM version", func() {
		it.Before(func() {
			temp, err := prepareAppWithEntry("Build-Jdk: 1.8")
			Expect(err).ToNot(HaveOccurred())
			appPath = temp
		})

		it.After(func() {
			os.RemoveAll(appPath)
		})

		it("from manifest via Build-Jdk-Spec", func() {
			jvmVersion := libjvm.JVMVersion{Logger: logger}

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("8"))
		})
	})

	context("detecting JVM version", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_JVM_VERSION", "17")).To(Succeed())
			temp, err := prepareAppWithEntry("Build-Jdk: 1.8")
			Expect(err).ToNot(HaveOccurred())
			appPath = temp
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_JVM_VERSION")).To(Succeed())
			os.RemoveAll(appPath)
		})

		it("prefers environment variable over manifest", func() {
			jvmVersion := libjvm.JVMVersion{Logger: logger}

			cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
			Expect(err).ToNot(HaveOccurred())
			version, err := jvmVersion.GetJVMVersion(appPath, cr)
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("17"))
		})
	})

}

func prepareAppWithEntry(entry string) (string, error) {
	temp, err := ioutil.TempDir("", "jre-app")
	if err != nil {
		return "", err
	}
	err = os.Mkdir(filepath.Join(temp, "META-INF"), 0744)
	if err != nil {
		return "", err
	}
	manifest := filepath.Join(temp, "META-INF", "MANIFEST.MF")
	manifestContent := []byte(entry)
	err = ioutil.WriteFile(manifest, manifestContent, 0644)
	if err != nil {
		return "", err
	}
	return temp, nil
}
