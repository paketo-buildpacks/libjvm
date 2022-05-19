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
	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sherpa"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/heroku/color"

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
	Executor          effect.Executor
}

func NewJRE(applicationPath string, dependency libpak.BuildpackDependency, cache libpak.DependencyCache, distributionType DistributionType, certificateLoader CertificateLoader, metadata map[string]interface{}) (JRE, libcnb.BOMEntry, error) {
	expected := map[string]interface{}{"dependency": dependency}

	if md, err := certificateLoader.Metadata(); err != nil {
		return JRE{}, libcnb.BOMEntry{}, fmt.Errorf("unable to generate certificate loader metadata\n%w", err)
	} else {
		for k, v := range md {
			expected[k] = v
		}
	}
	if jlinkArgs := sherpa.GetEnvWithDefault("BP_JVM_JLINK_ARGS", ""); jlinkArgs != "" {
		expected["jlink-args"] = jlinkArgs
	}

	contributor, be := libpak.NewDependencyLayer(dependency, cache, libcnb.LayerTypes{
		Build:  IsBuildContribution(metadata),
		Cache:  IsBuildContribution(metadata),
		Launch: IsLaunchContribution(metadata),
	})
	contributor.ExpectedMetadata = expected

	return JRE{
		ApplicationPath:   applicationPath,
		CertificateLoader: certificateLoader,
		DistributionType:  distributionType,
		LayerContributor:  contributor,
		Metadata:          metadata,
		Executor:          effect.NewExecutor(),
	}, be, nil
}

func (j JRE) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	j.LayerContributor.Logger = j.Logger

	return j.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		j.Logger.Bodyf("Expanding to %s", layer.Path)
		if err := crush.Extract(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand JRE\n%w", err)
		}

		var cacertsPath string
		if IsBeforeJava9(j.LayerContributor.Dependency.Version) && j.DistributionType == JDKType {
			cacertsPath = filepath.Join(layer.Path, "jre", "lib", "security", "cacerts")
		} else {
			cacertsPath = filepath.Join(layer.Path, "lib", "security", "cacerts")
		}

		if IsBeforeJava18(j.LayerContributor.Dependency.Version) {
			if err := j.CertificateLoader.Load(cacertsPath, "changeit"); err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to load certificates\n%w", err)
			}
		} else {
			j.Logger.Bodyf("%s: The JVM cacerts entries cannot be loaded with Java 18, for more information see: https://github.com/paketo-buildpacks/libjvm/issues/158", color.YellowString("Warning"))
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

			if !IsBeforeJava9(j.LayerContributor.Dependency.Version) && sherpa.ResolveBool("BP_JVM_JLINK_ENABLED") && j.DistributionType == JDKType {
				if err := j.buildCustomJRE(layer.Path); err != nil {
					return libcnb.Layer{}, fmt.Errorf("error building custom JRE with jlink \n%w", err)
				}
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
			sort.Strings(providers)
			layer.LaunchEnvironment.Default("BPI_JVM_SECURITY_PROVIDERS", strings.Join(providers, " "))

			layer.LaunchEnvironment.Default("JAVA_HOME", layer.Path)
			layer.LaunchEnvironment.Default("MALLOC_ARENA_MAX", "2")

			layer.LaunchEnvironment.Append("JAVA_TOOL_OPTIONS", " ", "-XX:+ExitOnOutOfMemoryError")
		}

		return layer, nil
	})
}

func (j JRE) Name() string {
	return j.LayerContributor.LayerName()
}

func (j JRE) buildCustomJRE(layerPath string) error {
	buildpackPath := filepath.Dir(layerPath)
	customDir := filepath.Join(buildpackPath, "custom-jre")

	def := fmt.Sprintf("--add-modules ALL-MODULE-PATH --no-man-pages --no-header-files --strip-debug --output %s", customDir)
	jlinkArgs := sherpa.GetEnvWithDefault("BP_JVM_JLINK_ARGS", def)
	if jlinkArgs != def {
		if strings.Contains(jlinkArgs, "output") {
			return fmt.Errorf("invalid custom jlink config, please remove 'output' arg\n")
		}
		jlinkArgs = fmt.Sprintf("%s --output %s", jlinkArgs, customDir)
	}

	if err := j.Executor.Execute(effect.Execution{
		Command: filepath.Join(layerPath, "bin", "jlink"),
		Args:    strings.Split(jlinkArgs, " "),
	}); err != nil {
		return fmt.Errorf("error running jlink\n%w", err)
	}
	if err := os.Rename(layerPath, filepath.Join(buildpackPath, "original-jdk")); err != nil {
		return fmt.Errorf("error renaming default JDK\n%w", err)
	}
	if err := os.Rename(customDir, layerPath); err != nil {
		return fmt.Errorf("error renaming custom JRE\n%w", err)
	}
	if err := os.RemoveAll(filepath.Join(buildpackPath, "original-jdk")); err != nil {
		return fmt.Errorf("error removing default JDK\n%w", err)
	}
	return nil
}
