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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-shellwords"
	"github.com/paketo-buildpacks/libpak/bard"
)

type JVMHeapDump struct {
	Logger bard.Logger
}

func (a JVMHeapDump) Execute() (map[string]string, error) {
	heapDumpPath, ok := os.LookupEnv("BPL_HEAP_DUMP_PATH")
	if !ok || heapDumpPath == "" {
		return nil, nil
	}

	// mkdir as the JVM will not create it, it just fails and you lose the dump
	err := os.MkdirAll(heapDumpPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("unable to create heap dump path %s\n%w", heapDumpPath, err)
	}

	heapDumpPath = filepath.Join(heapDumpPath, fmt.Sprintf("java_%s.hprof", time.Now().Format(time.RFC3339)))

	var values []string
	s, ok := os.LookupEnv("JAVA_TOOL_OPTIONS")
	if ok {
		values = append(values, s)
	}

	if p, err := shellwords.Parse(s); err != nil {
		return nil, fmt.Errorf("unable to parse $JAVA_TOOL_OPTIONS\n%w", err)
	} else {
		found := false

		for _, s := range p {
			if s == "-XX:+HeapDumpOnOutOfMemoryError" {
				found = true
				break
			}
		}

		if !found {
			a.Logger.Info("Enabling HeapDumpOnOutOfMemoryError")
			values = append(values, "-XX:+HeapDumpOnOutOfMemoryError")
		}

		found = false
		for _, s := range p {
			if strings.HasPrefix(s, "-XX:HeapDumpPath=") {
				found = true
				break
			}
		}

		if !found {
			a.Logger.Infof("Setting HeapDumpPath to %s", heapDumpPath)
			values = append(values, fmt.Sprintf("-XX:HeapDumpPath=%s", heapDumpPath))
		}
	}

	return map[string]string{"JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil
}
