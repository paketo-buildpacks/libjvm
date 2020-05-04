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

func testContributions(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("identifies a build contribution", func() {
		Expect(libjvm.IsBuildContribution(map[string]interface{}{"build": true})).To(BeTrue())
		Expect(libjvm.IsBuildContribution(map[string]interface{}{"launch": true})).To(BeFalse())
	})

	it("identifies a launch contribution", func() {
		Expect(libjvm.IsLaunchContribution(map[string]interface{}{"build": true})).To(BeFalse())
		Expect(libjvm.IsLaunchContribution(map[string]interface{}{"launch": true})).To(BeTrue())
	})

}
