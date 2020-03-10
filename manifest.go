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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/magiconair/properties"
)

// NewManifest reads the <APP>/META-INF/MANIFEST.MF file if it exists, normalizing it into the standard properties
// form.
func NewManifest(applicationPath string) (*properties.Properties, error) {
	file := filepath.Join(applicationPath, "META-INF", "MANIFEST.MF")

	in, err := os.Open(file)
	if os.IsNotExist(err) {
		return properties.NewProperties(), nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", file, err)
	}
	defer in.Close()

	b, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", file, err)
	}

	// The full grammar for manifests can be found here:
	// https://docs.oracle.com/javase/8/docs/technotes/guides/jar/jar.html#JARManifest

	// Convert Windows style line endings to UNIX
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))

	// The spec allows newlines to be single carriage-returns
	// this is a legacy line ending only supported on System 9
	// and before.
	b = bytes.ReplaceAll(b, []byte("\r"), []byte("\n"))

	// The spec only allowed for line lengths of 78 bytes.
	// All lines are blank, start a property name or are
	// a continuation of the previous lines (indicated by a leading space).
	b = bytes.ReplaceAll(b, []byte("\n "), []byte{})

	p, err := properties.Load(b, properties.UTF8)
	if err != nil {
		return nil, fmt.Errorf("unable to parse properties from %s: %w", file, err)
	}

	return p, nil
}
