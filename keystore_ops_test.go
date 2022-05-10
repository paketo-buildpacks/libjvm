package libjvm_test

import (
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/pavel-v-chernykh/keystore-go/v4"
	"github.com/sclevine/spec"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
	"testing"
)

func testKeystoreOps(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		c      = libjvm.CertificateLoader{
			CertFile: filepath.Join("testdata", "certificates", "certificate-1.pem"),
			Logger:   ioutil.Discard,
		}
		systemCerts, _ = c.LoadSystemCerts()
	)

	context("JKS Keystore", func() {
		var (
			path string
			jks  libjvm.JKSKeystore
		)

		it.Before(func() {
			in, err := os.Open(filepath.Join("testdata", "test-keystore.jks"))
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			out, err := ioutil.TempFile("", "certificate-loader")
			Expect(err).NotTo(HaveOccurred())
			defer out.Close()

			_, err = io.Copy(out, in)
			Expect(err).NotTo(HaveOccurred())

			path = out.Name()

			jks = libjvm.JKSKeystore{
				Logger:       ioutil.Discard,
				Keystore:     keystore.New(keystore.WithOrderedAliases()),
				SystemCerts:  systemCerts,
				KeyStorePath: path,
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("reads entries from JKS keystore", func() {

			Expect(jks.Keystore.Aliases()).To(HaveLen(0))
			err := jks.ReadKeystore()
			Expect(err).NotTo(HaveOccurred())
			Expect(jks.Keystore.Aliases()).To(HaveLen(1))
		})

		it("loads system certs to JKS keystore", func() {
			err := jks.LoadCerts()
			Expect(err).NotTo(HaveOccurred())
			Expect(jks.Keystore.Aliases()).To(HaveLen(1))
			Expect(jks.Keystore.Aliases()[0]).To(ContainSubstring("testdata/certificates/certificate-1.pem-0"))
		})

		it("writes system & JVM certs to JKS keystore", func() {

			err := jks.LoadCerts()
			Expect(err).NotTo(HaveOccurred())

			err = jks.WriteKeystore()
			Expect(err).NotTo(HaveOccurred())
			Expect(jks.KeyStorePath).To(BeARegularFile())

			in, err := os.Open(jks.KeyStorePath)
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			ks := keystore.New()
			err = ks.Load(in, []byte("changeit"))
			Expect(err).NotTo(HaveOccurred())
			Expect(ks.Aliases()).To(HaveLen(1))
		})

		it("combines system certs with existing entries for JKS keystore", func() {
			err := jks.ReadLoadWrite()
			Expect(err).NotTo(HaveOccurred())
			Expect(jks.Keystore.Aliases()).To(HaveLen(2))
		})

	})

	context("PKCS12 Keystore", func() {
		var (
			path string
			pkcs libjvm.PKCS12Keystore
		)

		it.Before(func() {
			in, err := os.Open(filepath.Join("testdata", "cacerts"))
			Expect(err).NotTo(HaveOccurred())
			defer in.Close()

			out, err := ioutil.TempFile("", "certificate-loader")
			Expect(err).NotTo(HaveOccurred())
			defer out.Close()

			_, err = io.Copy(out, in)
			Expect(err).NotTo(HaveOccurred())

			path = out.Name()

			pkcs = libjvm.PKCS12Keystore{
				Logger:       ioutil.Discard,
				SystemCerts:  systemCerts,
				KeyStorePath: path,
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("reads entries from PKCS keystore", func() {

			err := pkcs.ReadKeystore()
			Expect(err).NotTo(HaveOccurred())
			Expect(pkcs.JVMCerts).To(HaveLen(214))
		})

		it("loads system certs to PKCS keystore", func() {

			err := pkcs.LoadCerts()
			Expect(err).NotTo(HaveOccurred())
			Expect(pkcs.JVMCerts).To(HaveLen(1))
			Expect(pkcs.JVMCerts[0].Issuer.CommonName).To(Equal("ACCVRAIZ1"))
		})

		it("writes system certs to PKCS keystore", func() {

			err := pkcs.LoadCerts()
			Expect(err).NotTo(HaveOccurred())

			err = pkcs.WriteKeystore()
			Expect(err).NotTo(HaveOccurred())
			Expect(pkcs.KeyStorePath).To(BeARegularFile())

			in, err := ioutil.ReadFile(pkcs.KeyStorePath)
			Expect(err).NotTo(HaveOccurred())

			data, err := pkcs12.DecodeTrustStore(in, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(HaveLen(1))
		})

		it("combines system certs with existing entries for PKCS keystore", func() {
			err := pkcs.ReadLoadWrite()
			Expect(err).NotTo(HaveOccurred())
			Expect(pkcs.JVMCerts).To(HaveLen(215))

			Expect(pkcs.KeyStorePath).To(BeARegularFile())

			in, err := ioutil.ReadFile(pkcs.KeyStorePath)
			Expect(err).NotTo(HaveOccurred())

			data, err := pkcs12.DecodeTrustStore(in, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(HaveLen(215))
		})

	})
}
