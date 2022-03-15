/*
 * Copyright 2018-2022 the original author or authors.
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
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
		return nil, fmt.Errorf("unable to open %s\n%w", file, err)
	}
	defer in.Close()

	return loadManifest(in, file)
}

// NewManifestFromJAR reads the META-INF/MANIFEST.MF from a JAR file if it exists, normalizing it into the
// standard properties form.
func NewManifestFromJAR(jarFilePath string) (*properties.Properties, error) {
	// open the JAR file
	jarFile, err := zip.OpenReader(jarFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s\n%w", jarFilePath, err)
	}
	defer jarFile.Close()

	// look for the MANIFEST
	manifestFile, err := jarFile.Open("META-INF/MANIFEST.MF")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &properties.Properties{}, nil
		}
		return nil, fmt.Errorf("unable to read MANIFEST.MF in %s\n%w", jarFilePath, err)
	}

	return loadManifest(manifestFile, jarFilePath)
}

func loadManifest(reader io.Reader, source string) (*properties.Properties, error) {
	// read the MANIFEST
	manifestBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to read MANIFEST.MF in %s\n%w", source, err)
	}

	// Convert Windows style line endings to UNIX
	manifestBytes = bytes.ReplaceAll(manifestBytes, []byte("\r\n"), []byte("\n"))

	// The spec allows newlines to be single carriage-returns
	// this is a legacy line ending only supported on System 9
	// and before.
	manifestBytes = bytes.ReplaceAll(manifestBytes, []byte("\r"), []byte("\n"))

	// The spec only allowed for line lengths of 78 bytes.
	// All lines are blank, start a property name or are
	// a continuation of the previous lines (indicated by a leading space).
	manifestBytes = bytes.ReplaceAll(manifestBytes, []byte("\n "), []byte{})

	// parse the MANIFEST
	manifest, err := properties.Load(manifestBytes, properties.UTF8)
	if err != nil {
		return nil, fmt.Errorf("unable to parse properties in MANIFEST.MF in %s\n%w", source, err)
	}

	return manifest, nil
}
