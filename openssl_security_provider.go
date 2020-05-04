/*
 * Copyright 2018-2020, VMware, Inc. All Rights Reserved.
 * Proprietary and Confidential.
 * Unauthorized use, copying or distribution of this source code via any medium is
 * strictly prohibited without the express written consent of VMware, Inc.
 */

package libjvm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type OpenSSLSecurityProvider struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
	Metadata         map[string]interface{}
}

func NewOpenSSLSecurityProvider(dependency libpak.BuildpackDependency, cache libpak.DependencyCache,
	metadata map[string]interface{}, plan *libcnb.BuildpackPlan) OpenSSLSecurityProvider {

	return OpenSSLSecurityProvider{
		LayerContributor: libpak.NewDependencyLayerContributor(dependency, cache, plan),
		Metadata:         metadata,
	}
}

//go:generate statik -src . -include *.sh

func (o OpenSSLSecurityProvider) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	o.LayerContributor.Logger = o.Logger

	return o.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		o.Logger.Bodyf("Copying to %s", layer.Path)

		file := filepath.Join(layer.Path, filepath.Base(artifact.Name()))
		if err := sherpa.CopyFile(artifact, file); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy %s to %s\n%w", artifact.Name(), file, err)
		}

		layer.SharedEnvironment.Append("SECURITY_PROVIDERS", " 2|io.paketo.openssl.OpenSslProvider")
		layer.SharedEnvironment.PrependPath("SECURITY_PROVIDERS_CLASSPATH", file)

		s, err := sherpa.StaticFile("/openssl-security-provider.sh")
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load memory-calculator.sh\n%w", err)
		}

		layer.Profile.Add("openssl-security-provider.sh", s)

		if isBuildContribution(o.Metadata) {
			layer.Build = true
			layer.Cache = true
		}

		if isLaunchContribution(o.Metadata) {
			layer.Launch = true
		}

		return layer, nil
	})
}

func (OpenSSLSecurityProvider) Name() string {
	return "openssl-security-provider"
}
