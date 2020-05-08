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
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/effect/mocks"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"
)

func testCertificateLoader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		executor *mocks.Executor
	)

	it.Before(func() {
		executor = &mocks.Executor{}
	})

	it("does not return error for non-existent file", func() {
		executor.On("Execute", mock.Anything).Return(nil)

		c := libjvm.CertificateLoader{
			KeyTool:         "test-key-tool",
			SourcePath:      filepath.Join("testdata", "invalid-certificates.crt"),
			DestinationPath: "test-path",
			Executor:        executor,
			Logger:          bard.NewLogger(ioutil.Discard),
		}

		Expect(c.Load()).To(Succeed())
		Expect(len(executor.Calls)).To(Equal(0))
	})

	it("calls keytool", func() {
		executor.On("Execute", mock.Anything).Return(nil)

		c := libjvm.CertificateLoader{
			KeyTool:         "test-key-tool",
			SourcePath:      filepath.Join("testdata", "ca-certificates.crt"),
			DestinationPath: "test-path",
			Executor:        executor,
			Logger:          bard.NewLogger(ioutil.Discard),
		}

		Expect(c.Load()).To(Succeed())
		Expect(len(executor.Calls)).To(Equal(133))
	})

}
