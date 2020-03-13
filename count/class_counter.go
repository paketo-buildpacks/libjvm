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

package count

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const LoadFactor = 0.35

type ClassCounter struct {
	JVMClassCount int
	SourcePath    string
}

func (c ClassCounter) Execute() (int, error) {
	count, err := c.directory(c.SourcePath)
	if err != nil {
		return 0, fmt.Errorf("unable to count classes in %s\n%w", c.SourcePath, err)
	}
	count += c.JVMClassCount

	return count, nil
}

func (c ClassCounter) archive(path string) (int, error) {
	count := 0

	z, err := zip.OpenReader(path)
	if err != nil {
		return 0, fmt.Errorf("unable to open %s\n%w", path, err)
	}
	defer z.Close()

	for _, f := range z.File {
		if !f.Mode().IsDir() {
			count += c.count(f.Name)
		}
	}

	return count, nil
}

func (ClassCounter) count(path string) int {
	for _, e := range []string{".class", ".groovy", ".kts"} {
		if strings.HasSuffix(path, e) {
			return 1
		}
	}
	return 0
}

func (c ClassCounter) directory(path string) (int, error) {
	count := 0

	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".jar") {
			a, err := c.archive(path)
			if err != nil {
				return fmt.Errorf("unable to count classes in archive %s\n%w", path, err)
			}

			count += a
			return nil
		}

		count += c.count(path)

		return nil
	}); err != nil {
		return 0, fmt.Errorf("unable to walk %s\n%w", path, err)
	}

	return count, nil
}
