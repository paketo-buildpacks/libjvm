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
	"io"
	"os"
	"testing"

	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
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

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)

		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name).To(Equal("jdk"))
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

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(3))
		Expect(result.Layers[0].Name).To(Equal("jre"))
		Expect(result.Layers[1].Name).To(Equal("helper"))
		Expect(result.Layers[2].Name).To(Equal("java-security-properties"))
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

		var names []string
		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if name == "helper" {
				names = creator.(libpak.HelperLayerContributor).Names
			}
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[0].Name).To(Equal("jre"))
		Expect(result.Layers[1].Name).To(Equal("helper"))

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
		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if name == "helper" {
				names = creator.(libpak.HelperLayerContributor).Names
			}
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[0].Name).To(Equal("jre"))

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

		var id string
		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if name == "jdk" {
				id = creator.(libjvm.JRE).LayerContributor.Dependency.ID
			}
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[0].Name).To(Equal("jdk"))
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

		var id string
		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if name == "jdk" {
				id = creator.(libjvm.JRE).LayerContributor.Dependency.ID
			}
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[0].Name).To(Equal("jdk"))
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

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionBundledWithJDK, libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name).To(Equal("native-image-svm"))
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

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionBundledWithJDK, libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name).To(Equal("native-image-svm"))
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
			result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionSeparateFromJDK, libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
				name := creator.Name()
				layer, err := ctx.Layers.Layer(name)
				if name == "nik" {
					dep = creator.(libjvm.NIK).NativeDependency
				}
				if err != nil {
					return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
				}
				return layer, nil
			})).Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			Expect(result.Layers[0].Name).To(Equal("nik"))
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

			_, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionMissingCommand, libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
				name := creator.Name()
				layer, err := ctx.Layers.Layer(name)
				if err != nil {
					return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
				}
				return layer, nil
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

		result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), nativeOptionBundledWithJDK, libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
			name := creator.Name()
			layer, err := ctx.Layers.Layer(name)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
			}
			return layer, nil
		})).Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name).To(Equal("native-image-svm"))
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

			var jdkver string
			var jrever string
			result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
				name := creator.Name()
				layer, err := ctx.Layers.Layer(name)
				if name == "jdk" {
					jdkver = creator.(libjvm.JDK).LayerContributor.Dependency.Version
				}
				if name == "jre" {
					jrever = creator.(libjvm.JRE).LayerContributor.Dependency.Version
				}
				if err != nil {
					return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
				}
				return layer, nil
			})).Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers[0].Name).To(Equal("jdk"))
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

			var jdkid string
			result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
				name := creator.Name()
				layer, err := ctx.Layers.Layer(name)
				if name == "jdk" {
					jdkid = creator.(libjvm.JRE).LayerContributor.Dependency.ID
				}
				if err != nil {
					return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
				}
				return layer, nil
			})).Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Layers[0].Name).To(Equal("jdk"))
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

			var jdkid string
			var jreid string
			result, err := libjvm.NewBuild(bard.NewLogger(io.Discard), libjvm.WithCustomFlattenContributorFn(func(creator libjvm.LayerContributor, ctx libcnb.BuildContext) (libcnb.Layer, error) {
				name := creator.Name()
				layer, err := ctx.Layers.Layer(name)
				if name == "jdk" {
					jdkid = creator.(libjvm.JDK).LayerContributor.Dependency.ID
				}
				if name == "jre" {
					jreid = creator.(libjvm.JRE).LayerContributor.Dependency.ID
				}
				if err != nil {
					return libcnb.Layer{}, fmt.Errorf("unable to create layer %s\n%w", name, err)
				}
				return layer, nil
			})).Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Layers[0].Name).To(Equal("jdk"))
			Expect(jdkid).To(Equal("jdk"))
			Expect(jreid).To(Equal("jre"))

		})
	})
}
