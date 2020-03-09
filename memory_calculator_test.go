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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"
)

func testMemoryCalculator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = ioutil.TempDir("", "memory-calculator-application")
		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = ioutil.TempDir("", "memory-calculator-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it("contributes Memory Calculator", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-memory-calculator.tgz",
			SHA256: "3a357182b2314f0059eeb1b222ae0635d5ed6931defad24da33643b23e802647",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j := libjvm.NewMemoryCalculator(ctx.Application.Path, dep, dc, "8.0.212", &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(filepath.Join(layer.Path, "bin", "java-buildpack-memory-calculator")).To(BeARegularFile())
		Expect(layer.Profile["memory-calculator"]).To(Equal(fmt.Sprintf(`HEAD_ROOM=${BPL_HEAD_ROOM:=0}

if [[ -z "${BPL_LOADED_CLASS_COUNT+x}" ]]; then
    LOADED_CLASS_COUNT=$(class-counter --source "%s" --jvm-class-count "27867")
else
	LOADED_CLASS_COUNT=${BPL_LOADED_CLASS_COUNT}
fi

THREAD_COUNT=${BPL_THREAD_COUNT:=250}

TOTAL_MEMORY=$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)

if [ ${TOTAL_MEMORY} -eq 9223372036854771712 ]; then
  printf "Container memory limit unset. Configuring JVM for 1G container.\n"
  TOTAL_MEMORY=1073741824
elif [ ${TOTAL_MEMORY} -gt 70368744177664 ]; then
  printf "Container memory limit too large. Configuring JVM for 64T container.\n"
  TOTAL_MEMORY=70368744177664
fi

MEMORY_CONFIGURATION=$(java-buildpack-memory-calculator \
    --head-room "${HEAD_ROOM}" \
    --jvm-options "${JAVA_OPTS}" \
    --loaded-class-count "${LOADED_CLASS_COUNT}" \
    --thread-count "${THREAD_COUNT}" \
    --total-memory "${TOTAL_MEMORY}")

printf "Calculated JVM Memory Configuration: ${MEMORY_CONFIGURATION} (Head Room: ${HEAD_ROOM}%%%%, Loaded Class Count: ${LOADED_CLASS_COUNT}, Thread Count: ${THREAD_COUNT}, Total Memory: ${TOTAL_MEMORY})\n"
export JAVA_OPTS="${JAVA_OPTS} ${MEMORY_CONFIGURATION}"
`, ctx.Application.Path)))
	})

	context("jvmClassCount", func() {
		it("counts Java 8 JRE", func() {
			Expect(libjvm.MemoryCalculator{JavaVersion: "8.0.232"}.JvmClassCount()).To(Equal(27867))
		})

		it("counts Java 9 JRE", func() {
			Expect(libjvm.MemoryCalculator{JavaVersion: "9.0.4"}.JvmClassCount()).To(Equal(25565))
		})

		it("counts Java 10 JRE", func() {
			Expect(libjvm.MemoryCalculator{JavaVersion: "10.0.2"}.JvmClassCount()).To(Equal(28191))
		})

		it("counts Java 11 JRE", func() {
			Expect(libjvm.MemoryCalculator{JavaVersion: "11.0.1"}.JvmClassCount()).To(Equal(24219))
		})

		it("counts Java 12 JRE", func() {
			Expect(libjvm.MemoryCalculator{JavaVersion: "12.0.1"}.JvmClassCount()).To(Equal(24219))
		})

		it("counts Java 13 JRE", func() {
			Expect(libjvm.MemoryCalculator{JavaVersion: "13.0.1"}.JvmClassCount()).To(Equal(24219))
		})

		it("counts Java 14 JRE", func() {
			Expect(libjvm.MemoryCalculator{JavaVersion: "14.0.1"}.JvmClassCount()).To(Equal(24219))
		})

		it("counts unknown JRE", func() {
			Expect(libjvm.MemoryCalculator{JavaVersion: "24.0.1"}.JvmClassCount()).To(Equal(24219))
		})
	})
}
