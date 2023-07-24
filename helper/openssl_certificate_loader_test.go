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

package helper_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb/log"

	"github.com/paketo-buildpacks/libjvm"
	"github.com/paketo-buildpacks/libjvm/helper"
	"github.com/paketo-buildpacks/libjvm/internal"

	"github.com/paketo-buildpacks/libpak/bard"
)

func testOpenSSLCertificateLoader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cl = libjvm.CertificateLoader{
			CertDirs: []string{filepath.Join("testdata", "certificates")},
			Logger:   io.Discard,
		}

		path string
	)

	it.Before(func() {
		in, err := os.Open(filepath.Join("testdata", "test-keystore.jks"))
		Expect(err).NotTo(HaveOccurred())
		defer in.Close()

		out, err := ioutil.TempFile("", "certificate-loader")
		Expect(err).NotTo(HaveOccurred())
		defer out.Close()

		_, err = io.Copy(out, in)
		Expect(err).NotTo(HaveOccurred())

		path = out.Name()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("returns error if BPI_JVM_CACERTS is not set", func() {
		o := helper.OpenSSLCertificateLoader{CertificateLoader: cl, Logger: bard.NewLogger(ioutil.Discard)}
		o.Logger.Logger = log.NewDiscard()

		_, err := o.Execute()

		Expect(err).To(MatchError("$BPI_JVM_CACERTS must be set"))
	})

	context("$BPI_JVM_CACERTS", func() {

		it.Before(func() {
			Expect(os.Setenv("BPI_JVM_CACERTS", path)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BPI_JVM_CACERTS")).To(Succeed())
			_ = os.Remove(helper.TmpTrustStore)
		})

		it("loads additional certificates", func() {
			o := helper.OpenSSLCertificateLoader{CertificateLoader: cl, Logger: bard.NewLogger(ioutil.Discard)}
			o.Logger.Logger = log.NewDiscard()

			Expect(o.Execute()).To(BeNil())

			in, err := os.Open(path)
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			ks := keystore.New()
			err = ks.Load(in, []byte("changeit"))
			Expect(err).NotTo(HaveOccurred())
			Expect(ks.Aliases()).To(HaveLen(3))
		})

		if internal.IsRoot() {
			return
		}

		it("does use temp keystore if keystore is read-only", func() {
			Expect(os.Chmod(path, 0555)).To(Succeed())

			o := helper.OpenSSLCertificateLoader{CertificateLoader: cl, Logger: bard.NewLogger(ioutil.Discard)}
			o.Logger.Logger = log.NewDiscard()

			env, err := o.Execute()
			Expect(err).NotTo(HaveOccurred())

			in, err := os.Open(helper.TmpTrustStore)
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			ks := keystore.New()
			err = ks.Load(in, []byte("changeit"))
			Expect(err).NotTo(HaveOccurred())
			Expect(ks.Aliases()).To(HaveLen(1))

			Expect(env).To(HaveKeyWithValue("JAVA_TOOL_OPTIONS", fmt.Sprintf("-Djavax.net.ssl.trustStore=%s", helper.TmpTrustStore)))
		})

		it("does not return error when keystore and /tmp/truststore are read-only", func() {
			Expect(os.Chmod(path, 0555)).To(Succeed())
			_, err := os.OpenFile(helper.TmpTrustStore, os.O_CREATE, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chmod(helper.TmpTrustStore, 0555)).To(Succeed())

			o := helper.OpenSSLCertificateLoader{CertificateLoader: cl, Logger: bard.NewLogger(os.Stdout)}
			o.Logger.Logger = log.NewDiscard()

			env, err := o.Execute()
			Expect(env).To(BeNil())
			Expect(err).NotTo(HaveOccurred())

			in, err := os.Open(path)
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			ks := keystore.New()
			err = ks.Load(in, []byte("changeit"))
			Expect(err).NotTo(HaveOccurred())
			Expect(ks.Aliases()).To(HaveLen(1))
		})

	})

}
