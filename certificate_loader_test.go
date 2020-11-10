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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pavel-v-chernykh/keystore-go/v4"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
	"github.com/paketo-buildpacks/libjvm/internal"
)

func testCertificateLoader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("certificate sources", func() {
		it("returns default sources", func() {
			c := libjvm.NewCertificateLoader()

			Expect(c.CertFile).To(Equal(libjvm.DefaultCertFile))
			Expect(c.CertDirs).To(BeNil())
		})

		context("$SSL_CERT_DIR", func() {
			it.Before(func() {
				Expect(os.Setenv("SSL_CERT_FILE", "another-file")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("SSL_CERT_FILE")).To(Succeed())
			})

			it("returns configured file", func() {
				c := libjvm.NewCertificateLoader()

				Expect(c.CertFile).To(Equal("another-file"))
				Expect(c.CertDirs).To(BeNil())
			})
		})

		context("$SSL_CERT_DIR", func() {
			it.Before(func() {
				Expect(os.Setenv("SSL_CERT_DIR",
					strings.Join([]string{"test-1", "test-2"}, string(filepath.ListSeparator)))).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("SSL_CERT_DIR")).To(Succeed())
			})

			it("returns configured directories", func() {
				c := libjvm.NewCertificateLoader()

				Expect(c.CertFile).To(Equal(libjvm.DefaultCertFile))
				Expect(c.CertDirs).To(Equal([]string{"test-1", "test-2"}))
			})
		})
	})

	context("load", func() {
		var (
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

		it("ignores non-existent file", func() {
			c := libjvm.CertificateLoader{
				CertFile: filepath.Join("testdata", "non-existent-file"),
				Logger:   ioutil.Discard,
			}

			Expect(c.Load(path, "changeit")).To(Succeed())
		})

		it("ignores non-existent directory", func() {
			c := libjvm.CertificateLoader{
				CertDirs: []string{filepath.Join("testdata", "non-existent-directory")},
				Logger:   ioutil.Discard,
			}

			Expect(c.Load(path, "changeit")).To(Succeed())
		})

		it("loads additional certificates from file", func() {
			c := libjvm.CertificateLoader{
				CertFile: filepath.Join("testdata", "certificates", "certificate-1.pem"),
				Logger:   ioutil.Discard,
			}

			Expect(c.Load(path, "changeit")).To(Succeed())

			in, err := os.Open(path)
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			ks := keystore.New()
			err = ks.Load(in, []byte("changeit"))
			Expect(err).NotTo(HaveOccurred())
			Expect(ks.Aliases()).To(HaveLen(2))
		})

		it("loads additional certificates from directories", func() {
			c := libjvm.CertificateLoader{
				CertDirs: []string{filepath.Join("testdata", "certificates")},
				Logger:   ioutil.Discard,
			}

			Expect(c.Load(path, "changeit")).To(Succeed())

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

		it("does not return error when keystore is read-only", func() {
			Expect(os.Chmod(path, 0555)).To(Succeed())

			c := libjvm.CertificateLoader{
				CertDirs: []string{filepath.Join("testdata", "certificates")},
				Logger:   ioutil.Discard,
			}

			Expect(c.Load(path, "changeit")).To(Succeed())

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
