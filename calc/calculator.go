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

package calc

import (
	"fmt"
)

const (
	ClassSize     = int64(5_800)
	ClassOverhead = int64(14_000_000)
)

type Calculator struct {
	HeadRoom         int
	LoadedClassCount int
	ThreadCount      int
	TotalMemory      Size
}

func (c Calculator) Calculate(flags string) (MemoryRegions, error) {
	m, err := NewMemoryRegionsFromFlags(flags)
	if err != nil {
		return MemoryRegions{}, fmt.Errorf("unable to create memory regions from flags\n%w", err)
	}

	if m.Metaspace == nil {
		m.Metaspace = &Metaspace{
			Value:      ClassOverhead + (ClassSize * int64(c.LoadedClassCount)),
			Provenance: Calculated,
		}
	}

	f, err := m.FixedRegionsSize(c.ThreadCount)
	if err != nil {
		return MemoryRegions{}, fmt.Errorf("unable to calculate fixed regions size\n%w", err)
	}

	if f.Value > c.TotalMemory.Value {
		return MemoryRegions{}, fmt.Errorf(
			"fixed memory regions require %s which is greater than %s available for allocation: %s",
			f, c.TotalMemory, m.FixedRegionsString(c.ThreadCount),
		)
	}

	m.HeadRoom = &HeadRoom{
		Value:      int64((float64(c.HeadRoom) / 100) * float64(c.TotalMemory.Value)),
		Provenance: Calculated,
	}

	n, err := m.NonHeapRegionsSize(c.ThreadCount)
	if err != nil {
		return MemoryRegions{}, fmt.Errorf("unable to calculate non-heap regions size\n%w", err)
	}

	if n.Value > c.TotalMemory.Value {
		return MemoryRegions{}, fmt.Errorf(
			"non-heap memory regions require %s which is greater than %s available for allocation: %s",
			n, c.TotalMemory, m.NonHeapRegionsString(c.ThreadCount),
		)
	}

	if m.Heap == nil {
		m.Heap = &Heap{
			Value:      c.TotalMemory.Value - n.Value,
			Provenance: Calculated,
		}
	}

	a, err := m.AllRegionsSize(c.ThreadCount)
	if err != nil {
		return MemoryRegions{}, fmt.Errorf("unable to calculate all regions size\n%w", err)
	}

	if a.Value > c.TotalMemory.Value {
		return MemoryRegions{}, fmt.Errorf(
			"all memory regions require %s which is greater than %s available for allocation: %s",
			a, c.TotalMemory, m.AllRegionsString(c.ThreadCount))
	}

	return m, nil
}
