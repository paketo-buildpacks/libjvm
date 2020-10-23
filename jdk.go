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
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type JDK struct {
	CertificateDirs  []string
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewJDK(dependency libpak.BuildpackDependency, cache libpak.DependencyCache, certificateDirs []string, plan *libcnb.BuildpackPlan) (JDK, error) {
	expected := map[string]interface{}{"dependency": dependency}

	var err error
	expected["certificates"], err = sherpa.NewFileListing(certificateDirs...)
	if err != nil {
		return JDK{}, fmt.Errorf("unable to create file listing for %s\n%w", certificateDirs, err)
	}

	layerContributor := libpak.NewDependencyLayerContributor(dependency, cache, plan)
	layerContributor.LayerContributor.ExpectedMetadata = expected

	return JDK{
		CertificateDirs:  certificateDirs,
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
			CertificateDirs:  j.CertificateDirs,
			KeyStorePath:     keyStorePath,
			KeyStorePassword: "changeit",
			Logger:           j.Logger.BodyWriter(),
		}

		if err := c.Load(); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load certificates\n%w", err)
		}

		return layer, nil
	}, libpak.BuildLayer, libpak.CacheLayer)
}

func (j JDK) Name() string {
	return j.LayerContributor.LayerName()
}
