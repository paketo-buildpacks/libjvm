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

	"github.com/Masterminds/semver/v3"
	"github.com/buildpacks/libcnb"
	_ "github.com/paketo-buildpacks/libjvm/statik"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type SecurityProvidersConfigurer struct {
	JavaVersion      string
	LayerContributor libpak.HelperLayerContributor
	Logger           bard.Logger
	Metadata         map[string]interface{}
}

func NewSecurityProvidersConfigurer(buildpack libcnb.Buildpack, javaVersion string, metadata map[string]interface{},
	plan *libcnb.BuildpackPlan) SecurityProvidersConfigurer {

	return SecurityProvidersConfigurer{
		JavaVersion: javaVersion,
		LayerContributor: libpak.NewHelperLayerContributor(filepath.Join(buildpack.Path, "bin", "security-providers-configurer"),
			"Security Providers Configurer", buildpack.Info, plan),
		Metadata: metadata,
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

		j9, _ := semver.NewVersion("9")
		v, err := semver.NewVersion(s.JavaVersion)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to parse Java version %s\n%w", s.JavaVersion, err)
		}

		var jdkSource string
		var jreSource string
		if v.LessThan(j9) {
			jdkSource = filepath.Join("jre", "lib", "security", "java.security")
			jreSource = filepath.Join("lib", "security", "java.security")

			s, err := sherpa.StaticFile("/security-providers-classpath-8.sh")
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to load security-providers-classpath-8.sh\n%w", err)
			}

			layer.Profile.Add("security-providers-classpath.sh", s)
		} else {
			jdkSource = filepath.Join("conf", "security", "java.security")
			jreSource = filepath.Join("conf", "security", "java.security")

			s, err := sherpa.StaticFile("/security-providers-classpath-9.sh")
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to load security-providers-classpath-9.sh\n%w", err)
			}

			layer.Profile.Add("security-providers-classpath.sh", s)
		}

		t, err := sherpa.TemplateFile("/security-providers-configurer.sh", map[string]interface{}{
			"jdkSource": jdkSource,
			"jreSource": jreSource,
		})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load security-providers-configurer.sh\n%w", err)
		}

		layer.Profile.Add("security-providers-configurer.sh", t)

		if isBuildContribution(s.Metadata) {
			layer.Build = true
			layer.Cache = true
		}

		if isLaunchContribution(s.Metadata) {
			layer.Launch = true
		}

		return layer, nil
	})
}

func (SecurityProvidersConfigurer) Name() string {
	return "security-providers-configurer"
}
