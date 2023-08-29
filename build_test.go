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
	"testing"

	"github.com/paketo-buildpacks/libpak/v2"
	"github.com/paketo-buildpacks/libpak/v2/log"

	"github.com/buildpacks/libcnb/v2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm/v2"
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

		var layers []string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					layers = append(layers, name)
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers).To(HaveLen(1))
		Expect(layers[0]).To(Equal("jdk"))
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

		var layers []string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					layers = append(layers, name)
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers).To(HaveLen(3))
		Expect(layers[0]).To(Equal("jre"))
		Expect(layers[1]).To(Equal("helper"))
		Expect(layers[2]).To(Equal("java-security-properties"))
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

		var layers []string
		var names []string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					if name == "helper" {
						names = creator.(libpak.HelperLayerContributor).Names
					}
					layers = append(layers, name)
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers[0]).To(Equal("jre"))
		Expect(layers[1]).To(Equal("helper"))

		Expect(names).To(Equal([]string{
			"active-processor-count",
			"java-opts",
			"jvm-heap",
			"link-local-dns",
			"memory-calculator",
			"security-providers-configurer",
			"jmx",
			"jfr",
			"security-providers-classpath-8",
			"debug-8",
			"openssl-certificate-loader",
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

		var names []string
		var layers []string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					layers = append(layers, name)
					if name == "helper" {
						names = creator.(libpak.HelperLayerContributor).Names
					}
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers).To(HaveLen(3))
		Expect(layers[0]).To(Equal("jre"))
		Expect(layers[1]).To(Equal("helper"))

		Expect(names).To(Equal([]string{
			"active-processor-count",
			"java-opts",
			"jvm-heap",
			"link-local-dns",
			"memory-calculator",
			"security-providers-configurer",
			"jmx",
			"jfr",
			"security-providers-classpath-9",
			"debug-9",
			"nmt",
			"openssl-certificate-loader",
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

		var layers []string
		var id string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					layers = append(layers, name)
					if name == "jdk" {
						id = creator.(libjvm.JRE).LayerContributor.Dependency.ID
					}
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers).To(HaveLen(3))
		Expect(layers[0]).To(Equal("jdk"))
		Expect(id).To(Equal("jdk"))
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

		var layers []string
		var id string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					layers = append(layers, name)
					if name == "jdk" {
						id = creator.(libjvm.JRE).LayerContributor.Dependency.ID
					}
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers).To(HaveLen(3))
		Expect(layers[0]).To(Equal("jdk"))
		Expect(id).To(Equal("jdk"))
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

		var layers []string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), nativeOptionBundledWithJDK, libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					layers = append(layers, name)
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers).To(HaveLen(1))
		Expect(layers[0]).To(Equal("native-image-svm"))
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

		var layers []string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), nativeOptionBundledWithJDK, libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					layers = append(layers, name)
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers).To(HaveLen(1))
		Expect(layers[0]).To(Equal("native-image-svm"))
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

			var dep *libpak.BuildModuleDependency
			var layers []string
			_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), nativeOptionSeparateFromJDK, libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
				return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
					for _, creator := range layerContributors {
						name := creator.Name()
						if name == "nik" {
							dep = creator.(libjvm.NIK).NativeDependency
						}
						layers = append(layers, name)
					}
					return libcnb.NewBuildResult(), nil
				}
			})).Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(layers).To(HaveLen(1))
			Expect(layers[0]).To(Equal("nik"))
			Expect(dep).NotTo(BeNil())
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

			var layers []string
			_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), nativeOptionMissingCommand, libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
				return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
					for _, creator := range layerContributors {
						name := creator.Name()
						layers = append(layers, name)
					}
					return libcnb.NewBuildResult(), nil
				}
			})).Build(ctx)
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

		var layers []string
		_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), nativeOptionBundledWithJDK, libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
			return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				for _, creator := range layerContributors {
					name := creator.Name()
					layers = append(layers, name)
				}
				return libcnb.NewBuildResult(), nil
			}
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(layers).To(HaveLen(1))
		Expect(layers[0]).To(Equal("native-image-svm"))
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

			var layers []string
			var jdkver string
			var jrever string
			_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
				return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
					for _, creator := range layerContributors {
						name := creator.Name()
						if name == "jdk" {
							jdkver = creator.(libjvm.JDK).LayerContributor.Dependency.Version
						}
						if name == "jre" {
							jrever = creator.(libjvm.JRE).LayerContributor.Dependency.Version
						}
						layers = append(layers, name)
					}
					return libcnb.NewBuildResult(), nil
				}
			})).Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(layers).To(HaveLen(2))
			Expect(layers[0]).To(Equal("jdk"))
			Expect(jdkver).To(Equal("1.1.1"))
			Expect(jrever).To(Equal("1.1.1"))
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

			var layers []string
			var jdkid string
			_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
				return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
					for _, creator := range layerContributors {
						name := creator.Name()
						if name == "jdk" {
							jdkid = creator.(libjvm.JRE).LayerContributor.Dependency.ID
						}
						layers = append(layers, name)
					}
					return libcnb.NewBuildResult(), nil
				}
			})).Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(layers).To(HaveLen(3))
			Expect(layers[0]).To(Equal("jdk"))
			Expect(jdkid).To(Equal("jdk"))

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

			var layers []string
			var jdkid string
			var jreid string
			_, err := libjvm.NewBuild(log.NewPaketoLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(layerContributors []libpak.Contributable) libcnb.BuildFunc {
				return func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
					for _, creator := range layerContributors {
						name := creator.Name()
						if name == "jdk" {
							jdkid = creator.(libjvm.JDK).LayerContributor.Dependency.ID
						}
						if name == "jre" {
							jreid = creator.(libjvm.JRE).LayerContributor.Dependency.ID
						}
						layers = append(layers, name)
					}
					return libcnb.NewBuildResult(), nil
				}
			})).Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(layers).To(HaveLen(4))
			Expect(layers[0]).To(Equal("jdk"))
			Expect(jdkid).To(Equal("jdk"))
			Expect(jreid).To(Equal("jre"))

		})
	})
}
