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

		it("formats Mibi", func() {
			Expect(calc.Size{Value: calc.Mibi + 1023}.String()).To(Equal("1M"))
		})

		it("formats Gibi", func() {
			Expect(calc.Size{Value: calc.Gibi + 1023}.String()).To(Equal("1G"))
		})

		it("formats Tibi", func() {
			Expect(calc.Size{Value: calc.Tibi + 1023}.String()).To(Equal("1T"))
		})

		it("formats larger than Tibi", func() {
			Expect(calc.Size{Value: (calc.Tibi * 1024) + 1023}.String()).To(Equal("1024T"))
		})
	})

	context("parse", func() {

		it("parses bytes", func() {
			Expect(calc.ParseSize("1")).To(Equal(calc.Size{Value: 1}))
			Expect(calc.ParseSize("1b")).To(Equal(calc.Size{Value: 1}))
		})

		it("parses Kibi", func() {
			Expect(calc.ParseSize("1k")).To(Equal(calc.Size{Value: calc.Kibi}))
			Expect(calc.ParseSize("1K")).To(Equal(calc.Size{Value: calc.Kibi}))
		})

		it("parses Mibi", func() {
			Expect(calc.ParseSize("1m")).To(Equal(calc.Size{Value: calc.Mibi}))
			Expect(calc.ParseSize("1M")).To(Equal(calc.Size{Value: calc.Mibi}))
		})

		it("parses Gibi", func() {
			Expect(calc.ParseSize("1g")).To(Equal(calc.Size{Value: calc.Gibi}))
			Expect(calc.ParseSize("1G")).To(Equal(calc.Size{Value: calc.Gibi}))
		})

		it("parses Tibi", func() {
			Expect(calc.ParseSize("1t")).To(Equal(calc.Size{Value: calc.Tibi}))
			Expect(calc.ParseSize("1T")).To(Equal(calc.Size{Value: calc.Tibi}))
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

}
