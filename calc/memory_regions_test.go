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

func testMemoryRegions(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		m calc.MemoryRegions
	)

	context("NewMemoryRegionsFromFlags", func() {

		it("defaults", func() {
			Expect(calc.NewMemoryRegionsFromFlags("")).To(Equal(calc.MemoryRegions{
				DirectMemory:      calc.DefaultDirectMemory,
				ReservedCodeCache: calc.DefaultReservedCodeCache,
				Stack:             calc.DefaultStack,
			}))
		})

		it("with flags", func() {
			Expect(calc.NewMemoryRegionsFromFlags("-XX:MaxDirectMemorySize=1K -Xmx1K -XX:MaxMetaspaceSize=1K -XX:ReservedCodeCacheSize=1K -Xss1K")).
				To(Equal(calc.MemoryRegions{
					DirectMemory:      calc.DirectMemory{Value: calc.Kibi, Provenance: calc.UserConfigured},
					Heap:              &calc.Heap{Value: calc.Kibi, Provenance: calc.UserConfigured},
					Metaspace:         &calc.Metaspace{Value: calc.Kibi, Provenance: calc.UserConfigured},
					ReservedCodeCache: calc.ReservedCodeCache{Value: calc.Kibi, Provenance: calc.UserConfigured},
					Stack:             calc.Stack{Value: calc.Kibi, Provenance: calc.UserConfigured},
				}))
		})
	})

	context("all regions", func() {
		it.Before(func() {
			m = calc.MemoryRegions{
				DirectMemory:      calc.DirectMemory{Value: calc.Kibi},
				HeadRoom:          &calc.HeadRoom{Value: calc.Kibi},
				Heap:              &calc.Heap{Value: calc.Kibi},
				Metaspace:         &calc.Metaspace{Value: calc.Kibi},
				ReservedCodeCache: calc.ReservedCodeCache{Value: calc.Kibi},
				Stack:             calc.Stack{Value: calc.Kibi},
			}
		})

		it("returns error if heap is not set", func() {
			_, err := calc.MemoryRegions{}.AllRegionsSize(2)
			Expect(err).To(MatchError("unable to calculate all regions size without heap"))
		})

		it("returns size", func() {
			Expect(m.AllRegionsSize(2)).To(Equal(calc.Size{Value: 7 * calc.Kibi, Provenance: calc.Calculated}))
		})

		it("returns string", func() {
			Expect(m.AllRegionsString(2)).To(Equal(
				"-Xmx1K, 1K headroom, -XX:MaxDirectMemorySize=1K, -XX:MaxMetaspaceSize=1K, -XX:ReservedCodeCacheSize=1K, -Xss1K * 2 threads"))
		})
	})

	context("fixed regions", func() {
		it.Before(func() {
			m = calc.MemoryRegions{
				DirectMemory:      calc.DirectMemory{Value: calc.Kibi},
				Metaspace:         &calc.Metaspace{Value: calc.Kibi},
				ReservedCodeCache: calc.ReservedCodeCache{Value: calc.Kibi},
				Stack:             calc.Stack{Value: calc.Kibi},
			}
		})

		it("returns error if metaspace is not set", func() {
			_, err := calc.MemoryRegions{}.FixedRegionsSize(2)
			Expect(err).To(MatchError("unable to calculate fixed regions size without metaspace"))
		})

		it("returns size", func() {
			Expect(m.FixedRegionsSize(2)).To(Equal(calc.Size{Value: 5 * calc.Kibi, Provenance: calc.Calculated}))
		})

		it("returns string", func() {
			Expect(m.FixedRegionsString(2)).To(Equal(
				"-XX:MaxDirectMemorySize=1K, -XX:MaxMetaspaceSize=1K, -XX:ReservedCodeCacheSize=1K, -Xss1K * 2 threads"))
		})
	})

	context("non-heap regions", func() {
		it.Before(func() {
			m = calc.MemoryRegions{
				DirectMemory:      calc.DirectMemory{Value: calc.Kibi},
				HeadRoom:          &calc.HeadRoom{Value: calc.Kibi},
				Metaspace:         &calc.Metaspace{Value: calc.Kibi},
				ReservedCodeCache: calc.ReservedCodeCache{Value: calc.Kibi},
				Stack:             calc.Stack{Value: calc.Kibi},
			}
		})

		it("returns error if headroom is not set", func() {
			_, err := calc.MemoryRegions{}.NonHeapRegionsSize(2)
			Expect(err).To(MatchError("unable to calculate non-heap regions size without headroom"))
		})

		it("returns size", func() {
			Expect(m.NonHeapRegionsSize(2)).To(Equal(calc.Size{Value: 6 * calc.Kibi, Provenance: calc.Calculated}))
		})

		it("returns string", func() {
			Expect(m.NonHeapRegionsString(2)).To(Equal(
				"1K headroom, -XX:MaxDirectMemorySize=1K, -XX:MaxMetaspaceSize=1K, -XX:ReservedCodeCacheSize=1K, -Xss1K * 2 threads"))
		})
	})
}
