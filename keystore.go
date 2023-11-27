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
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"software.sslmate.com/src/go-pkcs12"
)

type Keystore interface {
	Add(string, *pem.Block) error
	Write() error
}

func DetectKeystore(location string) (Keystore, error) {
	buf, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}

	if len(buf) == 0 {
		return nil, errors.New("empty keystore found")
	}

	if len(buf) > 3 && buf[0] == 0xFE && buf[1] == 0xED && buf[2] == 0xFE && buf[3] == 0xED {
		return NewJKSKeystore(location, "changeit")
	}

	return NewPasswordLessPKCS12Keystore(location)
}

var _ Keystore = &JKSKeystore{}

type JKSKeystore struct {
	location string
	password string
	store    keystore.KeyStore
}

func NewJKSKeystore(location, password string) (*JKSKeystore, error) {
	in, err := os.Open(location)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s\n%w", location, err)
	}
	defer in.Close()

	ks := keystore.New(keystore.WithOrderedAliases())
	if err := ks.Load(in, []byte(password)); err != nil {
		return nil, fmt.Errorf("unable to decode keystore\n %w", err)
	}
	return &JKSKeystore{
		location: location,
		password: password,
		store:    ks,
	}, nil
}

func (k *JKSKeystore) Add(name string, b *pem.Block) error {
	entry := keystore.TrustedCertificateEntry{
		CreationTime: NormalizedDateTime,
		Certificate: keystore.Certificate{
			Type:    "X.509",
			Content: b.Bytes,
		},
	}
	if err := k.store.SetTrustedCertificateEntry(name, entry); err != nil {
		return fmt.Errorf("unable to add trusted entry\n%w", err)
	}
	return nil
}

func (k *JKSKeystore) Write() error {
	out, err := os.OpenFile(k.location, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open %s\n%w", k.location, err)
	}
	defer out.Close()

	if err := k.store.Store(out, []byte(k.password)); err != nil {
		return fmt.Errorf("unable to encode keystore\n%w", err)
	}

	return nil
}

func (k *JKSKeystore) Len() int {
	return len(k.store.Aliases())
}

var _ Keystore = &PasswordLessPKCS12Keystore{}

type PasswordLessPKCS12Keystore struct {
	location string
	entries  []pkcs12.TrustStoreEntry
}

func NewPasswordLessPKCS12Keystore(location string) (*PasswordLessPKCS12Keystore, error) {
	in, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}

	x509Certs, err := pkcs12.DecodeTrustStore(in, "")
	if err != nil {
		return nil, err
	}

	var entries []pkcs12.TrustStoreEntry
	for _, x509Cert := range x509Certs {
		entries = append(entries, pkcs12.TrustStoreEntry{
			Cert:         x509Cert,
			FriendlyName: x509Cert.Subject.String(),
		})
	}

	return &PasswordLessPKCS12Keystore{
		location: location,
		entries:  entries,
	}, nil
}

func (k *PasswordLessPKCS12Keystore) Add(name string, b *pem.Block) error {
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return err
	}

	k.entries = append(k.entries, pkcs12.TrustStoreEntry{
		Cert:         cert,
		FriendlyName: name,
	})

	return nil
}

func (k *PasswordLessPKCS12Keystore) Write() error {
	out, err := os.OpenFile(k.location, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	data, err := pkcs12.Passwordless.EncodeTrustStoreEntries(k.entries, "")
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

func (k *PasswordLessPKCS12Keystore) Len() int {
	return len(k.entries)
}
