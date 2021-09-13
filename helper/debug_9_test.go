package helper_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm/helper"
	"github.com/sclevine/spec"
)

func testDebug9(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		d      = helper.Debug9{}
	)

	it("does nothing if $BPL_DEBUG_ENABLED is no set", func() {
		Expect(d.Execute()).To(BeNil())
	})

	context("$BPL_DEBUG_ENABLED", func() {

		it.Before(func() {
			Expect(os.Setenv("BPL_DEBUG_ENABLED", "true")).
				To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BPL_DEBUG_ENABLED")).To(Succeed())
		})

		it("contributes configuration", func() {
			Expect(d.Execute()).To(Equal(map[string]string{
				"JAVA_TOOL_OPTIONS": "-agentlib:jdwp=transport=dt_socket,server=y,address=*:8000,suspend=n",
			}))
		})

		context("$BPL_DEBUG_PORT", func() {
			it.Before(func() {
				Expect(os.Setenv("BPL_DEBUG_PORT", "8001")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPL_DEBUG_PORT")).To(Succeed())
			})

			it("contributes port configuration from $BPL_DEBUG_PORT", func() {
				Expect(d.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-agentlib:jdwp=transport=dt_socket,server=y,address=*:8001,suspend=n",
				}))
			})
		})

		context("$BPL_DEBUG_SUSPEND", func() {
			it.Before(func() {
				Expect(os.Setenv("BPL_DEBUG_SUSPEND", "true")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPL_DEBUG_SUSPEND")).To(Succeed())
			})

			it("contributes suspend configuration from $BPL_DEBUG_SUSPEND", func() {
				Expect(d.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-agentlib:jdwp=transport=dt_socket,server=y,address=*:8000,suspend=y",
				}))
			})
		})

		context("$JAVA_TOOL_OPTIONS", func() {
			it.Before(func() {
				Expect(os.Setenv("JAVA_TOOL_OPTIONS", "test-java-tool-options")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
			})

			it("contributes configuration appended to existing $JAVA_TOOL_OPTIONS", func() {
				Expect(d.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "test-java-tool-options -agentlib:jdwp=transport=dt_socket,server=y,address=*:8000,suspend=n",
				}))
			})
		})

	})
}
