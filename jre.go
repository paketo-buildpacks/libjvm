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
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/magiconair/properties"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"

	"github.com/paketo-buildpacks/libjvm/count"
)

type JRE struct {
	ApplicationPath   string
	CertificateLoader CertificateLoader
	DistributionType  DistributionType
	LayerContributor  libpak.DependencyLayerContributor
	Logger            bard.Logger
	Metadata          map[string]interface{}
}

func NewJRE(applicationPath string, dependency libpak.BuildpackDependency, cache libpak.DependencyCache,
	distributionType DistributionType, certificateLoader CertificateLoader, metadata map[string]interface{},
	plan *libcnb.BuildpackPlan) (JRE, error) {

	expected := map[string]interface{}{"dependency": dependency}

	if md, err := certificateLoader.Metadata(); err != nil {
		return JRE{}, fmt.Errorf("unable to generate certificate loader metadata")
	} else {
		for k, v := range md {
			expected[k] = v
		}
	}

	layerContributor := libpak.NewDependencyLayerContributor(dependency, cache, plan)
	layerContributor.LayerContributor.ExpectedMetadata = expected

	return JRE{
		ApplicationPath:   applicationPath,
		CertificateLoader: certificateLoader,
		DistributionType:  distributionType,
		LayerContributor:  layerContributor,
		Metadata:          metadata,
	}, nil
}

func (j JRE) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	j.LayerContributor.Logger = j.Logger

	var flags []libpak.LayerFlag
	if IsBuildContribution(j.Metadata) {
		flags = append(flags, libpak.BuildLayer, libpak.CacheLayer)
	}
	if IsLaunchContribution(j.Metadata) {
		flags = append(flags, libpak.LaunchLayer)
	}

	return j.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		j.Logger.Bodyf("Expanding to %s", layer.Path)
		if err := crush.ExtractTarGz(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand JRE\n%w", err)
		}

		var cacertsPath string
		if IsBeforeJava9(j.LayerContributor.Dependency.Version) && j.DistributionType == JDKType {
			cacertsPath = filepath.Join(layer.Path, "jre", "lib", "security", "cacerts")
		} else {
			cacertsPath = filepath.Join(layer.Path, "lib", "security", "cacerts")
		}

		if err := j.CertificateLoader.Load(cacertsPath, "changeit"); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load certificates\n%w", err)
		}

		if IsBuildContribution(j.Metadata) {
			layer.BuildEnvironment.Default("JAVA_HOME", layer.Path)
		}

		if IsLaunchContribution(j.Metadata) {
			layer.LaunchEnvironment.Default("BPI_APPLICATION_PATH", j.ApplicationPath)
			layer.LaunchEnvironment.Default("BPI_JVM_CACERTS", cacertsPath)

			if c, err := count.Classes(layer.Path); err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to count JVM classes\n%w", err)
			} else {
				layer.LaunchEnvironment.Default("BPI_JVM_CLASS_COUNT", c)
			}

			if IsBeforeJava9(j.LayerContributor.Dependency.Version) && j.DistributionType == JDKType {
				layer.LaunchEnvironment.Default("BPI_JVM_EXT_DIR", filepath.Join(layer.Path, "jre", "lib", "ext"))
			} else if IsBeforeJava9(j.LayerContributor.Dependency.Version) && j.DistributionType == JREType {
				layer.LaunchEnvironment.Default("BPI_JVM_EXT_DIR", filepath.Join(layer.Path, "lib", "ext"))
			}

			var file string
			if IsBeforeJava9(j.LayerContributor.Dependency.Version) && j.DistributionType == JDKType {
				file = filepath.Join(layer.Path, "jre", "lib", "security", "java.security")
			} else if IsBeforeJava9(j.LayerContributor.Dependency.Version) && j.DistributionType == JREType {
				file = filepath.Join(layer.Path, "lib", "security", "java.security")
			} else {
				file = filepath.Join(layer.Path, "conf", "security", "java.security")
			}

			p, err := properties.LoadFile(file, properties.UTF8)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to read properties file %s\n%w", file, err)
			}
			p = p.FilterStripPrefix("security.provider.")

			var providers []string
			for k, v := range p.Map() {
				providers = append(providers, fmt.Sprintf("%s|%s", k, v))
			}
			layer.LaunchEnvironment.Default("BPI_JVM_SECURITY_PROVIDERS", strings.Join(providers, " "))

			layer.LaunchEnvironment.Default("JAVA_HOME", layer.Path)
			layer.LaunchEnvironment.Default("MALLOC_ARENA_MAX", "2")
		}

		return layer, nil
	}, flags...)
}

func (j JRE) Name() string {
	return j.LayerContributor.LayerName()
}
