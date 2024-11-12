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
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/libpak/bard"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext

		nativeOptionBundledWithJDK = libjvm.WithNativeImage(libjvm.NativeImage{
			BundledWithJDK: true,
		})
		nativeOptionSeparateFromJDK = libjvm.WithNativeImage(libjvm.NativeImage{
			BundledWithJDK: false,
			CustomCommand:  "/bin/gu",
			CustomArgs:     []string{"install", "--local-file"},
		})
		nativeOptionMissingCommand = libjvm.WithNativeImage(libjvm.NativeImage{
			BundledWithJDK: false,
			CustomCommand:  "",
			CustomArgs:     []string{"install", "--local-file"},
		})
	)

	it.Before(func() {
		t.Setenv("BP_ARCH", "amd64")
	})

	it.After(func() {
		Expect(os.Unsetenv("BP_JVM_VERSION")).To(Succeed())
	})

	it("contributes JDK", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jdk"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jdk",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name()).To(Equal("jdk"))

		Expect(result.BOM.Entries).To(HaveLen(1))
		Expect(result.BOM.Entries[0].Name).To(Equal("jdk"))
		Expect(result.BOM.Entries[0].Launch).To(BeFalse())
		Expect(result.BOM.Entries[0].Build).To(BeTrue())
	})

	it("contributes JRE", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.API = "0.6"
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jre",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(3))
		Expect(result.Layers[0].Name()).To(Equal("jre"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[2].Name()).To(Equal("java-security-properties"))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("jre"))
		Expect(result.BOM.Entries[0].Launch).To(BeTrue())
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
		Expect(result.BOM.Entries[1].Launch).To(BeTrue())
	})

	it("contributes available next JRE version when Manifest refers to not available version", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.API = "0.6"

		appPath, err := os.MkdirTemp("", "application")
		Expect(prepareAppWithEntry(appPath, "Build-Jdk: 22")).ToNot(HaveOccurred())

		ctx.Application.Path = appPath
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jre",
					"version": "8.0.432",
					"stacks":  []interface{}{"test-stack-id"},
				},
				{
					"id":      "jre",
					"version": "23.0.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
				{
					"id":      "jre",
					"version": "43.43.43",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(3))
		Expect(result.Layers[0].Name()).To(Equal("jre"))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[2].Name()).To(Equal("java-security-properties"))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("jre"))
		Expect(result.BOM.Entries[0].Metadata["version"]).To(Equal("23.0.1"))
		Expect(result.BOM.Entries[0].Launch).To(BeTrue())
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
		Expect(result.BOM.Entries[1].Launch).To(BeTrue())
	})

	it("provides meaningful error message if user requested via sdkmanrc a non available JRE", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.API = "0.6"

		appPath, err := os.MkdirTemp("", "application")
		sdkmanrcFile := filepath.Join(appPath, ".sdkmanrc")
		Expect(os.WriteFile(sdkmanrcFile, []byte(`java=20.0.2-tem`), 0644)).To(Succeed())

		ctx.Application.Path = appPath
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jre",
					"version": "8.0.432",
					"stacks":  []interface{}{"test-stack-id"},
				},
				{
					"id":      "jre",
					"version": "23.0.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
				{
					"id":      "jre",
					"version": "43.43.43",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		_, err = libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unable to find dependency for JRE 20 even as a JDK - make sure the buildpack includes the Java version you have requested"))
		Expect(err.Error()).To(ContainSubstring("no valid dependencies for jre, 20, and test-stack-id in [(jre, 8.0.432, [test-stack-id]) (jre, 23.0.1, [test-stack-id]) (jre, 43.43.43, [test-stack-id])]"))
	})

	it("provides meaningful error message if user requested via env.var a non available JRE", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.API = "0.6"

		Expect(os.Setenv("BP_JVM_VERSION", "24")).To(Succeed())

		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jre",
					"version": "8.0.432",
					"stacks":  []interface{}{"test-stack-id"},
				},
				{
					"id":      "jre",
					"version": "23.0.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
				{
					"id":      "jre",
					"version": "43.43.43",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		_, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unable to find dependency for JRE 24 even as a JDK - make sure the buildpack includes the Java version you have requested"))
		Expect(err.Error()).To(ContainSubstring("no valid dependencies for jre, 24, and test-stack-id in [(jre, 8.0.432, [test-stack-id]) (jre, 23.0.1, [test-stack-id]) (jre, 43.43.43, [test-stack-id])]"))
	})

	it("contributes security-providers-classpath-8 before Java 9", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jre",
					"version": "8.0.0",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{
			"java-opts",
			"jvm-heap",
			"link-local-dns",
			"memory-calculator",
			"security-providers-configurer",
			"jmx",
			"jfr",
			"openssl-certificate-loader",
			"security-providers-classpath-8",
			"debug-8",
			"active-processor-count",
		}))
	})

	it("contributes security-providers-classpath-9 after Java 9", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jre",
					"version": "11.0.0",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{
			"java-opts",
			"jvm-heap",
			"link-local-dns",
			"memory-calculator",
			"security-providers-configurer",
			"jmx",
			"jfr",
			"openssl-certificate-loader",
			"security-providers-classpath-9",
			"debug-9",
			"nmt",
			"active-processor-count",
		}))
	})

	it("contributes active-processor-count before Java 17", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jre",
					"version": "11.0.0",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{
			"java-opts",
			"jvm-heap",
			"link-local-dns",
			"memory-calculator",
			"security-providers-configurer",
			"jmx",
			"jfr",
			"openssl-certificate-loader",
			"security-providers-classpath-9",
			"debug-9",
			"nmt",
			"active-processor-count",
		}))
	})

	it("does not contribute active-processor-count with Java 17 and later", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jre",
					"version": "17.0.0",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{
			"java-opts",
			"jvm-heap",
			"link-local-dns",
			"memory-calculator",
			"security-providers-configurer",
			"jmx",
			"jfr",
			"openssl-certificate-loader",
			"security-providers-classpath-9",
			"debug-9",
			"nmt",
		}))
	})

	it("contributes JDK when no JRE and only a JRE is wanted", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.API = "0.6"
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jdk",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[0].Name()).To(Equal("jdk"))
		Expect(result.Layers[0].(libjvm.JRE).LayerContributor.Dependency.ID).To(Equal("jdk"))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("jdk"))
		Expect(result.BOM.Entries[0].Launch).To(BeTrue())
		Expect(result.BOM.Entries[0].Build).To(BeTrue())
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
		Expect(result.BOM.Entries[1].Launch).To(BeTrue())
	})

	it("contributes JDK when no JRE and both a JDK and JRE are wanted", func() {
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jdk", Metadata: LaunchContribution})
		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
		ctx.Buildpack.API = "0.6"
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "jdk",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[0].Name()).To(Equal("jdk"))
		Expect(result.Layers[0].(libjvm.JRE).LayerContributor.Dependency.ID).To(Equal("jdk"))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("jdk"))
		Expect(result.BOM.Entries[0].Launch).To(BeTrue())
		Expect(result.BOM.Entries[0].Build).To(BeTrue())
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
		Expect(result.BOM.Entries[1].Launch).To(BeTrue())
	})

	it("contributes NIK API <= 0.6", func() {

		ctx.Plan.Entries = append(
			ctx.Plan.Entries,
			libcnb.BuildpackPlanEntry{Name: "jdk", Metadata: map[string]interface{}{}},
			libcnb.BuildpackPlanEntry{Name: "native-image-builder"},
		)
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "native-image-svm",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.Buildpack.API = "0.6"
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionBundledWithJDK).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name()).To(Equal("native-image-svm"))

		Expect(result.BOM.Entries).To(HaveLen(1))
		Expect(result.BOM.Entries[0].Name).To(Equal("native-image-svm"))
		Expect(result.BOM.Entries[0].Launch).To(BeFalse())
		Expect(result.BOM.Entries[0].Build).To(BeTrue())
	})

	it("contributes NIK API >= 0.7", func() {
		ctx.Plan.Entries = append(
			ctx.Plan.Entries,
			libcnb.BuildpackPlanEntry{Name: "jdk", Metadata: map[string]interface{}{}},
			libcnb.BuildpackPlanEntry{Name: "native-image-builder"},
		)
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "native-image-svm",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
					"cpes":    []interface{}{"cpe:2.3:a:bellsoft:nik:1.1.1:*:*:*:*:*:*:*"},
					"purl":    "pkg:generic/provider-nik@1.1.1?arch=amd64",
				},
			},
		}
		ctx.Buildpack.API = "0.7"
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionBundledWithJDK).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name()).To(Equal("native-image-svm"))

		Expect(result.BOM.Entries).To(HaveLen(1))
		Expect(result.BOM.Entries[0].Name).To(Equal("native-image-svm"))
		Expect(result.BOM.Entries[0].Launch).To(BeFalse())
		Expect(result.BOM.Entries[0].Build).To(BeTrue())
	})

	context("native image enabled for API 0.7+ (not bundled with JDK)", func() {
		it("contributes native image dependency", func() {
			ctx.Plan.Entries = append(ctx.Plan.Entries,
				libcnb.BuildpackPlanEntry{
					Name: "jdk",
				},
				libcnb.BuildpackPlanEntry{
					Name: "native-image-builder",
				},
			)
			ctx.Buildpack.Metadata = map[string]interface{}{
				"dependencies": []map[string]interface{}{
					{
						"id":      "jdk",
						"version": "1.1.1",
						"stacks":  []interface{}{"test-stack-id"},
						"cpes":    []string{"cpe:2.3:a:oracle:graalvm:21.2.0:*:*:*:community:*:*:*"},
						"purl":    "pkg:generic/graalvm-jdk@21.2.0",
					},
					{
						"id":      "native-image-svm",
						"version": "2.2.2",
						"stacks":  []interface{}{"test-stack-id"},
						"cpes":    []string{"cpe:2.3:a:oracle:graalvm:21.2.0:*:*:*:community:*:*:*"},
						"purl":    "pkg:generic/graalvm-svm@21.2.0",
					},
				},
			}
			ctx.StackID = "test-stack-id"
			ctx.Buildpack.API = "0.7"

			result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionSeparateFromJDK).Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			Expect(result.Layers[0].Name()).To(Equal("nik"))
			Expect(result.Layers[0].(libjvm.NIK).NativeDependency).NotTo(BeNil())

			Expect(result.BOM.Entries).To(HaveLen(2))
			Expect(result.BOM.Entries[0].Name).To(Equal("jdk"))
			Expect(result.BOM.Entries[0].Launch).To(BeFalse())
			Expect(result.BOM.Entries[0].Build).To(BeTrue())
			Expect(result.BOM.Entries[1].Name).To(Equal("native-image-svm"))
			Expect(result.BOM.Entries[1].Launch).To(BeTrue())
			Expect(result.BOM.Entries[1].Build).To(BeTrue())
		})
	})

	context("native image enabled for API 0.7+ (not bundled with JDK) - custom command missing", func() {
		it("contributes native image dependency", func() {
			ctx.Plan.Entries = append(ctx.Plan.Entries,
				libcnb.BuildpackPlanEntry{
					Name: "jdk",
				},
				libcnb.BuildpackPlanEntry{
					Name: "native-image-builder",
				},
			)
			ctx.Buildpack.Metadata = map[string]interface{}{
				"dependencies": []map[string]interface{}{
					{
						"id":      "jdk",
						"version": "1.1.1",
						"stacks":  []interface{}{"test-stack-id"},
						"cpes":    []string{"cpe:2.3:a:oracle:graalvm:21.2.0:*:*:*:community:*:*:*"},
						"purl":    "pkg:generic/graalvm-jdk@21.2.0",
					},
					{
						"id":      "native-image-svm",
						"version": "2.2.2",
						"stacks":  []interface{}{"test-stack-id"},
						"cpes":    []string{"cpe:2.3:a:oracle:graalvm:21.2.0:*:*:*:community:*:*:*"},
						"purl":    "pkg:generic/graalvm-svm@21.2.0",
					},
				},
			}
			ctx.StackID = "test-stack-id"
			ctx.Buildpack.API = "0.7"

			_, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionMissingCommand).Build(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to create NIK, custom command has not been supplied by buildpack"))
		})
	})

	it("contributes NIK alternative buildplan (NIK bundled with JDK)", func() {
		// NIK includes a JDK, so we don't need a second JDK
		ctx.Plan.Entries = append(
			ctx.Plan.Entries,
			libcnb.BuildpackPlanEntry{Name: "native-image-builder"},
			libcnb.BuildpackPlanEntry{Name: "jdk", Metadata: map[string]interface{}{}},
			libcnb.BuildpackPlanEntry{Name: "jre", Metadata: map[string]interface{}{}})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "native-image-svm",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.Buildpack.API = "0.6"
		ctx.StackID = "test-stack-id"

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionBundledWithJDK).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name()).To(Equal("native-image-svm"))

		Expect(result.BOM.Entries).To(HaveLen(1))
		Expect(result.BOM.Entries[0].Name).To(Equal("native-image-svm"))
		Expect(result.BOM.Entries[0].Launch).To(BeFalse())
		Expect(result.BOM.Entries[0].Build).To(BeTrue())
	})

	context("$BP_JVM_VERSION", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_JVM_VERSION", "1.1.1")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_JVM_VERSION")).To(Succeed())
		})

		it("selects versions based on BP_JVM_VERSION", func() {
			ctx.Plan.Entries = append(ctx.Plan.Entries,
				libcnb.BuildpackPlanEntry{Name: "jdk"},
				libcnb.BuildpackPlanEntry{Name: "jre"},
			)
			ctx.Buildpack.Metadata = map[string]interface{}{
				"dependencies": []map[string]interface{}{
					{
						"id":      "jdk",
						"version": "1.1.1",
						"stacks":  []interface{}{"test-stack-id"},
					},
					{
						"id":      "jdk",
						"version": "2.2.2",
						"stacks":  []interface{}{"test-stack-id"},
					},
					{
						"id":      "jre",
						"version": "1.1.1",
						"stacks":  []interface{}{"test-stack-id"},
					},
					{
						"id":      "jre",
						"version": "2.2.2",
						"stacks":  []interface{}{"test-stack-id"},
					},
				},
			}
			ctx.StackID = "test-stack-id"

			result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers[0].(libjvm.JDK).LayerContributor.Dependency.Version).To(Equal("1.1.1"))
			Expect(result.Layers[1].(libjvm.JRE).LayerContributor.Dependency.Version).To(Equal("1.1.1"))
		})
	})

	context("$BP_JVM_TYPE", func() {

		it.After(func() {
			Expect(os.Unsetenv("BP_JVM_TYPE")).To(Succeed())
		})

		it("contributes JDK when specified explicitly in $BP_JVM_TYPE", func() {
			Expect(os.Setenv("BP_JVM_TYPE", "jdk")).To(Succeed())

			ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jdk", Metadata: LaunchContribution})
			ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
			ctx.Buildpack.Metadata = map[string]interface{}{
				"dependencies": []map[string]interface{}{
					{
						"id":      "jdk",
						"version": "0.0.2",
						"stacks":  []interface{}{"test-stack-id"},
					},
					{
						"id":      "jre",
						"version": "2.2.2",
						"stacks":  []interface{}{"test-stack-id"},
					},
				},
			}
			ctx.StackID = "test-stack-id"

			result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Layers[0].Name()).To(Equal("jdk"))
			Expect(result.Layers[0].(libjvm.JRE).LayerContributor.Dependency.ID).To(Equal("jdk"))

			Expect(result.BOM.Entries[0].Name).To(Equal("jdk"))
			Expect(result.BOM.Entries[0].Launch).To(BeTrue())
		})

		it("contributes JRE when specified explicitly in $BP_JVM_TYPE", func() {
			Expect(os.Setenv("BP_JVM_TYPE", "jre")).To(Succeed())

			ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jdk", Metadata: LaunchContribution})
			ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "jre", Metadata: LaunchContribution})
			ctx.Buildpack.Metadata = map[string]interface{}{
				"dependencies": []map[string]interface{}{
					{
						"id":      "jdk",
						"version": "0.0.1",
						"stacks":  []interface{}{"test-stack-id"},
					},
					{
						"id":      "jre",
						"version": "1.1.1",
						"stacks":  []interface{}{"test-stack-id"},
					},
				},
			}
			ctx.StackID = "test-stack-id"

			result, err := libjvm.NewBuild(bard.NewLogger(io.Discard)).Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Layers[0].Name()).To(Equal("jdk"))
			Expect(result.Layers[0].(libjvm.JDK).LayerContributor.Dependency.ID).To(Equal("jdk"))
			Expect(result.Layers[1].(libjvm.JRE).LayerContributor.Dependency.ID).To(Equal("jre"))

			Expect(result.BOM.Entries).To(HaveLen(3))
			Expect(result.BOM.Entries[0].Name).To(Equal("jdk"))
			Expect(result.BOM.Entries[0].Launch).To(BeFalse())
			Expect(result.BOM.Entries[0].Build).To(BeTrue())

			Expect(result.BOM.Entries[1].Name).To(Equal("jre"))
			Expect(result.BOM.Entries[1].Launch).To(BeTrue())
			Expect(result.BOM.Entries[1].Build).To(BeTrue())

		})
	})
}
