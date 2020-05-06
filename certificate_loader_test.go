/*
 * Copyright 2018-2020, VMware, Inc. All Rights Reserved.
 * Proprietary and Confidential.
 * Unauthorized use, copying or distribution of this source code via any medium is
 * strictly prohibited without the express written consent of VMware, Inc.
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
