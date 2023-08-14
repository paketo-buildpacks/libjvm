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
	"reflect"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/libpak/v2/bard"
	"github.com/paketo-buildpacks/libpak/v2/crush"
	"github.com/paketo-buildpacks/libpak/v2/effect"
	"github.com/paketo-buildpacks/libpak/v2/effect/mocks"

	"github.com/stretchr/testify/mock"

	"github.com/buildpacks/libcnb/v2"
	. "github.com/onsi/gomega"

	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
)

func testJLink(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cl = libjvm.CertificateLoader{
			CertDirs: []string{filepath.Join("testdata", "certificates")},
			Logger:   io.Discard,
		}

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		Expect(os.Setenv("BP_JVM_JLINK_ENABLED", "true")).To(Succeed())
		ctx.Layers.Path = t.TempDir()
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
		Expect(os.Unsetenv("BP_JVM_JLINK_ENABLED")).To(Succeed())

	})

	it("contributes jlink JRE with default args", func() {

		args := []string{"--no-man-pages", "--no-header-files", "--strip-debug"}
		exec := &mocks.Executor{}
		j, err := libjvm.NewJLink(ctx.ApplicationPath, exec, args, cl, LaunchContribution, false)
		Expect(err).NotTo(HaveOccurred())
		j.Logger = bard.NewLogger(io.Discard)

		Expect(j.LayerContributor.ExpectedMetadata.(map[string]interface{})["cert-dir"]).To(HaveLen(4))

		layer, err := ctx.Layers.Layer("jlink")
		Expect(err).NotTo(HaveOccurred())

		exec.On("Execute", mock.MatchedBy(func(ex effect.Execution) bool {
			return reflect.DeepEqual(ex.Args, []string{"--list-modules"})
		})).Return(func(ex effect.Execution) error {
			_, err := ex.Stdout.Write([]byte("java.se,java.base"))
			Expect(err).ToNot(HaveOccurred())
			return nil
		})

		exec.On("Execute", mock.MatchedBy(func(ex effect.Execution) bool {
			return strings.Contains(ex.Command, "jlink")
		})).Run(func(args mock.Arguments) {
			err = os.MkdirAll(filepath.Join(ctx.Layers.Path, "jlink"), os.ModePerm)
			jre, err := os.Open("testdata/3aa01010c0d3592ea248c8353d60b361231fa9bf9a7479b4f06451fef3e64524/stub-jre-11.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			err = crush.Extract(jre, layer.Path, 1)
			Expect(err).NotTo(HaveOccurred())
		}).Return(nil)

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		e := exec.Calls[1].Arguments[0].(effect.Execution)
		Expect(e.Args).To(ContainElement("--add-modules"))
		Expect(e.Args).To(ContainElement("java.se,java.base"))
		Expect(e.Args).To(ContainElement("--output"))
	})

	it("contributes jlink JRE with user provided args & modules", func() {

		args := []string{"--no-man-pages", "--no-header-files", "--strip-debug", "--add-modules", "java.se"}
		exec := &mocks.Executor{}
		j, err := libjvm.NewJLink(ctx.ApplicationPath, exec, args, cl, LaunchContribution, true)
		Expect(err).NotTo(HaveOccurred())
		j.Logger = bard.NewLogger(io.Discard)

		Expect(j.LayerContributor.ExpectedMetadata.(map[string]interface{})["cert-dir"]).To(HaveLen(4))

		layer, err := ctx.Layers.Layer("jlink")
		Expect(err).NotTo(HaveOccurred())

		exec.On("Execute", mock.Anything).Run(func(args mock.Arguments) {
			err = os.MkdirAll(filepath.Join(ctx.Layers.Path, "jlink"), os.ModePerm)
			jre, err := os.Open("testdata/3aa01010c0d3592ea248c8353d60b361231fa9bf9a7479b4f06451fef3e64524/stub-jre-11.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			err = crush.Extract(jre, layer.Path, 1)
			Expect(err).NotTo(HaveOccurred())
		}).Return(nil)

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		e := exec.Calls[0].Arguments[0].(effect.Execution)
		Expect(e.Args).To(ContainElement("--output"))
		Expect(e.Args).To(ContainElement("--add-modules"))
		Expect(e.Args).To(ContainElement("java.se"))
	})

	it("contributes jlink JRE with user provided all caps --add-modules argument", func() {

		args := []string{"--no-man-pages", "--no-header-files", "--strip-debug", "--add-modules", "ALL-MODULE-PATH"}
		exec := &mocks.Executor{}
		j, err := libjvm.NewJLink(ctx.ApplicationPath, exec, args, cl, LaunchContribution, true)
		Expect(err).NotTo(HaveOccurred())
		j.Logger = bard.NewLogger(io.Discard)

		Expect(j.LayerContributor.ExpectedMetadata.(map[string]interface{})["cert-dir"]).To(HaveLen(4))

		layer, err := ctx.Layers.Layer("jlink")
		Expect(err).NotTo(HaveOccurred())

		exec.On("Execute", mock.Anything).Run(func(args mock.Arguments) {
			err = os.MkdirAll(filepath.Join(ctx.Layers.Path, "jlink"), os.ModePerm)
			jre, err := os.Open("testdata/3aa01010c0d3592ea248c8353d60b361231fa9bf9a7479b4f06451fef3e64524/stub-jre-11.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			err = crush.Extract(jre, layer.Path, 1)
			Expect(err).NotTo(HaveOccurred())
		}).Return(nil)

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		e := exec.Calls[0].Arguments[0].(effect.Execution)
		Expect(e.Args).To(ContainElement("--output"))
		Expect(e.Args).To(ContainElement("--no-man-pages"))
		Expect(e.Args).To(ContainElement("--no-header-files"))
		Expect(e.Args).To(ContainElement("--strip-debug"))
		Expect(e.Args).To(ContainElement("--add-modules"))
		Expect(e.Args).To(ContainElement("ALL-MODULE-PATH"))
	})

	it("contributes jlink JRE with user provided args & missing modules", func() {

		args := []string{"--no-man-pages", "--no-header-files", "--strip-debug"}
		exec := &mocks.Executor{}
		j, err := libjvm.NewJLink(ctx.ApplicationPath, exec, args, cl, LaunchContribution, true)
		Expect(err).NotTo(HaveOccurred())
		j.Logger = bard.NewLogger(io.Discard)

		Expect(j.LayerContributor.ExpectedMetadata.(map[string]interface{})["cert-dir"]).To(HaveLen(4))

		layer, err := ctx.Layers.Layer("jlink")
		Expect(err).NotTo(HaveOccurred())

		exec.On("Execute", mock.MatchedBy(func(ex effect.Execution) bool {
			return strings.Contains(ex.Command, "java")
		})).Return(func(ex effect.Execution) error {
			_, err := ex.Stdout.Write([]byte("java.se,java.base"))
			Expect(err).ToNot(HaveOccurred())
			return nil
		})

		exec.On("Execute", mock.MatchedBy(func(ex effect.Execution) bool {
			return strings.Contains(ex.Command, "jlink")
		})).Run(func(args mock.Arguments) {
			err = os.MkdirAll(filepath.Join(ctx.Layers.Path, "jlink"), os.ModePerm)
			jre, err := os.Open("testdata/3aa01010c0d3592ea248c8353d60b361231fa9bf9a7479b4f06451fef3e64524/stub-jre-11.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			err = crush.Extract(jre, layer.Path, 1)
			Expect(err).NotTo(HaveOccurred())
		}).Return(nil)

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		e := exec.Calls[1].Arguments[0].(effect.Execution)
		Expect(e.Args).To(ContainElement("--add-modules"))
		Expect(e.Args).To(ContainElement("java.se,java.base"))
		Expect(e.Args).To(ContainElement("--output"))
	})

}
