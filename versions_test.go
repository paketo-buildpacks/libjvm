/*
 * Copyright 2018-2020, VMware, Inc. All Rights Reserved.
 * Proprietary and Confidential.
 * Unauthorized use, copying or distribution of this source code via any medium is
 * strictly prohibited without the express written consent of VMware, Inc.
 */

package libjvm_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/sclevine/spec"
)

func testVersions(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("determines whether a version is before Java 9", func() {
		Expect(libjvm.IsBeforeJava9("8.0.0")).To(BeTrue())
		Expect(libjvm.IsBeforeJava9("9.0.0")).To(BeFalse())
		Expect(libjvm.IsBeforeJava9("11.0.0")).To(BeFalse())
		Expect(libjvm.IsBeforeJava9("")).To(BeFalse())
	})

}
