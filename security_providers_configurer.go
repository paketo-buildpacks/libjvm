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
	_ "github.com/paketo-buildpacks/libjvm/statik"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type DistributionType uint8

const (
	JDKType DistributionType = iota
	JREType
)

func (d DistributionType) String() string {
	return []string{"jdk", "jre"}[d]
}

type SecurityProvidersConfigurer struct {
	DistributionType DistributionType
	JavaVersion      string
	LayerContributor libpak.HelperLayerContributor
	Logger           bard.Logger
}

func NewSecurityProvidersConfigurer(buildpack libcnb.Buildpack, distributionType DistributionType, javaVersion string,
	plan *libcnb.BuildpackPlan) SecurityProvidersConfigurer {

	layerContributor := libpak.NewHelperLayerContributor(filepath.Join(buildpack.Path, "bin", "security-providers-configurer"),
		"Security Providers Configurer", buildpack.Info, plan)
	layerContributor.LayerContributor.ExpectedMetadata = map[string]interface{}{
		"distribution-type": distributionType.String(),
		"info":              buildpack.Info,
		"java-version":      javaVersion,
	}

	return SecurityProvidersConfigurer{
		DistributionType: distributionType,
		JavaVersion:      javaVersion,
		LayerContributor: layerContributor,
	}
}

//go:generate statik -src . -include *.sh

func (s SecurityProvidersConfigurer) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	s.LayerContributor.Logger = s.Logger

	return s.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		s.Logger.Bodyf("Copying to %s", layer.Path)
		if err := sherpa.CopyFile(artifact, filepath.Join(layer.Path, "bin", "security-providers-configurer")); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy\n%w", err)
		}

		var source string
		if IsBeforeJava9(s.JavaVersion) {
			var extDir string
			switch s.DistributionType {
			case JDKType:
				extDir = filepath.Join("jre", "lib", "ext")
				source = filepath.Join("jre", "lib", "security", "java.security")
			case JREType:
				extDir = filepath.Join("lib", "ext")
				source = filepath.Join("lib", "security", "java.security")
			}

			s, err := sherpa.TemplateFile("/security-providers-classpath-8.sh", map[string]interface{}{"extDir": extDir})
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to load security-providers-classpath-8.sh\n%w", err)
			}

			layer.Profile.Add("security-providers-classpath.sh", s)
		} else {
			source = filepath.Join("conf", "security", "java.security")

			s, err := sherpa.StaticFile("/security-providers-classpath-9.sh")
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to load security-providers-classpath-9.sh\n%w", err)
			}

			layer.Profile.Add("security-providers-classpath.sh", s)
		}

		t, err := sherpa.TemplateFile("/security-providers-configurer.sh", map[string]interface{}{"source": source})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load security-providers-configurer.sh\n%w", err)
		}

		layer.Profile.Add("security-providers-configurer.sh", t)

		layer.Launch = true
		return layer, nil
	})
}

func (SecurityProvidersConfigurer) Name() string {
	return "security-providers-configurer"
}
