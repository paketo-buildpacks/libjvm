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

package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/paketo-buildpacks/libpak/bard"

	"github.com/paketo-buildpacks/libjvm/calc"
	"github.com/paketo-buildpacks/libjvm/count"
)

const (
	ClassLoadFactor          = 0.35
	DefaultHeadroom          = 0
	DefaultMemoryLimitPathV1 = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	DefaultMemoryLimitPathV2 = "/sys/fs/cgroup/memory.max"
	DefaultMemoryInfoPath    = "/proc/meminfo"
	DefaultThreadCount       = 250
	MaxJVMSize               = 64 * calc.Tebi
	UnsetTotalMemory         = int64(9_223_372_036_854_771_712)
)

type MemoryCalculator struct {
	Logger            bard.Logger
	MemoryLimitPathV1 string
	MemoryLimitPathV2 string
	MemoryInfoPath    string
}

func (m MemoryCalculator) Execute() (map[string]string, error) {
	var (
		err error
		c   = calc.Calculator{
			HeadRoom:    DefaultHeadroom,
			ThreadCount: DefaultThreadCount,
		}
		deprecatedHeadroom bool
	)

	if s, ok := os.LookupEnv("BPL_JVM_HEADROOM"); ok {
		if c.HeadRoom, err = strconv.Atoi(s); err != nil {
			return nil, fmt.Errorf("unable to convert $BPL_JVM_HEADROOM=%s to integer\n%w", s, err)
		}
		deprecatedHeadroom = true
		m.Logger.Info("WARNING: BPL_JVM_HEADROOM is deprecated and will be removed, please switch to BPL_JVM_HEAD_ROOM")
	}

	if s, ok := os.LookupEnv("BPL_JVM_HEAD_ROOM"); ok {
		if c.HeadRoom, err = strconv.Atoi(s); err != nil {
			return nil, fmt.Errorf("unable to convert $BPL_JVM_HEAD_ROOM=%s to integer\n%w", s, err)
		}
		if deprecatedHeadroom {
			m.Logger.Info("WARNING: You have set both BPL_JVM_HEAD_ROOM and BPL_JVM_HEADROOM. BPL_JVM_HEADROOM has been deprecated, so it will be ignored.")
		}
	}

	if s, ok := os.LookupEnv("BPL_JVM_LOADED_CLASS_COUNT"); ok {
		if c.LoadedClassCount, err = strconv.Atoi(s); err != nil {
			return nil, fmt.Errorf("unable to convert $BPL_JVM_LOADED_CLASS_COUNT=%s to integer\n%w", s, err)
		}
	} else {
		appPath, ok := os.LookupEnv("BPI_APPLICATION_PATH")
		if !ok {
			return nil, fmt.Errorf("$BPI_APPLICATION_PATH must be set")
		}

		var j int
		if s, ok := os.LookupEnv("BPI_JVM_CLASS_COUNT"); !ok {
			return nil, fmt.Errorf("$BPI_JVM_CLASS_COUNT must be set")
		} else {
			if j, err = strconv.Atoi(s); err != nil {
				return nil, fmt.Errorf("unable to convert $BPI_JVM_CLASS_COUNT=%s to integer\n%w", s, err)
			}
		}

		staticAdjustment := 0
		adjustmentFactor := uint64(100)
		if adj, ok := os.LookupEnv("BPL_JVM_CLASS_ADJUSTMENT"); ok {
			if strings.HasSuffix(adj, "%") {
				if adjustmentFactor, err = strconv.ParseUint(strings.TrimSuffix(adj, "%"), 10, 32); err != nil {
					return nil, fmt.Errorf("unable to parse $BPL_JVM_CLASS_ADJUSTMENT %s as a percentage: %w", adj, err)
				}
			} else {
				if staticAdjustment, err = strconv.Atoi(adj); err != nil {
					return nil, fmt.Errorf("unable to parse $BPL_JVM_CLASS_ADJUSTMENT %s as an integer: %w", adj, err)
				}
			}
		}

		a, err := count.Classes(appPath)

		totalClasses := float64(j+a+staticAdjustment) * float64(adjustmentFactor) / 100.0

		if err != nil {
			return nil, fmt.Errorf("unable to determine class count\n%w", err)
		}
		m.Logger.Debugf("Memory Calculation: (%d%% * (%d + %d + %d)) * %0.2f", adjustmentFactor, j, a, staticAdjustment, ClassLoadFactor)
		c.LoadedClassCount = int(totalClasses * ClassLoadFactor)
	}

	if threadCount, ok := os.LookupEnv("BPL_JVM_THREAD_COUNT"); ok {
		if c.ThreadCount, err = strconv.Atoi(threadCount); err != nil {
			return nil, fmt.Errorf("unable to convert $BPL_JVM_THREAD_COUNT=%s to integer\n%w", threadCount, err)
		}
	}

	totalMemory := m.getMemoryLimitFromPath(m.MemoryLimitPathV1)
	if totalMemory == UnsetTotalMemory {
		totalMemory = m.getMemoryLimitFromPath(m.MemoryLimitPathV2)
	}

	if totalMemory == UnsetTotalMemory {
		if b, err := ioutil.ReadFile(m.MemoryInfoPath); err != nil && !os.IsNotExist(err) {
			m.Logger.Infof(`WARNING: failed to read %q: %s`, m.MemoryInfoPath, err)
		} else if err == nil {
			if mem, err := parseMemInfo(string(b)); err != nil {
				m.Logger.Infof(`WARNING: failed to parse available memory from path %q: %s`, m.MemoryInfoPath, err)
			} else {
				m.Logger.Infof("Calculating JVM memory based on %s available memory", calc.Size{Value: mem}.String())
				totalMemory = mem
			}
		}
	}

	if totalMemory == UnsetTotalMemory {
		m.Logger.Info("WARNING: Unable to determine memory limit. Configuring JVM for 1G container.")
		c.TotalMemory = calc.Size{Value: calc.Gibi}
	} else if totalMemory > MaxJVMSize {
		m.Logger.Info("WARNING: Container memory limit too large. Configuring JVM for 64T container.")
		c.TotalMemory = calc.Size{Value: MaxJVMSize}
	} else {
		c.TotalMemory = calc.Size{Value: totalMemory}
	}

	var values []string
	s, ok := os.LookupEnv("JAVA_TOOL_OPTIONS")
	if ok {
		values = append(values, s)
	}

	r, err := c.Calculate(s)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate memory configuration\n%w", err)
	}

	var calculated []string
	if r.DirectMemory.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.DirectMemory.String())
	}
	if r.Heap.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.Heap.String())
	}
	if r.Metaspace.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.Metaspace.String())
	}
	if r.ReservedCodeCache.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.ReservedCodeCache.String())
	}
	if r.Stack.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.Stack.String())
	}
	values = append(values, calculated...)

	m.Logger.Infof("Calculated JVM Memory Configuration: %s (Total Memory: %s, Thread Count: %d, Loaded Class Count: %d, Headroom: %d%%)",
		strings.Join(calculated, " "), c.TotalMemory, c.ThreadCount, c.LoadedClassCount, c.HeadRoom)

	return map[string]string{"JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil
}

func (m MemoryCalculator) getMemoryLimitFromPath(memoryLimitPath string) int64 {
	if b, err := ioutil.ReadFile(memoryLimitPath); err != nil && !os.IsNotExist(err) {
		m.Logger.Infof("WARNING: Unable to read %s: %s", memoryLimitPath, err)
	} else if err == nil {
		if size, err := calc.ParseSize(strings.TrimSpace(string(b))); err != nil {
			m.Logger.Infof("WARNING: Unable to convert memory limit %q from path %q as int: %s", strings.TrimSpace(string(b)), memoryLimitPath, err)
		} else {
			return size.Value
		}
	}
	return UnsetTotalMemory
}

func parseMemInfo(s string) (int64, error) {
	pattern := `MemAvailable:\s*(\d+)(.*)`
	rp := regexp.MustCompile(pattern)
	if !rp.MatchString(s) {
		return 0, fmt.Errorf("failed to match pattern '%s'", pattern)
	}
	matches := rp.FindStringSubmatch(s)

	num, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to convert available memory %s to integer\n%w", matches[1], err)
	}
	unit, err := calc.ParseUnit(matches[2])
	if err != nil {
		return 0, err
	}
	return num * unit, nil
}
