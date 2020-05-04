/*
 * Copyright 2018-2020, VMware, Inc. All Rights Reserved.
 * Proprietary and Confidential.
 * Unauthorized use, copying or distribution of this source code via any medium is
 * strictly prohibited without the express written consent of VMware, Inc.
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

		j := libjvm.NewOpenSSLSecurityProvider(dep, dc, NoContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(filepath.Join(layer.Path, "stub-openssl-security-provider.jar")).To(BeARegularFile())
		Expect(layer.SharedEnvironment["SECURITY_PROVIDERS.append"]).To(Equal(" 2|io.paketo.openssl.OpenSslProvider"))
		Expect(layer.SharedEnvironment["SECURITY_PROVIDERS_CLASSPATH"]).To(Equal(filepath.Join(layer.Path, "stub-openssl-security-provider.jar")))
		Expect(layer.Profile["openssl-security-provider.sh"]).To(Equal(`if [[ -f /etc/ssl/certs/ca-certificates.crt ]]; then
  export JAVA_OPTS="${JAVA_OPTS} -Dio.paketo.openssl.ca-certificates=/etc/ssl/certs/ca-certificates.crt"
fi
`))
	})

	it("marks layer for build", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-openssl-security-provider.jar",
			SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j := libjvm.NewOpenSSLSecurityProvider(dep, dc, BuildContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Build).To(BeTrue())
		Expect(layer.Cache).To(BeTrue())
	})

	it("marks layer for launch", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-openssl-security-provider.jar",
			SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j := libjvm.NewOpenSSLSecurityProvider(dep, dc, LaunchContribution, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
	})

}
