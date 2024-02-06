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
	"sort"
	"strings"

	"github.com/buildpacks/libcnb/v2"
	"github.com/magiconair/properties"
	"github.com/paketo-buildpacks/libpak/v2"
	"github.com/paketo-buildpacks/libpak/v2/crush"
	"github.com/paketo-buildpacks/libpak/v2/log"

	"github.com/paketo-buildpacks/libjvm/v2/count"
)

type JRE struct {
	ApplicationPath   string
	CertificateLoader CertificateLoader
	DistributionType  DistributionType
	LayerContributor  libpak.DependencyLayerContributor
	Logger            log.Logger
	Metadata          map[string]interface{}
}

func NewJRE(applicationPath string, dependency libpak.BuildModuleDependency, cache libpak.DependencyCache, distributionType DistributionType, certificateLoader CertificateLoader, metadata map[string]interface{}) (JRE, error) {
	expected := map[string]interface{}{"dependency": dependency.GetMetadata()}

	if md, err := certificateLoader.Metadata(); err != nil {
		return JRE{}, fmt.Errorf("unable to generate certificate loader metadata\n%w", err)
	} else {
		for k, v := range md {
			expected[k] = v
		}
	}

	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Build:  IsBuildContribution(metadata),
		Cache:  IsBuildContribution(metadata),
		Launch: IsLaunchContribution(metadata),
	}, cache.Logger)
	contributor.ExpectedMetadata = expected

	return JRE{
		ApplicationPath:   applicationPath,
		CertificateLoader: certificateLoader,
		DistributionType:  distributionType,
		LayerContributor:  contributor,
		Metadata:          metadata,
		Logger:            cache.Logger,
	}, nil
}

type ConfigJREContext struct {
	Layer             *libcnb.Layer
	Logger            log.Logger
	JavaHome          string
	JavaVersion       string
	ApplicationPath   string
	IsBuild           bool
	IsLaunch          bool
	SkipCerts         bool
	CertificateLoader CertificateLoader
	DistType          DistributionType
}

func ConfigureJRE(configCtx ConfigJREContext) error {

	configCtx.Logger.Bodyf("Applying configuration for Java %s at %s", configCtx.JavaVersion, configCtx.JavaHome)

	// cacerts processing.
	var cacertsPath string
	if IsBeforeJava9(configCtx.JavaVersion) && configCtx.DistType == JDKType {
		cacertsPath = filepath.Join(configCtx.JavaHome, "jre", "lib", "security", "cacerts")
	} else {
		cacertsPath = filepath.Join(configCtx.JavaHome, "lib", "security", "cacerts")
	}
	if !configCtx.SkipCerts {
		if err := os.Chmod(cacertsPath, 0664); err != nil {
			return fmt.Errorf("unable to set keystore file permissions\n%w", err)
		}
		if err := configCtx.CertificateLoader.Load(cacertsPath, "changeit"); err != nil {
			return fmt.Errorf("unable to load certificates\n%w", err)
		}
	} else {
		//if we are skipping certs.. disable the runtime helper too.
		configCtx.Layer.BuildEnvironment.Override("BP_RUNTIME_CERT_BINDING_DISABLED", true)
	}

	if configCtx.IsBuild {
		configCtx.Logger.Body("Configuring for Build")
		configCtx.Layer.BuildEnvironment.Default("JAVA_HOME", configCtx.JavaHome)
	}

	if configCtx.IsLaunch {
		configCtx.Logger.Body("Configuring for Launch")

		configCtx.Layer.LaunchEnvironment.Default("BPI_APPLICATION_PATH", configCtx.ApplicationPath)
		configCtx.Layer.LaunchEnvironment.Default("BPI_JVM_CACERTS", cacertsPath)

		//count the classes in the runtime (used by memory calculator)
		if c, err := count.Classes(configCtx.JavaHome); err != nil {
			return fmt.Errorf("unable to count JVM classes\n%w", err)
		} else {
			configCtx.Layer.LaunchEnvironment.Default("BPI_JVM_CLASS_COUNT", c)
		}

		//applies to java 8 only...
		//locate ext dir to allow appending rather than replacing ext dir
		if IsBeforeJava9(configCtx.JavaVersion) && configCtx.DistType == JDKType {
			configCtx.Layer.LaunchEnvironment.Default("BPI_JVM_EXT_DIR", filepath.Join(configCtx.JavaHome, "jre", "lib", "ext"))
		} else if IsBeforeJava9(configCtx.JavaVersion) && configCtx.DistType == JREType {
			configCtx.Layer.LaunchEnvironment.Default("BPI_JVM_EXT_DIR", filepath.Join(configCtx.JavaHome, "lib", "ext"))
		}

		//locate the java security properties file...
		var securityFile string
		if IsBeforeJava9(configCtx.JavaVersion) && configCtx.DistType == JDKType {
			securityFile = filepath.Join(configCtx.JavaHome, "jre", "lib", "security", "java.security")
		} else if IsBeforeJava9(configCtx.JavaVersion) && configCtx.DistType == JREType {
			securityFile = filepath.Join(configCtx.JavaHome, "lib", "security", "java.security")
		} else {
			securityFile = filepath.Join(configCtx.JavaHome, "conf", "security", "java.security")
		}

		//Extract the security providers from the security properties file, and set
		//into BPI_JVM_SECURITY_PROVIDERS
		p, err := properties.LoadFile(securityFile, properties.UTF8)
		if err != nil {
			return fmt.Errorf("unable to read properties file %s\n%w", securityFile, err)
		}
		p = p.FilterStripPrefix("security.provider.")
		var providers []string
		for k, v := range p.Map() {
			providers = append(providers, fmt.Sprintf("%s|%s", k, v))
		}
		sort.Strings(providers)
		configCtx.Layer.LaunchEnvironment.Default("BPI_JVM_SECURITY_PROVIDERS", strings.Join(providers, " "))

		configCtx.Layer.LaunchEnvironment.Default("JAVA_HOME", configCtx.JavaHome)
		configCtx.Layer.LaunchEnvironment.Default("MALLOC_ARENA_MAX", "2")

		configCtx.Layer.LaunchEnvironment.Append("JAVA_TOOL_OPTIONS", " ", "-XX:+ExitOnOutOfMemoryError")
	}
	return nil
}

func (j JRE) Contribute(layer *libcnb.Layer) error {

	return j.LayerContributor.Contribute(layer, func(layer *libcnb.Layer, artifact *os.File) error {
		j.Logger.Bodyf("Expanding to %s", layer.Path)
		if err := crush.Extract(artifact, layer.Path, 1); err != nil {
			return fmt.Errorf("unable to expand JRE\n%w", err)
		}

		return ConfigureJRE(ConfigJREContext{
			Layer:             layer,
			Logger:            j.Logger,
			JavaHome:          layer.Path,
			JavaVersion:       j.LayerContributor.Dependency.Version,
			ApplicationPath:   j.ApplicationPath,
			IsBuild:           IsBuildContribution(j.Metadata),
			IsLaunch:          IsLaunchContribution(j.Metadata),
			SkipCerts:         false,
			CertificateLoader: j.CertificateLoader,
			DistType:          j.DistributionType,
		})
	})
}

func (j JRE) Name() string {
	return j.LayerContributor.LayerName()
}
