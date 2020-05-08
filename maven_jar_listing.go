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

package libjvm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
)

var maven = regexp.MustCompile(".+/(.*)-([\\d].*)\\.jar")

// MavenJAR is metadata about a JRE entry that follows Maven naming conventions.
type MavenJAR struct {

	// Name is the name of the JAR, without the version or extension.
	Name string `toml:"name"`

	// Version is the version of the JAR, without the name or extension.
	Version string `toml:"version"`

	// SHA256 is the SHA256 hash of the JAR.
	SHA256 string `toml:"sha256"`
}

type result struct {
	err   error
	value MavenJAR
}

// NewMavenJARListing generates a listing of all JAR that follow Maven naming convention under the roots.
func NewMavenJARListing(roots ...string) ([]MavenJAR, error) {
	paths := make(chan string)
	results := make(chan result)

	go func() {
		for _, root := range roots {
			p, err := filepath.EvalSymlinks(root)
			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				results <- result{err: fmt.Errorf("unable to resolve %s\n%w", root, err)}
				return
			}

			if err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				if filepath.Ext(path) != ".jar" {
					return nil
				}

				paths <- path
				return nil
			}); err != nil {
				results <- result{err: fmt.Errorf("error walking path %s\n%w", root, err)}
				return
			}
		}

		close(paths)
	}()

	go func() {
		var workers sync.WaitGroup
		for i := 0; i < 128; i++ {
			workers.Add(1)
			go worker(paths, results, &workers)
		}

		workers.Wait()
		close(results)
	}()

	var m []MavenJAR
	for r := range results {
		if r.err != nil {
			return nil, fmt.Errorf("unable to create file listing: %s", r.err)
		}
		m = append(m, r.value)
	}
	sort.Slice(m, func(i, j int) bool {
		return m[i].Name < m[j].Name
	})

	return m, nil
}

func worker(paths chan string, results chan result, wg *sync.WaitGroup) {
	for path := range paths {
		m, err := process(path)
		results <- result{value: m, err: err}
	}

	wg.Done()
}

func process(path string) (MavenJAR, error) {
	m := MavenJAR{
		Name:    filepath.Base(path),
		Version: "unknown",
	}

	if p := maven.FindStringSubmatch(path); p != nil {
		m.Name = p[1]
		m.Version = p[2]
	}

	s := sha256.New()

	in, err := os.Open(path)
	if err != nil {
		return MavenJAR{}, fmt.Errorf("unable to open file %s\n%w", path, err)
	}
	defer in.Close()

	if _, err := io.Copy(s, in); err != nil {
		return MavenJAR{}, fmt.Errorf("unable to hash file %s\n%w", path, err)
	}

	m.SHA256 = hex.EncodeToString(s.Sum(nil))
	return m, nil
}
