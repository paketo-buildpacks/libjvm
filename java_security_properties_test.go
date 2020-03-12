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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/sclevine/spec"
)

func testJavaSecurityProperties(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "java-security-properties-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it("contributes Java Security Properties", func() {
		l := libjvm.NewJavaSecurityProperties(ctx.Buildpack.Info)
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = l.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())

		file := filepath.Join(layer.Path, "java-security.properties")
		Expect(file).To(BeARegularFile())
		Expect(layer.LaunchEnvironment["JAVA_OPTS.append"]).To(Equal(fmt.Sprintf(` -Djava.security.properties=%s`, file)))
		Expect(layer.LaunchEnvironment["JAVA_SECURITY_PROPERTIES.override"]).To(Equal(file))
	})
}
