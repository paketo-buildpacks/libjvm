package libjvm

import (
	"strings"

	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type JVMVersion struct {
	Logger bard.Logger
}

func (jvmVersion JVMVersion) GetJVMVersion(appPath string, cr libpak.ConfigurationResolver) (string, error) {
	version, explicit := cr.Resolve("BP_JVM_VERSION")

	if !explicit {
		manifest, err := NewManifest(appPath)
		if err != nil {
			return version, err
		}

		javaVersion := ""

		buildJdkSpecVersion, ok := manifest.Get("Build-Jdk-Spec")
		if ok {
			javaVersion = buildJdkSpecVersion
		}

		buildJdkVersion, ok := manifest.Get("Build-Jdk")
		if ok {
			javaVersion = buildJdkVersion
		}

		if len(javaVersion) > 0 {
			javaVersionFromMaven := extractMajorVersion(javaVersion)
			f := color.New(color.Faint)
			jvmVersion.Logger.Body(f.Sprintf("Using Java version %s extracted from MANIFEST.MF", javaVersionFromMaven))
			return javaVersionFromMaven, nil
		}
	}

	return version, nil
}

func extractMajorVersion(version string) string {
	versionParts := strings.Split(version, ".")

	if versionParts[0] == "1" {
		return versionParts[1]
	}

	return versionParts[0]
}
