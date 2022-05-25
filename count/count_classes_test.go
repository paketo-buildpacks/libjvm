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

package count_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm/count"
)

func testCountClasses(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		var err error

		path, err = ioutil.TempDir("", "class-counter")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("counts files on filesystem", func() {
		Expect(ioutil.WriteFile(filepath.Join(path, "alpha.class"), []byte{}, 0644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(path, "bravo"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(path, "bravo", "charlie.class"), []byte{}, 0644)).To(Succeed())

		Expect(count.Classes(path)).To(Equal(2))
	})

	it("counts files in single archive", func() {
		Expect(count.JarClasses("testdata/stub-dependency.jar")).To(Equal(2))
	})

	it("counts files with nested archives 1 level down", func() {
		Expect(count.Classes("testdata/nested")).To(Equal(4))
	})

	it("counts files including any nested 1 level down", func() {
		Expect(count.Classes("testdata")).To(Equal(6))
	})

	it("skips empty zip/jar files with none in the name", func() {
		Expect(ioutil.WriteFile(filepath.Join(path, "test-none.jar"), []byte{}, 0644)).To(Succeed())

		Expect(count.Classes(path)).To(Equal(0))
	})

	it("skips directories with .jar suffix", func() {
		Expect(os.MkdirAll(filepath.Join(path, "bad-dir.jar"), 0755)).To(Succeed())

		Expect(count.Classes(path)).To(Equal(0))
	})

	it("skips bad jar files", func() {
		Expect(ioutil.WriteFile(filepath.Join(path, "bad-jar.jar"), []byte{}, 0755)).To(Succeed())

		Expect(count.Classes(path)).To(Equal(0))
	})
}
