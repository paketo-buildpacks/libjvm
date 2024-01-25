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

package helper_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/anthonydahanne/libjvm/calc"
	"github.com/anthonydahanne/libjvm/helper"
)

func testMemoryCalculator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		applicationPath   string
		memoryLimitPathV1 string
		memoryLimitPathV2 string
		memoryInfoPath    string
		m                 helper.MemoryCalculator
	)

	it.Before(func() {
		var err error

		applicationPath, err = ioutil.TempDir("", "memory-calculator-application")
		Expect(err).NotTo(HaveOccurred())

		limitV1, err := ioutil.TempFile("", "memory-calculator-memory-limit-v1")
		Expect(err).NotTo(HaveOccurred())
		Expect(limitV1.Close()).To(Succeed())
		Expect(os.RemoveAll(limitV1.Name())).To(Succeed())
		memoryLimitPathV1 = limitV1.Name()

		limitV2, err := ioutil.TempFile("", "memory-calculator-memory-limit-v2")
		Expect(err).NotTo(HaveOccurred())
		Expect(limitV2.Close()).To(Succeed())
		Expect(os.RemoveAll(limitV2.Name())).To(Succeed())
		memoryLimitPathV2 = limitV2.Name()

		info, err := ioutil.TempFile("", "memory-calculator-memory-info")
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Close()).To(Succeed())
		Expect(os.RemoveAll(info.Name())).To(Succeed())
		memoryInfoPath = info.Name()

		m = helper.MemoryCalculator{
			MemoryLimitPathV1: memoryLimitPathV1,
			MemoryLimitPathV2: memoryLimitPathV2,
			MemoryInfoPath:    memoryInfoPath,
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(applicationPath)).To(Succeed())
		Expect(os.RemoveAll(memoryLimitPathV1)).To(Succeed())
		Expect(os.RemoveAll(memoryLimitPathV2)).To(Succeed())
	})

	it("returns error if $BPI_APPLICATION_PATH is not set", func() {
		_, err := m.Execute()

		Expect(err).To(MatchError("$BPI_APPLICATION_PATH must be set"))
	})

	context("$BPI_APPLICATION_PATH", func() {
		it.Before(func() {
			Expect(os.Setenv("BPI_APPLICATION_PATH", applicationPath)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BPI_APPLICATION_PATH")).To(Succeed())
		})

		it("returns error if $BPI_JVM_CLASS_COUNT is not set", func() {
			_, err := m.Execute()

			Expect(err).To(MatchError("$BPI_JVM_CLASS_COUNT must be set"))
		})

		context("$BPI_JVM_CLASS_COUNT", func() {
			it.Before(func() {
				Expect(os.Setenv("BPI_JVM_CLASS_COUNT", "100")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPI_JVM_CLASS_COUNT")).To(Succeed())
			})

			it("returns default options", func() {
				Expect(m.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
				}))
			})

			context("$BPL_JVM_CLASS_ADJUSTMENT", func() {
				context("set to a static number", func() {
					context("positive number", func() {
						it.Before(func() {
							Expect(os.Setenv("BPL_JVM_CLASS_ADJUSTMENT", "10")).To(Succeed())
						})

						it.After(func() {
							Expect(os.Unsetenv("BPL_JVM_CLASS_ADJUSTMENT")).To(Succeed())
						})

						it("adjusts by a static factor", func() {
							Expect(m.Execute()).To(Equal(map[string]string{
								"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522688K -XX:MaxMetaspaceSize=13887K -XX:ReservedCodeCacheSize=240M -Xss1M",
							}))
						})
					})

					context("negative number", func() {
						it.Before(func() {
							Expect(os.Setenv("BPL_JVM_CLASS_ADJUSTMENT", "-40")).To(Succeed())
						})

						it.After(func() {
							Expect(os.Unsetenv("BPL_JVM_CLASS_ADJUSTMENT")).To(Succeed())
						})

						it("adjusts by a static factor", func() {
							Expect(m.Execute()).To(Equal(map[string]string{
								"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522785K -XX:MaxMetaspaceSize=13790K -XX:ReservedCodeCacheSize=240M -Xss1M",
							}))
						})
					})
				})

				context("set to a percentage", func() {
					context("value above 100%", func() {
						it.Before(func() {
							Expect(os.Setenv("BPL_JVM_CLASS_ADJUSTMENT", "110%")).To(Succeed())
						})

						it.After(func() {
							Expect(os.Unsetenv("BPL_JVM_CLASS_ADJUSTMENT")).To(Succeed())
						})

						it("adjusts by a percentage", func() {
							Expect(m.Execute()).To(Equal(map[string]string{
								"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522688K -XX:MaxMetaspaceSize=13887K -XX:ReservedCodeCacheSize=240M -Xss1M",
							}))
						})
					})

					context("value between 0 and 100%", func() {
						it.Before(func() {
							Expect(os.Setenv("BPL_JVM_CLASS_ADJUSTMENT", "60%")).To(Succeed())
						})

						it.After(func() {
							Expect(os.Unsetenv("BPL_JVM_CLASS_ADJUSTMENT")).To(Succeed())
						})

						it("adjusts by a percentage", func() {
							Expect(m.Execute()).To(Equal(map[string]string{
								"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522785K -XX:MaxMetaspaceSize=13790K -XX:ReservedCodeCacheSize=240M -Xss1M",
							}))
						})
					})

					context("negative number", func() {
						it.Before(func() {
							Expect(os.Setenv("BPL_JVM_CLASS_ADJUSTMENT", "-40%")).To(Succeed())
						})

						it.After(func() {
							Expect(os.Unsetenv("BPL_JVM_CLASS_ADJUSTMENT")).To(Succeed())
						})

						it("fails", func() {
							_, err := m.Execute()
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError(HavePrefix("unable to parse $BPL_JVM_CLASS_ADJUSTMENT %s as a percentage", "-40%")))
						})
					})
				})
			})

			context("$BPL_JVM_HEADROOM", func() {
				it.Before(func() {
					Expect(os.Setenv("BPL_JVM_HEADROOM", "10")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BPL_JVM_HEADROOM")).To(Succeed())
				})

				it("passes $BPL_JVM_HEADROOM to calculator", func() {
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx417848K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})
			})

			context("$BPL_JVM_HEAD_ROOM", func() {
				it.Before(func() {
					Expect(os.Setenv("BPL_JVM_HEAD_ROOM", "10")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BPL_JVM_HEAD_ROOM")).To(Succeed())
				})

				it("passes $BPL_JVM_HEAD_ROOM to calculator", func() {
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx417848K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})
			})

			context("$BPL_JVM_HEADROOM and $BPL_JVM_HEAD_ROOM", func() {
				it.Before(func() {
					Expect(os.Setenv("BPL_JVM_HEADROOM", "20")).To(Succeed())
					Expect(os.Setenv("BPL_JVM_HEAD_ROOM", "10")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BPL_JVM_HEADROOM")).To(Succeed())
					Expect(os.Unsetenv("BPL_JVM_HEAD_ROOM")).To(Succeed())
				})

				it("passes $BPL_JVM_HEAD_ROOM to calculator", func() {
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx417848K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})
			})

			context("$BPL_JVM_LOADED_CLASS_COUNT", func() {
				it.Before(func() {
					Expect(os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", "100")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")).To(Succeed())
				})

				it("passes $BPL_JVM_LOADED_CLASS_COUNT to calculator", func() {
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522337K -XX:MaxMetaspaceSize=14238K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})
			})

			context("$BPL_JVM_THREAD_COUNT", func() {
				it.Before(func() {
					Expect(os.Setenv("BPL_JVM_THREAD_COUNT", "100")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BPL_JVM_THREAD_COUNT")).To(Succeed())
				})

				it("passes $BPL_JVM_THREAD_COUNT to calculator", func() {
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx676305K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})
			})

			it("limits total memory to all available memory if no memory limit set", func() {
				const s = `
					MemTotal:       16400152 kB
					MemFree:        10477724 kB
					MemAvailable:   11514136 kB
					Buffers:          112396 kB
				`

				Expect(ioutil.WriteFile(memoryLimitPathV1, strconv.AppendInt([]byte{}, helper.UnsetTotalMemory, 10), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(memoryInfoPath, []byte(s), 0755)).To(Succeed())

				Expect(m.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx10988265K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
				}))
			})

			it("limits total memory to 1G if unable to get amount of available memory", func() {
				const s = `
					MemTotal:       16400152 kB
					MemFree:        10477724 kB
					MemAvailable:   WILL NOT PARSE
					Buffers:          112396 kB
				`

				Expect(ioutil.WriteFile(memoryLimitPathV1, strconv.AppendInt([]byte{}, helper.UnsetTotalMemory, 10), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(memoryInfoPath, []byte(s), 0755)).To(Succeed())

				Expect(m.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
				}))
			})

			it("limits total memory to 1G if unable to determine total memory", func() {
				Expect(ioutil.WriteFile(memoryLimitPathV1, strconv.AppendInt([]byte{}, helper.UnsetTotalMemory, 10), 0755)).To(Succeed())

				Expect(m.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
				}))
			})

			it("limits total memory to 64T", func() {
				Expect(ioutil.WriteFile(memoryLimitPathV1, strconv.AppendInt([]byte{}, helper.MaxJVMSize+1, 10), 0755)).To(Succeed())

				Expect(m.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx68718950865K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
				}))
			})

			it("limits total memory to container size if set", func() {
				Expect(ioutil.WriteFile(memoryLimitPathV1, strconv.AppendInt([]byte{}, 10*calc.Gibi, 10), 0755)).To(Succeed())

				Expect(m.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx9959889K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
				}))
			})

			it("limits total memory to container size if V2 set", func() {
				Expect(ioutil.WriteFile(memoryLimitPathV2, strconv.AppendInt([]byte{}, 11*calc.Gibi, 10), 0755)).To(Succeed())

				Expect(m.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx11008465K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
				}))
			})

			context("$JAVA_TOOL_OPTIONS", func() {
				it.Before(func() {
					Expect(os.Setenv("JAVA_TOOL_OPTIONS", "test-java-tool-options")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
				})

				it("returns default options appended to existing $JAVA_TOOL_OPTIONS", func() {
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "test-java-tool-options -XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})
			})

			context("$JAVA_TOOL_OPTIONS with agents", func() {
				it.Before(func() {
					Expect(os.Setenv("JAVA_TOOL_OPTIONS", fmt.Sprintf("-javaagent:%s", filepath.Join("../count/testdata", "stub-dependency.jar")))).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
				})

				it("counts classes of agent jars supplied via $JAVA_TOOL_OPTIONS", func() {
					c, err := m.CountAgentClasses(os.Getenv("JAVA_TOOL_OPTIONS"))
					Expect(err).NotTo(HaveOccurred())
					Expect(c).To(Equal(2))
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": fmt.Sprintf("-javaagent:%s -XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M", filepath.Join("../count/testdata", "stub-dependency.jar")),
					}))
				})

				it("skips counting classes if agent jar(s) supplied via $JAVA_TOOL_OPTIONS can't be found", func() {
					Expect(os.Setenv("JAVA_TOOL_OPTIONS", fmt.Sprintf("-javaagent:!abc -javaagent:%s", filepath.Join("../count/testdata", "stub-dependency.jar")))).To(Succeed())
					c, err := m.CountAgentClasses(os.Getenv("JAVA_TOOL_OPTIONS"))
					Expect(err).NotTo(HaveOccurred())
					Expect(c).To(Equal(2))
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": fmt.Sprintf("-javaagent:!abc -javaagent:%s -XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M", filepath.Join("../count/testdata", "stub-dependency.jar")),
					}))
				})

				it("skips counting agent classes if no agent jar(s) are supplied", func() {
					Expect(os.Setenv("JAVA_TOOL_OPTIONS", "")).To(Succeed())
					c, err := m.CountAgentClasses(os.Getenv("JAVA_TOOL_OPTIONS"))
					Expect(err).NotTo(HaveOccurred())
					Expect(c).To(Equal(0))
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": " -XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})

				it("does not change Metaspace if it has been user configured", func() {
					Expect(os.Setenv("JAVA_TOOL_OPTIONS", fmt.Sprintf("-XX:MaxMetaspaceSize=20000K -javaagent:%s", filepath.Join("../count/testdata", "stub-dependency.jar")))).To(Succeed())
					c, err := m.CountAgentClasses(os.Getenv("JAVA_TOOL_OPTIONS"))
					Expect(err).NotTo(HaveOccurred())
					Expect(c).To(Equal(2))
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "-XX:MaxMetaspaceSize=20000K -javaagent:../count/testdata/stub-dependency.jar -XX:MaxDirectMemorySize=10M -Xmx516576K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})
			})

			context("user configured", func() {
				it.Before(func() {
					Expect(os.Setenv("JAVA_TOOL_OPTIONS", "-XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
				})

				it("returns default options appended to existing $JAVA_TOOL_OPTIONS", func() {
					Expect(m.Execute()).To(Equal(map[string]string{
						"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx522705K -XX:MaxMetaspaceSize=13870K -XX:ReservedCodeCacheSize=240M -Xss1M",
					}))
				})
			})
		})
	})
}
