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
	"strings"

	"github.com/mattn/go-shellwords"
)

type MemoryRegions struct {
	DirectMemory      DirectMemory
	HeadRoom          *HeadRoom
	Heap              *Heap
	Metaspace         *Metaspace
	ReservedCodeCache ReservedCodeCache
	Stack             Stack
}

func NewMemoryRegionsFromFlags(flags string) (MemoryRegions, error) {
	m := MemoryRegions{
		DirectMemory:      DefaultDirectMemory,
		ReservedCodeCache: DefaultReservedCodeCache,
		Stack:             DefaultStack,
	}

	p, err := shellwords.Parse(flags)
	if err != nil {
		return MemoryRegions{}, fmt.Errorf("unable to parse flags\n%w", err)
	}

	for _, f := range p {
		if MatchDirectMemory(f) {
			m.DirectMemory, err = ParseDirectMemory(f)
			if err != nil {
				return MemoryRegions{}, fmt.Errorf("unable to parse direct memory\n%w", err)
			}
			m.DirectMemory.Provenance = UserConfigured
		} else if MatchHeap(f) {
			m.Heap, err = ParseHeap(f)
			if err != nil {
				return MemoryRegions{}, fmt.Errorf("unable to parse heap\n%w", err)
			}
			m.Heap.Provenance = UserConfigured
		} else if MatchMetaspace(f) {
			m.Metaspace, err = ParseMetaspace(f)
			if err != nil {
				return MemoryRegions{}, fmt.Errorf("unable to parse metaspace\n%w", err)
			}
			m.Metaspace.Provenance = UserConfigured
		} else if MatchReservedCodeCache(f) {
			m.ReservedCodeCache, err = ParseReservedCodeCache(f)
			if err != nil {
				return MemoryRegions{}, fmt.Errorf("unable to parse reserved code cache\n%w", err)
			}
			m.ReservedCodeCache.Provenance = UserConfigured
		} else if MatchStack(f) {
			m.Stack, err = ParseStack(f)
			if err != nil {
				return MemoryRegions{}, fmt.Errorf("unable to parse stack\n%w", err)
			}
			m.Stack.Provenance = UserConfigured
		}
	}

	return m, nil
}

func (m MemoryRegions) AllRegionsSize(threadCount int) (Size, error) {
	if m.HeadRoom == nil {
		return Size{}, fmt.Errorf("unable to calculate all regions size without heap")
	}

	s, err := m.NonHeapRegionsSize(threadCount)
	if err != nil {
		return Size{}, fmt.Errorf("unable to calculate non-heap regions size\n%w", err)
	}

	return Size{
		Value:      s.Value + m.Heap.Value,
		Provenance: Calculated,
	}, nil
}

func (m MemoryRegions) AllRegionsString(threadCount int) string {
	var s []string

	if m.HeadRoom != nil {
		s = append(s, m.Heap.String())
	}
	s = append(s, m.NonHeapRegionsString(threadCount))

	return strings.Join(s, ", ")
}

func (m MemoryRegions) FixedRegionsSize(threadCount int) (Size, error) {
	if m.Metaspace == nil {
		return Size{}, fmt.Errorf("unable to calculate fixed regions size without metaspace")
	}

	return Size{
		Value:      m.DirectMemory.Value + m.Metaspace.Value + m.ReservedCodeCache.Value + (m.Stack.Value * int64(threadCount)),
		Provenance: Calculated,
	}, nil
}

func (m MemoryRegions) FixedRegionsString(threadCount int) string {
	var s []string

	s = append(s, m.DirectMemory.String())
	if m.Metaspace != nil {
		s = append(s, m.Metaspace.String())
	}
	s = append(s, m.ReservedCodeCache.String())
	s = append(s, fmt.Sprintf("%s * %d threads", m.Stack.String(), threadCount))

	return strings.Join(s, ", ")
}

func (m MemoryRegions) NonHeapRegionsSize(threadCount int) (Size, error) {
	if m.HeadRoom == nil {
		return Size{}, fmt.Errorf("unable to calculate non-heap regions size without headroom")
	}

	s, err := m.FixedRegionsSize(threadCount)
	if err != nil {
		return Size{}, fmt.Errorf("unable to calculate fixed regions size\n%w", err)
	}

	return Size{
		Value:      s.Value + m.HeadRoom.Value,
		Provenance: Calculated,
	}, nil
}

func (m MemoryRegions) NonHeapRegionsString(threadCount int) string {
	var s []string

	if m.HeadRoom != nil {
		s = append(s, fmt.Sprintf("%s headroom", m.HeadRoom.String()))
	}
	s = append(s, m.FixedRegionsString(threadCount))

	return strings.Join(s, ", ")
}
