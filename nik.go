package libjvm

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

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb/v2"
	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak/v2"
	"github.com/paketo-buildpacks/libpak/v2/bard"
	"github.com/paketo-buildpacks/libpak/v2/crush"
	"github.com/paketo-buildpacks/libpak/v2/effect"
)

type NIK struct {
	CertificateLoader CertificateLoader
	DependencyCache   libpak.DependencyCache
	Executor          effect.Executor
	JDKDependency     libpak.BuildModuleDependency
	LayerContributor  libpak.LayerContributor
	Logger            bard.Logger
	NativeDependency  *libpak.BuildModuleDependency
	CustomCommand     string
	CustomArgs        []string
}

func NewNIK(jdkDependency libpak.BuildModuleDependency, nativeDependency *libpak.BuildModuleDependency, cache libpak.DependencyCache, certificateLoader CertificateLoader, customCommand string, customArgs []string) (NIK, error) {
	dependencies := []libpak.BuildModuleDependency{jdkDependency}

	if nativeDependency != nil {
		dependencies = append(dependencies, *nativeDependency)
	}

	expected := map[string]interface{}{"dependencies": dependencies}

	if md, err := certificateLoader.Metadata(); err != nil {
		return NIK{}, fmt.Errorf("unable to generate certificate loader metadata")
	} else {
		for k, v := range md {
			expected[k] = v
		}
	}

	contributor := libpak.NewLayerContributor(
		bard.FormatIdentity(jdkDependency.Name, jdkDependency.Version),
		expected,
		libcnb.LayerTypes{
			Build: true,
			Cache: true,
		},
	)
	n := NIK{
		CertificateLoader: certificateLoader,
		DependencyCache:   cache,
		Executor:          effect.NewExecutor(),
		JDKDependency:     jdkDependency,
		NativeDependency:  nativeDependency,
		LayerContributor:  contributor,
		CustomCommand:     customCommand,
		CustomArgs:        customArgs,
	}

	return n, nil
}

func (n NIK) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	n.LayerContributor.Logger = n.Logger

	return n.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		artifact, err := n.DependencyCache.Artifact(n.JDKDependency)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to get dependency %s\n%w", n.JDKDependency.ID, err)
		}
		defer artifact.Close()

		n.Logger.Bodyf("Expanding to %s", layer.Path)
		if err := crush.Extract(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand JDK\n%w", err)
		}

		layer.BuildEnvironment.Override("JAVA_HOME", layer.Path)
		layer.BuildEnvironment.Override("JDK_HOME", layer.Path)

		var keyStorePath string
		if IsBeforeJava9(n.JDKDependency.Version) {
			keyStorePath = filepath.Join(layer.Path, "jre", "lib", "security", "cacerts")
		} else {
			keyStorePath = filepath.Join(layer.Path, "lib", "security", "cacerts")
		}
		if err := os.Chmod(keyStorePath, 0664); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to set keystore file permissions\n%w", err)
		}

		if IsBeforeJava18(n.JDKDependency.Version) {
			if err := n.CertificateLoader.Load(keyStorePath, "changeit"); err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to load certificates\n%w", err)
			}
		} else {
			n.Logger.Bodyf("%s: The JVM cacerts entries cannot be loaded with Java 18+, for more information see: https://github.com/paketo-buildpacks/libjvm/issues/158", color.YellowString("Warning"))
		}

		if n.NativeDependency != nil {
			n.Logger.Header(color.BlueString("%s %s", n.NativeDependency.Name, n.NativeDependency.Version))

			artifact, err := n.DependencyCache.Artifact(*n.NativeDependency)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to get dependency %s\n%w", n.NativeDependency.ID, err)
			}
			defer artifact.Close()

			n.Logger.Body("Installing substrate VM")

			n.CustomArgs = append(n.CustomArgs, artifact.Name())

			if err := n.Executor.Execute(effect.Execution{
				Command: filepath.Join(layer.Path, n.CustomCommand),
				Args:    n.CustomArgs,
				Dir:     layer.Path,
				Stdout:  n.Logger.DebugWriter(),
				Stderr:  n.Logger.DebugWriter(),
			}); err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to run custom NIK command\n%w", err)
			}
		}

		return layer, nil
	})
}

func (NIK) Name() string {
	return "nik"
}
