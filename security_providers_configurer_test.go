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

func testSecurityProvidersConfigurer(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Buildpack.Path, err = ioutil.TempDir("", "security-providers-configurer-buildpack")
		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = ioutil.TempDir("", "security-providers-configurer-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	context("Java 8", func() {
		it("contributes Security Providers Configurer", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "security-providers-configurer"), []byte{}, 0755)).To(Succeed())

			l := libjvm.NewSecurityProvidersConfigurer(ctx.Buildpack, "8.0.212", &libcnb.BuildpackPlan{})
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = l.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Launch).To(BeTrue())
			Expect(filepath.Join(layer.Path, "bin", "security-providers-configurer")).To(BeARegularFile())
			Expect(layer.Profile["security-providers-classpath"]).To(Equal(`[[ -z "${SECURITY_PROVIDERS_CLASSPATH+x}" ]] && return

EXT_DIRS="${JAVA_HOME}/lib/ext"

for I in ${SECURITY_PROVIDERS_CLASSPATH//:/$'\n'}; do
  EXT_DIRS="${EXT_DIRS}:$(dirname "${I}")"
done

export JAVA_OPTS="${JAVA_OPTS} -Djava.ext.dirs=${EXT_DIRS}"`))
			Expect(layer.Profile["security-providers-configurer"]).To(Equal(
				`security-providers-configurer --source "${JAVA_HOME}"/lib/security/java.security --additional-providers "$(echo "${SECURITY_PROVIDERS}" | tr ' ' ,)"`))
		})
	})

	context("Java 11", func() {
		it("contributes Security Providers Configurer", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "security-providers-configurer"), []byte{}, 0755)).To(Succeed())

			l := libjvm.NewSecurityProvidersConfigurer(ctx.Buildpack, "11.0.3", &libcnb.BuildpackPlan{})
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = l.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Launch).To(BeTrue())
			Expect(filepath.Join(layer.Path, "bin", "security-providers-configurer")).To(BeARegularFile())
			Expect(layer.Profile["security-providers-classpath"]).To(Equal(`[[ -z "${SECURITY_PROVIDERS_CLASSPATH+x}" ]] && return

export CLASSPATH="${CLASSPATH}:${SECURITY_PROVIDERS_CLASSPATH}"`))
			Expect(layer.Profile["security-providers-configurer"]).To(Equal(
				`security-providers-configurer --source "${JAVA_HOME}"/conf/security/java.security --additional-providers "$(echo "${SECURITY_PROVIDERS}" | tr ' ' ,)"`))
		})
	})

}
