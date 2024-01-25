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

	"github.com/anthonydahanne/libjvm/calc"
)

func testSize(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("format", func() {
		it("formats bytes", func() {
			Expect(calc.Size{Value: 1023}.String()).To(Equal("0"))
		})

		it("formats Kibi", func() {
			Expect(calc.Size{Value: calc.Kibi + 1023}.String()).To(Equal("1K"))
		})

		it("formats Mebi", func() {
			Expect(calc.Size{Value: calc.Mebi + 1023}.String()).To(Equal("1M"))
		})

		it("formats Gibi", func() {
			Expect(calc.Size{Value: calc.Gibi + 1023}.String()).To(Equal("1G"))
		})

		it("formats Tebi", func() {
			Expect(calc.Size{Value: calc.Tebi + 1023}.String()).To(Equal("1T"))
		})

		it("formats larger than Tebi", func() {
			Expect(calc.Size{Value: (calc.Tebi * 1024) + 1023}.String()).To(Equal("1024T"))
		})
	})

	context("parse", func() {
		it("parses bytes", func() {
			Expect(calc.ParseSize("1")).To(Equal(calc.Size{Value: 1}))
		})

		it("parses Kibi", func() {
			Expect(calc.ParseSize("1k")).To(Equal(calc.Size{Value: calc.Kibi}))
			Expect(calc.ParseSize("1K")).To(Equal(calc.Size{Value: calc.Kibi}))
		})

		it("parses Mebi", func() {
			Expect(calc.ParseSize("1m")).To(Equal(calc.Size{Value: calc.Mebi}))
			Expect(calc.ParseSize("1M")).To(Equal(calc.Size{Value: calc.Mebi}))
		})

		it("parses Gibi", func() {
			Expect(calc.ParseSize("1g")).To(Equal(calc.Size{Value: calc.Gibi}))
			Expect(calc.ParseSize("1G")).To(Equal(calc.Size{Value: calc.Gibi}))
		})

		it("parses Tebi", func() {
			Expect(calc.ParseSize("1t")).To(Equal(calc.Size{Value: calc.Tebi}))
			Expect(calc.ParseSize("1T")).To(Equal(calc.Size{Value: calc.Tebi}))
		})

		it("parses zero", func() {
			Expect(calc.ParseSize("0")).To(Equal(calc.Size{Value: 0}))
		})

		it("trims whitespace", func() {
			Expect(calc.ParseSize("\t\r\n 1")).To(Equal(calc.Size{Value: 1}))
			Expect(calc.ParseSize("1 \t\r\n")).To(Equal(calc.Size{Value: 1}))
		})

		it("does not parse empty value", func() {
			_, err := calc.ParseSize("")
			Expect(err).To(HaveOccurred())
		})

		it("does not parse negative value", func() {
			_, err := calc.ParseSize("-1")
			Expect(err).To(HaveOccurred())
		})

		it("does not parse unknown units", func() {
			_, err := calc.ParseSize("1A")
			Expect(err).To(HaveOccurred())
		})

		it("does not parse non-decimal value", func() {
			_, err := calc.ParseSize("0x1")
			Expect(err).To(HaveOccurred())
		})

		it("does not parse non-integral value", func() {
			_, err := calc.ParseSize("1.0")
			Expect(err).To(HaveOccurred())
		})

		it("does not parse embedded whitespace", func() {
			_, err := calc.ParseSize("1 0")
			Expect(err).To(HaveOccurred())
		})
	})

	context("parse", func() {
		it("parses bytes", func() {
			Expect(calc.ParseUnit("")).To(Equal(int64(1)))
			Expect(calc.ParseUnit("B")).To(Equal(int64(1)))
		})

		it("parses Kibi", func() {
			Expect(calc.ParseUnit("kB")).To(Equal(calc.Kibi))
			Expect(calc.ParseUnit("KB")).To(Equal(calc.Kibi))
			Expect(calc.ParseUnit("KiB")).To(Equal(calc.Kibi))
		})

		it("parses Mebi", func() {
			Expect(calc.ParseUnit("MB")).To(Equal(calc.Mebi))
			Expect(calc.ParseUnit("MiB")).To(Equal(calc.Mebi))
		})

		it("parses Gibi", func() {
			Expect(calc.ParseUnit("GB")).To(Equal(calc.Gibi))
			Expect(calc.ParseUnit("GiB")).To(Equal(calc.Gibi))
		})

		it("parses Tebi", func() {
			Expect(calc.ParseUnit("TB")).To(Equal(calc.Tebi))
			Expect(calc.ParseUnit("TiB")).To(Equal(calc.Tebi))
		})

		it("trims whitespace", func() {
			Expect(calc.ParseUnit("\t\r\n kB")).To(Equal(calc.Kibi))
			Expect(calc.ParseUnit("GB \t\r\n")).To(Equal(calc.Gibi))
		})

		it("does not parse unknown units", func() {
			_, err := calc.ParseUnit("X")
			Expect(err).To(HaveOccurred())
		})
	})
}
