/*
 * Copyright 2018-2022 the original author or authors.
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

	"github.com/anthonydahanne/libjvm"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testSDKMAN(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		var err error
		path, err = ioutil.TempDir("", "sdkman")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("parses single entry sdkmanrc file", func() {
		sdkmanrcFile := filepath.Join(path, "sdkmanrc")
		Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=17.0.2-tem`), 0644)).To(Succeed())

		res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal([]libjvm.SDKInfo{
			{Type: "java", Version: "17.0.2", Vendor: "tem"},
		}))
	})

	it("parses single entry sdkmanrc file and forces lowercase", func() {
		sdkmanrcFile := filepath.Join(path, "sdkmanrc")
		Expect(ioutil.WriteFile(sdkmanrcFile, []byte(` jAva = 17.0.2-TEM `), 0644)).To(Succeed())

		res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal([]libjvm.SDKInfo{
			{Type: "java", Version: "17.0.2", Vendor: "tem"},
		}))
	})

	it("parses multiple entry sdkmanrc and doesn't care if there's overlap", func() {
		sdkmanrcFile := filepath.Join(path, "sdkmanrc")
		Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=11.0.2-tem
java=17.0.2-tem`), 0644)).To(Succeed())

		res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal([]libjvm.SDKInfo{
			{Type: "java", Version: "11.0.2", Vendor: "tem"},
			{Type: "java", Version: "17.0.2", Vendor: "tem"},
		}))
	})

	context("handles comments and whitespace", func() {
		it("ignores full-line comments", func() {
			sdkmanrcFile := filepath.Join(path, "sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`# Enable auto-env through the sdkman_auto_env config
# Add key=value pairs of SDKs to use below
java=17.0.2-tem
    # has some leading whitespace`), 0644)).To(Succeed())

			res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]libjvm.SDKInfo{
				{Type: "java", Version: "17.0.2", Vendor: "tem"},
			}))
		})

		it("ignores trailing-line comments", func() {
			sdkmanrcFile := filepath.Join(path, "sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=17.0.2-tem # comment`), 0644)).To(Succeed())

			res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]libjvm.SDKInfo{
				{Type: "java", Version: "17.0.2", Vendor: "tem"},
			}))
		})

		it("ignores empty lines", func() {
			sdkmanrcFile := filepath.Join(path, "sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`
# Enable auto-env through the sdkman_auto_env config
              
java=17.0.2-tem

`), 0644)).To(Succeed())

			res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]libjvm.SDKInfo{
				{Type: "java", Version: "17.0.2", Vendor: "tem"},
			}))
		})
	})

	context("handles malformed key/values", func() {
		it("parses an empty value", func() {
			sdkmanrcFile := filepath.Join(path, "sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`java=`), 0644)).To(Succeed())

			res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]libjvm.SDKInfo{
				{Type: "java", Version: "", Vendor: ""},
			}))
		})

		it("parses with an empty key", func() {
			sdkmanrcFile := filepath.Join(path, "sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`=foo-vend`), 0644)).To(Succeed())

			res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]libjvm.SDKInfo{
				{Type: "", Version: "foo", Vendor: "vend"},
			}))
		})

		it("parses with an empty vendor", func() {
			sdkmanrcFile := filepath.Join(path, "sdkmanrc")
			Expect(ioutil.WriteFile(sdkmanrcFile, []byte(`foo=bar`), 0644)).To(Succeed())

			res, err := libjvm.ReadSDKMANRC(sdkmanrcFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]libjvm.SDKInfo{
				{Type: "foo", Version: "bar", Vendor: ""},
			}))
		})
	})
}
