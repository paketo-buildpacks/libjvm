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

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
)

func testManifest(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		var err error
		path, err = ioutil.TempDir("", "manifest")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("returns empty manifest if file doesn't exist", func() {
		m, err := libjvm.NewManifest(path)
		Expect(err).NotTo(HaveOccurred())

		Expect(m.Len()).To(Equal(0))
	})

	it("returns populated manifest if file exists", func() {
		Expect(os.MkdirAll(filepath.Join(path, "META-INF"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(path, "META-INF", "MANIFEST.MF"), []byte("test-key=test-value"), 0644)).To(Succeed())

		m, err := libjvm.NewManifest(path)
		Expect(err).NotTo(HaveOccurred())

		k, ok := m.Get("test-key")
		Expect(ok).To(BeTrue())
		Expect(k).To(Equal("test-value"))
	})

	it("returns proper values when lines are broken", func() {
		Expect(os.MkdirAll(filepath.Join(path, "META-INF"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(path, "META-INF", "MANIFEST.MF"), []byte(`
Manifest-Version: 1.0
Implementation-Title: petclinic
Implementation-Version: 2.1.0.BUILD-SNAPSHOT
Start-Class: org.springframework.samples.petclinic.PetClinicApplicatio
 n
Spring-Boot-Classes: BOOT-INF/classes/
Spring-Boot-Lib: BOOT-INF/lib/
Build-Jdk-Spec: 1.8
Spring-Boot-Version: 2.1.6.RELEASE
Created-By: Maven Archiver 3.4.0
Main-Class: org.springframework.boot.loader.JarLauncher
`), 0644)).To(Succeed())

		m, err := libjvm.NewManifest(path)
		Expect(err).NotTo(HaveOccurred())

		k, ok := m.Get("Start-Class")
		Expect(ok).To(BeTrue())
		Expect(k).To(Equal("org.springframework.samples.petclinic.PetClinicApplication"))
	})

}
