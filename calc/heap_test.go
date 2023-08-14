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

package calc_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/libjvm/v2/calc"
)

func testHeap(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("formats", func() {
		Expect(calc.Heap{Value: calc.Kibi}.String()).To(Equal("-Xmx1K"))
	})

	it("matches -Xmx", func() {
		Expect(calc.MatchHeap("-Xmx1K")).To(BeTrue())
	})

	it("does not match non -Xmx", func() {
		Expect(calc.MatchHeap("-Xss1K")).To(BeFalse())
	})

	it("parses", func() {
		Expect(calc.ParseHeap("-Xmx1K")).To(Equal(&calc.Heap{Value: calc.Kibi}))
	})

}
