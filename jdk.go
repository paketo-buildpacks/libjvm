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

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
)

type JDK struct {
	Certificates     string
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewJDK(dependency libpak.BuildpackDependency, cache libpak.DependencyCache, certificates string, plan *libcnb.BuildpackPlan) (JDK, error) {
	expected := map[string]interface{}{"dependency": dependency}

	in, err := os.Open(certificates)
	if err != nil && !os.IsNotExist(err) {
		return JDK{}, fmt.Errorf("unable to open file %s\n%w", certificates, err)
	} else if err == nil {
		defer in.Close()

		s := sha256.New()
		if _, err := io.Copy(s, in); err != nil {
			return JDK{}, fmt.Errorf("unable to hash file %s\n%w", certificates, err)
		}
		expected["cacerts-sha256"] = hex.EncodeToString(s.Sum(nil))
	}

	layerContributor := libpak.NewDependencyLayerContributor(dependency, cache, plan)
	layerContributor.LayerContributor.ExpectedMetadata = expected

	return JDK{
		Certificates:     certificates,
		LayerContributor: layerContributor,
	}, nil
}

func (j JDK) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	j.LayerContributor.Logger = j.Logger

	return j.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		j.Logger.Bodyf("Expanding to %s", layer.Path)
		if err := crush.ExtractTarGz(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand JDK\n%w", err)
		}

		layer.BuildEnvironment.Override("JAVA_HOME", layer.Path)
		layer.BuildEnvironment.Override("JDK_HOME", layer.Path)

		var keyStorePath string
		if IsBeforeJava9(j.LayerContributor.Dependency.Version) {
			keyStorePath = filepath.Join(layer.Path, "jre", "lib", "security", "cacerts")
		} else {
			keyStorePath = filepath.Join(layer.Path, "lib", "security", "cacerts")
		}

		c := CertificateLoader{
			CACertificatesPath: j.Certificates,
			KeyStorePath:       keyStorePath,
			KeyStorePassword:   []byte("changeit"),
			Logger:             j.Logger.BodyWriter(),
		}

		if err := c.Load(); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load certificates\n%w", err)
		}

		layer.Build = true
		layer.Cache = true
		return layer, nil
	})
}

func (JDK) Name() string {
	return "jdk"
}
