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
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
)

func testOpenSSLCertificateLoader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "openssl-security-provider-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	context("Java 8", func() {
		it("contributes OpenSSL Certificate Loader", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "openssl-certificate-loader"), []byte{}, 0755)).To(Succeed())

			l := libjvm.NewOpenSSLCertificateLoader(ctx.Buildpack, libjvm.JREType, "8.0.212", &libcnb.BuildpackPlan{})
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = l.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Launch).To(BeTrue())
			Expect(filepath.Join(layer.Path, "bin", "openssl-certificate-loader")).To(BeARegularFile())
		})

		it("contributes JDK profiles", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "openssl-certificate-loader"), []byte{}, 0755)).To(Succeed())

			l := libjvm.NewOpenSSLCertificateLoader(ctx.Buildpack, libjvm.JDKType, "8.0.212", &libcnb.BuildpackPlan{})
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = l.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Profile["openssl-certificate-loader.sh"]).To(Equal(`openssl-certificate-loader \
  --ca-certificates="/etc/ssl/certs/ca-certificates.crt" \
  --keystore-path="${JAVA_HOME}/jre/lib/security/cacerts" \
  --keystore-password="changeit"
`))
		})

		it("contributes JRE profiles", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "openssl-certificate-loader"), []byte{}, 0755)).To(Succeed())

			l := libjvm.NewOpenSSLCertificateLoader(ctx.Buildpack, libjvm.JREType, "8.0.212", &libcnb.BuildpackPlan{})
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = l.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Profile["openssl-certificate-loader.sh"]).To(Equal(`openssl-certificate-loader \
  --ca-certificates="/etc/ssl/certs/ca-certificates.crt" \
  --keystore-path="${JAVA_HOME}/lib/security/cacerts" \
  --keystore-password="changeit"
`))
		})
	})

	context("Java 11", func() {
		it("contributes OpenSSL Certificate Loader", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Buildpack.Path, "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Buildpack.Path, "bin", "openssl-certificate-loader"), []byte{}, 0755)).To(Succeed())

			l := libjvm.NewOpenSSLCertificateLoader(ctx.Buildpack, libjvm.JREType, "11.0.3", &libcnb.BuildpackPlan{})
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = l.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Launch).To(BeTrue())
			Expect(filepath.Join(layer.Path, "bin", "openssl-certificate-loader")).To(BeARegularFile())
			Expect(layer.Profile["openssl-certificate-loader.sh"]).To(Equal(`openssl-certificate-loader \
  --ca-certificates="/etc/ssl/certs/ca-certificates.crt" \
  --keystore-path="${JAVA_HOME}/lib/security/cacerts" \
  --keystore-password="changeit"
`))
		})
	})

}
