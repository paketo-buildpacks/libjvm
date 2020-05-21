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
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/pavel-v-chernykh/keystore-go"
	"github.com/sclevine/spec"
)

func testCertificateLoader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

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

	it("short circuits if no CA certificates file does not exist", func() {
		c := libjvm.CertificateLoader{
			CACertificatesPath: filepath.Join("testdata", "non-existent-file"),
			KeyStorePath:       path,
			KeyStorePassword:   []byte("changeit"),
			Logger:             ioutil.Discard,
		}

		Expect(c.Load()).To(Succeed())
	})

	it("loads additional certificates", func() {
		c := libjvm.CertificateLoader{
			CACertificatesPath: filepath.Join("testdata", "test-certificates.crt"),
			KeyStorePath:       path,
			KeyStorePassword:   []byte("changeit"),
			Logger:             ioutil.Discard,
		}

		Expect(c.Load()).To(Succeed())

		in, err := os.Open(path)
		Expect(err).NotTo(HaveOccurred())
		defer in.Close()

		ks, err := keystore.Decode(in, []byte("changeit"))
		Expect(ks).To(HaveLen(2))
	})

}
