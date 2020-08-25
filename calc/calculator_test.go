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

	"github.com/paketo-buildpacks/libjvm/calc"
)

func testCalculator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("calculates metaspace", func() {
		c := calc.Calculator{
			HeadRoom:         0,
			LoadedClassCount: 100,
			ThreadCount:      2,
			TotalMemory:      calc.Size{Value: calc.Gibi},
		}

		m, err := c.Calculate("")
		Expect(err).NotTo(HaveOccurred())

		Expect(m.Metaspace).To(Equal(&calc.Metaspace{Value: 14_580_000, Provenance: calc.Calculated}))
	})

	it("returns error if fixed regions are too large", func() {
		c := calc.Calculator{
			HeadRoom:         0,
			LoadedClassCount: 100,
			ThreadCount:      2,
			TotalMemory:      calc.Size{Value: calc.Kibi},
		}

		_, err := c.Calculate("")
		Expect(err).To(MatchError(
			"fixed memory regions require 272286K which is greater than 1K available for allocation: -XX:MaxDirectMemorySize=10M, -XX:MaxMetaspaceSize=14238K, -XX:ReservedCodeCacheSize=240M, -Xss1M * 2 threads"))
	})

	it("calculates head room", func() {
		c := calc.Calculator{
			HeadRoom:         1,
			LoadedClassCount: 100,
			ThreadCount:      2,
			TotalMemory:      calc.Size{Value: calc.Gibi},
		}

		m, err := c.Calculate("")
		Expect(err).NotTo(HaveOccurred())

		s := 0.01 * float64(calc.Gibi)
		Expect(m.HeadRoom).To(Equal(&calc.HeadRoom{Value: int64(s), Provenance: calc.Calculated}))
	})

	it("returns error if non-heap regions are too large", func() {
		c := calc.Calculator{
			HeadRoom:         1,
			LoadedClassCount: 100,
			ThreadCount:      2,
			TotalMemory:      calc.Size{Value: 272287 * calc.Kibi},
		}

		_, err := c.Calculate("")
		Expect(err).To(MatchError(
			"non-heap memory regions require 275009K which is greater than 272287K available for allocation: 2722K headroom, -XX:MaxDirectMemorySize=10M, -XX:MaxMetaspaceSize=14238K, -XX:ReservedCodeCacheSize=240M, -Xss1M * 2 threads"))
	})

	it("calculates heap", func() {
		c := calc.Calculator{
			HeadRoom:         0,
			LoadedClassCount: 100,
			ThreadCount:      2,
			TotalMemory:      calc.Size{Value: calc.Gibi},
		}

		m, err := c.Calculate("")
		Expect(err).NotTo(HaveOccurred())

		Expect(m.Heap).To(Equal(&calc.Heap{Value: 794920672, Provenance: calc.Calculated}))
	})

	it("returns error of all regions are too large", func() {
		c := calc.Calculator{
			HeadRoom:         0,
			LoadedClassCount: 100,
			ThreadCount:      2,
			TotalMemory:      calc.Size{Value: 272287 * calc.Kibi},
		}

		_, err := c.Calculate("-Xmx1M")
		Expect(err).To(MatchError(
			"all memory regions require 273310K which is greater than 272287K available for allocation: -Xmx1M, 0 headroom, -XX:MaxDirectMemorySize=10M, -XX:MaxMetaspaceSize=14238K, -XX:ReservedCodeCacheSize=240M, -Xss1M * 2 threads"))
	})

}
