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
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/paketo-buildpacks/libpak/v2/log"
	"github.com/paketo-buildpacks/libpak/v2/sherpa"
	"golang.org/x/sys/unix"
)

const DefaultCertFile = "/etc/ssl/certs/ca-certificates.crt"

var NormalizedDateTime = time.Date(1980, time.January, 1, 0, 0, 1, 0, time.UTC)

type CertificateLoader struct {
	CertFile string
	CertDirs []string
	Logger   log.Logger
}

func NewCertificateLoader(logger log.Logger) CertificateLoader {
	c := CertificateLoader{CertFile: DefaultCertFile}

	if s, ok := os.LookupEnv("SSL_CERT_FILE"); ok {
		c.CertFile = s
	}

	if s, ok := os.LookupEnv("SSL_CERT_DIR"); ok {
		c.CertDirs = filepath.SplitList(s)
	}

	c.Logger = logger

	return c
}

func (c *CertificateLoader) Load(path string, password string) error {
	if unix.Access(path, unix.W_OK) != nil {
		return nil
	}

	ks, err := DetectKeystore(path)
	if err != nil {
		return err
	}

	files, err := c.certFiles()
	if err != nil {
		return fmt.Errorf("unable to identify cert files in %s and %s\n%w", c.CertFile, c.CertDirs, err)
	}

	added := 0
	for _, f := range files {
		blocks, err := c.readBlocks(f)
		if err != nil {
			return fmt.Errorf("unable to read certificates from %s\n%w", f, err)
		}

		for i, b := range blocks {
			ks.Add(fmt.Sprintf("%s-%d", f, i), b)
			added++
		}
	}

	c.Logger.Bodyf("Adding %d container CA certificates to JVM truststore\n", added)

	if err := ks.Write(); err != nil {
		return fmt.Errorf("unable to write keystore\n%w", err)
	}

	return nil
}

func (c CertificateLoader) certFiles() ([]string, error) {
	var files []string

	if _, err := os.Stat(c.CertFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to stat %s\n%w", c.CertFile, err)
	} else if err == nil {
		files = append(files, c.CertFile)
	}

	re := regexp.MustCompile(`^[[:xdigit:]]{8}\.[\d]$`)
	for _, d := range c.CertDirs {
		c, err := os.ReadDir(d)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("unable to list children of %s\n%w", d, err)
		}

		for _, c := range c {
			if c.IsDir() || !re.MatchString(c.Name()) {
				continue
			}

			files = append(files, filepath.Join(d, c.Name()))
		}
	}

	return files, nil
}

func (c *CertificateLoader) Metadata() (map[string]interface{}, error) {
	var (
		err      error
		metadata = make(map[string]interface{})
	)

	if in, err := os.Open(c.CertFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to open %s\n%w", c.CertFile, err)
	} else if err == nil {
		defer in.Close()

		out := sha256.New()
		if _, err := io.Copy(out, in); err != nil {
			return nil, fmt.Errorf("unable to hash file %s\n%w", c.CertFile, err)
		}
		metadata["cert-file"] = hex.EncodeToString(out.Sum(nil))
	}

	if metadata["cert-dir"], err = sherpa.NewFileListing(c.CertDirs...); err != nil {
		return nil, fmt.Errorf("unable to create file listing for %s\n%w", c.CertDirs, err)
	}

	return metadata, nil
}

func (c CertificateLoader) readBlocks(path string) ([]*pem.Block, error) {
	var (
		block  *pem.Block
		blocks []*pem.Block
	)

	rest, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s\n%w", path, err)
	}

	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}
