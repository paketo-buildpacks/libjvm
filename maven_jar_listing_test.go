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
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm"
)

func testMavenJARListing(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("parses maven JARs", func() {
		for i := 0; i < 1000; i++ {
			Expect(libjvm.NewMavenJARListing(filepath.Join("testdata", "listing"))).To(Equal([]libjvm.MavenJAR{
				{
					Name:    "3-test-artifact.jar",
					Version: "unknown",
					SHA256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Name:    "test-artifact-1",
					Version: "1.2.3",
					SHA256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Name:    "test-artifact-1",
					Version: "7.8.9",
					SHA256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Name:    "test-artifact-2",
					Version: "4.5.6-SNAPSHOT",
					SHA256:  "06f961b802bc46ee168555f066d28f4f0e9afdf3f88174c1ee6f9de004fc30a0",
				},
				{
					Name:    "test-artifact-2",
					Version: "4.5.6-SNAPSHOT",
					SHA256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			}))
		}
	})

}
