package libjvm

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/pavel-v-chernykh/keystore-go/v4"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"os"
	pkcs "software.sslmate.com/src/go-pkcs12"
)

type KeystoreOps interface {
	ReadKeystore() error
	LoadCerts() error
	WriteKeystore() error
	ReadLoadWrite() error
	CanRead() error
}

type PKCS12Keystore struct {
	Logger            io.Writer
	SystemCerts       map[string][]*pem.Block
	TrustStoreEntries []pkcs.TrustStoreEntry
	JVMCerts          []*x509.Certificate
	KeyStorePath      string
}

func CombineCerts(path string, systemCerts map[string][]*pem.Block, logger io.Writer) error {
	jks := JKSKeystore{
		Logger:       logger,
		SystemCerts:  systemCerts,
		KeyStorePath: path,
		Keystore:     keystore.New(keystore.WithOrderedAliases()),
	}
	pkcs := PKCS12Keystore{
		Logger:       logger,
		SystemCerts:  systemCerts,
		KeyStorePath: path,
	}
	var err error
	if err = jks.CanRead(); err == nil {
		if err = jks.ReadLoadWrite(); err != nil {
			return fmt.Errorf("unable to combine certs for JKS keystore\n %w", err)
		}
		return nil
	}
	if err = pkcs.CanRead(); err == nil {
		if err = pkcs.ReadLoadWrite(); err != nil {
			return fmt.Errorf("unable to combine certs for PKCS12 keystore\n %w", err)
		}
		return nil
	}
	return fmt.Errorf("unable to read JKS or PKCS12 keystore types\n %w", err)
}

func (p *PKCS12Keystore) ReadKeystore() error {
	if data, err := ioutil.ReadFile(p.KeyStorePath); err != nil {
		return fmt.Errorf("unable to open %s\n%w", p.KeyStorePath, err)
	} else {
		if certs, err := pkcs.DecodeTrustStore(data, ""); err != nil {
			return fmt.Errorf("unable to decode password-less PKCS12 keystore\n %w", err)
		} else {
			p.JVMCerts = append(p.JVMCerts, certs...)
		}
	}
	return nil
}

func (p *PKCS12Keystore) LoadCerts() error {
	added := 0
	// Load system certs as X509 certificates
	for _, certData := range p.SystemCerts {
		for _, block := range certData {
			if c, err := x509.ParseCertificate(block.Bytes); err != nil {
				return fmt.Errorf("unable to add trusted entry\n%w", err)
			} else {
				p.JVMCerts = append(p.JVMCerts, c)
			}
			added++
		}
	}
	_, _ = fmt.Fprintf(p.Logger, "Adding %d container CA certificates to JVM truststore\n", added)

	return nil
}

func (p *PKCS12Keystore) WriteKeystore() error {
	if unix.Access(p.KeyStorePath, unix.W_OK) != nil {
		_, _ = fmt.Fprintf(p.Logger, "WARNING: Unable to add container CA certificates to JVM because %s is read-only", p.KeyStorePath)
		return nil
	}
	out, err := os.OpenFile(p.KeyStorePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open %s\n%w", p.KeyStorePath, err)
	}
	defer out.Close()

	pfx, err := pkcs.EncodeTrustStore(rand.Reader, p.JVMCerts, "")
	if err != nil {
		return fmt.Errorf("unable to encode trusted entry\n%w", err)
	}
	if _, err = out.Write(pfx); err != nil {
		return fmt.Errorf("unable to write PKCS keystore\n%w", err)
	}
	return nil
}

func (p *PKCS12Keystore) ReadLoadWrite() error {
	if err := p.ReadKeystore(); err != nil {
		return fmt.Errorf("unable to read PKCS12 keystore\n%w", err)
	}
	if err := p.LoadCerts(); err != nil {
		return fmt.Errorf("unable to load system certs to PKCS12 keystore \n%w", err)
	}
	if err := p.WriteKeystore(); err != nil {
		return fmt.Errorf("unable to write to PKCS12 keystore\n%w", err)
	}
	return nil
}

func (p PKCS12Keystore) CanRead() error {
	if err := p.ReadKeystore(); err != nil {
		return fmt.Errorf("unable to read PKCS keystore\n%w", err)
	} else {
		return nil
	}
}

type JKSKeystore struct {
	Logger       io.Writer
	Keystore     keystore.KeyStore
	SystemCerts  map[string][]*pem.Block
	KeyStorePath string
}

func (j JKSKeystore) ReadKeystore() error {
	in, err := os.Open(j.KeyStorePath)
	if err != nil {
		return fmt.Errorf("unable to open %s\n%w", j.KeyStorePath, err)
	}
	defer in.Close()

	if err := j.Keystore.Load(in, []byte("changeit")); err != nil {
		return fmt.Errorf("unable to decode JKS keystore\n %w", err)
	}
	return nil
}

func (j JKSKeystore) LoadCerts() error {
	added := 0
	for file, certData := range j.SystemCerts {
		for i, block := range certData {

			entry := keystore.TrustedCertificateEntry{
				CreationTime: NormalizedDateTime,
				Certificate: keystore.Certificate{
					Type:    "X.509",
					Content: block.Bytes,
				},
			}
			if err := j.Keystore.SetTrustedCertificateEntry(fmt.Sprintf("%s-%d", file, i), entry); err != nil {
				return fmt.Errorf("unable to add trusted entry to JKS keystore\n%w", err)
			}
			added++
		}
	}
	_, _ = fmt.Fprintf(j.Logger, "Adding %d container CA certificates to JVM truststore\n", added)
	return nil
}

func (j JKSKeystore) WriteKeystore() error {
	if unix.Access(j.KeyStorePath, unix.W_OK) != nil {
		_, _ = fmt.Fprintf(j.Logger, "WARNING: Unable to add container CA certificates to JVM because %s is read-only", j.KeyStorePath)
		return nil
	}
	out, err := os.OpenFile(j.KeyStorePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open %s\n%w", j.KeyStorePath, err)
	}
	defer out.Close()

	if err := j.Keystore.Store(out, []byte("changeit")); err != nil {
		return fmt.Errorf("unable to encode JKS keystore\n%w", err)
	}
	return nil
}

func (j JKSKeystore) ReadLoadWrite() error {
	if err := j.ReadKeystore(); err != nil {
		return fmt.Errorf("unable to read JKS keystore\n%w", err)
	}
	if err := j.LoadCerts(); err != nil {
		return fmt.Errorf("unable to load system certs to JKS keystore \n%w", err)
	}
	if err := j.WriteKeystore(); err != nil {
		return fmt.Errorf("unable to write to JKS keystore\n%w", err)
	}
	return nil
}

func (j JKSKeystore) CanRead() error {
	if err := j.ReadKeystore(); err != nil {
		return fmt.Errorf("unable to read JKS keystore\n%w", err)
	} else {
		return nil
	}
}
