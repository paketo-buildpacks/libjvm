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
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"
)

func testOpenSSLSecurityProvider(t *testing.T, context spec.G, it spec.S) {
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

	it("contributes OpenSSLSecurityProvider", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-openssl-security-provider.jar",
			SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j := libjvm.NewOpenSSLSecurityProvider(dep, dc, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(filepath.Join(layer.Path, "stub-openssl-security-provider.jar")).To(BeARegularFile())
		Expect(layer.LaunchEnvironment["SECURITY_PROVIDERS.append"]).To(Equal(" 2|io.paketo.openssl.OpenSslProvider"))
		Expect(layer.LaunchEnvironment["SECURITY_PROVIDERS_CLASSPATH"]).To(Equal(filepath.Join(layer.Path, "stub-openssl-security-provider.jar")))
		Expect(layer.Profile["openssl-security-provider.sh"]).To(Equal(`if [[ -f /etc/ssl/certs/ca-certificates.crt ]]; then
  export JAVA_OPTS="${JAVA_OPTS} -Dio.paketo.openssl.ca-certificates=/etc/ssl/certs/ca-certificates.crt"
fi
`))
	})

}
