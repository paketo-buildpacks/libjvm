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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var ClassExtensions = []string{".class", ".classdata", ".clj", ".groovy", ".kts"}

func Classes(path string) (int, error) {
	file := filepath.Join(path, "lib", "modules")
	if _, err := os.Stat(file); err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("unable to stat %s\n%w", file, err)
	} else if os.IsNotExist(err) {
		return JarClasses(path)
	} else {
		return ModuleClasses(file)
	}
}

func JarClasses(path string) (int, error) {
	count := 0

	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, e := range ClassExtensions {
			if strings.HasSuffix(path, e) {
				count++
				return nil
			}
		}

		if !strings.HasSuffix(path, ".jar") || info.IsDir() {
			return nil
		}

		// Check for zero byte JAR files with name containing 'none' - these can not be unzipped
		// examples of these were found in the JDK, e.g. svm-none.jar
		// See https://github.com/paketo-buildpacks/libjvm/issues/84
		if info.Size() == 0 && strings.Contains(info.Name(), "none") {
			return nil
		}

		z, err := zip.OpenReader(path)
		if err != nil {
			if !(errors.Is(err, zip.ErrFormat)) {
				return fmt.Errorf("unable to open Jar %s\n%w", path, err)
			} else {
				return nil
			}
		}
		defer z.Close()

		for _, f := range z.File {
			if strings.HasSuffix(f.FileInfo().Name(), ".jar") {
				c, err := nestedJarContents(f)
				if err != nil {
					return fmt.Errorf("unable to counted nested jar%w\n", err)
				}
				count += c
			}
			count += jarContents(f)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("unable to walk %s\n%w", path, err)
	}

	return count, nil
}

func ModuleClasses(file string) (int, error) {
	count := 0

	i, err := NewImage(file)
	if err != nil {
		return 0, fmt.Errorf("unable to find JVM modules in %s\n%w", file, err)
	}

	for _, o := range i.Offsets.Entries {
		l, err := i.Locations.Get(o)
		if err != nil {
			return 0, fmt.Errorf("unable to inventory JVM modules\n%w", err)
		}

		if l.ExtensionOffset != 0 {
			e, err := l.Extension(i.Strings)
			if err != nil {
				return 0, fmt.Errorf("unable to inventory JVM modules\n%w", err)
			}

			if e == "class" {
				count++
			}
		}
	}

	return count, nil
}

func JarClassesFrom(paths ...string) (int, int, error) {
	var agentClassCount, skippedPaths int

	for _, path := range paths {
		if c, err := JarClasses(path); err == nil {
			agentClassCount += c
		} else if errors.Is(err, fs.ErrNotExist) {
			skippedPaths++
			continue
		} else {
			return 0, 0, fmt.Errorf("unable to count classes of jar at %s\n%w", path, err)
		}
	}
	return agentClassCount, skippedPaths, nil
}

func jarContents(file *zip.File) int {
	var count = 0
	for _, e := range ClassExtensions {
		if strings.HasSuffix(file.Name, e) {
			count++
			break
		}
	}
	return count
}

func nestedJarContents(jarFile *zip.File) (int, error) {
	var count = 0

	reader, err := jarFile.Open()
	if err != nil {
		return 0, fmt.Errorf("unable to open nested jar%w\n", err)
	}
	defer reader.Close()

	var b bytes.Buffer
	size, err := io.Copy(&b, reader)
	if err != nil {
		return 0, fmt.Errorf("error copying nested Jar \n%w", err)
	}
	br := bytes.NewReader(b.Bytes())
	nj, err := zip.NewReader(br, size)
	if err != nil {
		if !(errors.Is(err, zip.ErrFormat)) {
			return 0, fmt.Errorf("error reading nested Jar contents\n%w", err)
		} else {
			return 0, nil
		}
	}
	for _, nestedJar := range nj.File {
		count += jarContents(nestedJar)
	}
	return count, nil
}
