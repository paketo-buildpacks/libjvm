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
	"encoding/pem"
	"io"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/sclevine/spec"
)

func testKeystore(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		path   string
	)

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("jks keystore", func() {
		it.Before(func() {
			in, err := os.Open(filepath.Join("testdata", "test-keystore.jks"))
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			out, err := os.CreateTemp("", "certificate-loader")
			Expect(err).NotTo(HaveOccurred())
			defer out.Close()

			_, err = io.Copy(out, in)
			Expect(err).NotTo(HaveOccurred())

			path = out.Name()
		})

		it("is detected correctly", func() {
			ks, err := libjvm.DetectKeystore(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(ks).To(BeAssignableToTypeOf(&libjvm.JKSKeystore{}))
		})

		it("can be written", func() {
			ks, err := libjvm.NewJKSKeystore(path, "changeit")
			Expect(err).ToNot(HaveOccurred())
			Expect(ks.Len()).To(Equal(1))
			cert, err := os.ReadFile(filepath.Join("testdata", "cert.pem"))
			Expect(err).ToNot(HaveOccurred())
			block, _ := pem.Decode(cert)
			ks.Add("foo", block)
			Expect(ks.Len()).To(Equal(2))
			err = ks.Write()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	context("pkcs12 keystore", func() {
		it.Before(func() {
			in, err := os.Open(filepath.Join("testdata", "test-keystore.pkcs12"))
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			out, err := os.CreateTemp("", "certificate-loader")
			Expect(err).NotTo(HaveOccurred())
			defer out.Close()

			_, err = io.Copy(out, in)
			Expect(err).NotTo(HaveOccurred())

			path = out.Name()
		})

		it("is detected correctly", func() {
			ks, err := libjvm.DetectKeystore(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(ks).To(BeAssignableToTypeOf(&libjvm.PasswordLessPKCS12Keystore{}))
		})

		it("can be written", func() {
			ks, err := libjvm.NewPasswordLessPKCS12Keystore(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(ks.Len()).To(Equal(1))
			cert, err := os.ReadFile(filepath.Join("testdata", "cert.pem"))
			Expect(err).ToNot(HaveOccurred())
			block, _ := pem.Decode(cert)
			err = ks.Add("foo", block)
			Expect(err).ToNot(HaveOccurred())
			Expect(ks.Len()).To(Equal(2))
			err = ks.Write()
			Expect(err).ToNot(HaveOccurred())
		})
	})
}
