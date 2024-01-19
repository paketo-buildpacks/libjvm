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

package helper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/libpak/v2/log"
	"github.com/paketo-buildpacks/libpak/v2/sherpa"
	"golang.org/x/sys/unix"

	"github.com/paketo-buildpacks/libjvm/v2"
)

var TmpTrustStore = filepath.Join(os.TempDir(), "truststore")

type OpenSSLCertificateLoader struct {
	CertificateLoader libjvm.CertificateLoader
	Logger            log.Logger
}

func (o OpenSSLCertificateLoader) prepareTempTrustStore(trustStore, tempTrustStore string) (map[string]string, error) {
	o.Logger.Bodyf("Using readonly truststore: %s", tempTrustStore)

	trustStoreFile, err := os.Open(trustStore)
	if err != nil {
		return nil, fmt.Errorf("unable to open trust store %s\n%w", trustStore, err)
	}
	defer trustStoreFile.Close()

	err = sherpa.CopyFile(trustStoreFile, tempTrustStore)
	if err != nil {
		return nil, fmt.Errorf("unable to copy dir (%s, %s)\n%w", trustStore, tempTrustStore, err)
	}

	opts := sherpa.AppendToEnvVar("JAVA_TOOL_OPTIONS", " ", fmt.Sprintf("-Djavax.net.ssl.trustStore=%s", tempTrustStore))
	o.Logger.Bodyf("changed JAVA_TOOL_OPTIONS: '%s'", opts)

	return map[string]string{"JAVA_TOOL_OPTIONS": opts}, nil
}

func (o OpenSSLCertificateLoader) Execute() (map[string]string, error) {
	trustStore, ok := os.LookupEnv("BPI_JVM_CACERTS")
	if !ok {
		return nil, fmt.Errorf("$BPI_JVM_CACERTS must be set")
	}

	trustStoreWriteable := true
	if unix.Access(trustStore, unix.W_OK) != nil {
		trustStoreWriteable = false
	}

	var opts map[string]string
	if !trustStoreWriteable {
		tmpOpts, err := o.prepareTempTrustStore(trustStore, TmpTrustStore)
		if err != nil {
			o.Logger.Body("Warning: import of certificates into Java Truststore will be skipped, because the JVM directory is read-only and creating a writable copy failed: ", err.Error())
			o.Logger.Body("If you need the certificates, please provide a writable temporary directroy, e.g. by mounting an empty ephemeral volume. Otherwise you can ignore this message.")
			return nil, nil
		}
		trustStore = TmpTrustStore
		opts = tmpOpts

	}

	if err := o.CertificateLoader.Load(trustStore, "changeit"); err != nil {
		return nil, fmt.Errorf("unable to load certificates\n%w", err)
	}

	return opts, nil

}
