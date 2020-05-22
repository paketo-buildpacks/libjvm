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
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/pavel-v-chernykh/keystore-go"
)

type CertificateLoader struct {
	CACertificatesPath string
	KeyStorePassword   string
	KeyStorePath       string
	Logger             io.Writer
}

func (c *CertificateLoader) Load() error {
	blocks, err := c.ReadBlocks()
	if err != nil {
		return fmt.Errorf("unable to read CA certificates\n%w", err)
	}

	switch i := len(blocks); {
	case i == 0:
		return nil
	default:
		_, _ = fmt.Fprintf(c.Logger, "Adding %d container CA certificates to JVM truststore\n", len(blocks))
	}

	ks, err := c.ReadKeyStore()
	if err != nil {
		return fmt.Errorf("unable to read keystore\n%w", err)
	}

	for i, b := range blocks {
		ks[fmt.Sprintf("openssl-%03d", i)] = &keystore.TrustedCertificateEntry{
			Entry: keystore.Entry{
				CreationDate: time.Now(),
			},
			Certificate: keystore.Certificate{
				Type:    "X.509",
				Content: b.Bytes,
			},
		}
	}

	if err := c.WriteKeyStore(ks); err != nil {
		return fmt.Errorf("unable to write keystore\n%w", err)
	}

	return nil
}

func (c CertificateLoader) ReadBlocks() ([]*pem.Block, error) {
	rest, err := ioutil.ReadFile(c.CACertificatesPath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read %s\n%w", c.CACertificatesPath, err)
	}

	var (
		block  *pem.Block
		blocks []*pem.Block
	)
	for len(rest) != 0 {
		block, rest = pem.Decode(rest)
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (c CertificateLoader) ReadKeyStore() (keystore.KeyStore, error) {
	in, err := os.Open(c.KeyStorePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s\n%w", c.KeyStorePath, err)
	}
	defer in.Close()

	ks, err := keystore.Decode(in, []byte(c.KeyStorePassword))
	if err != nil {
		return nil, fmt.Errorf("unable to decode keystore\n %w", err)
	}

	return ks, nil
}

func (c CertificateLoader) WriteKeyStore(ks keystore.KeyStore) error {
	out, err := os.OpenFile(c.KeyStorePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open %s\n%w", c.KeyStorePath, err)
	}
	defer out.Close()

	if err := keystore.Encode(out, ks, []byte(c.KeyStorePassword)); err != nil {
		return fmt.Errorf("unable to encode keystore\n%w", err)
	}

	return nil
}
