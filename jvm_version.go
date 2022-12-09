package libjvm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type JVMVersion struct {
	Logger bard.Logger
}

func NewJVMVersion(logger bard.Logger) JVMVersion {
	return JVMVersion{Logger: logger}
}

func (j JVMVersion) ResolveMetadataVersion(planEntries ...libcnb.BuildpackPlanEntry) (string, error) {
	version := ""

	for _, entry := range planEntries {
		requiredVersion, exists := entry.Metadata["version"]
		if exists {
			v := requiredVersion.(string)
			j.Logger.Body(faint.Sprintf("Found required Java version %s for %s", v, entry.Name))

			if version == "" {
				version = v
			} else if version != v {
				return "", fmt.Errorf("Buildplan requires conflicting Java versions %s!=%s", version, v)
			}
		}
	}

	return version, nil
}

var faint = color.New(color.Faint)

func (j JVMVersion) GetJVMVersion(appPath string, cr libpak.ConfigurationResolver, metadataVersion string) (string, error) {
	version, explicit := cr.Resolve("BP_JVM_VERSION")
	if explicit {
		j.Logger.Body(faint.Sprintf("Using Java version %s from BP_JVM_VERSION", version))
		return version, nil
	}

	if metadataVersion != "" {
		j.Logger.Body(faint.Sprintf("Using Java version %s from metadata", metadataVersion))
		return metadataVersion, nil
	}

	sdkmanrcJavaVersion, err := readJavaVersionFromSDKMANRCFile(appPath)
	if err != nil {
		return "", fmt.Errorf("unable to read Java version from SDMANRC file\n%w", err)
	}

	if len(sdkmanrcJavaVersion) > 0 {
		sdkmanrcJavaMajorVersion := extractMajorVersion(sdkmanrcJavaVersion)
		j.Logger.Body(faint.Sprintf("Using Java version %s extracted from .sdkmanrc", sdkmanrcJavaMajorVersion))
		return sdkmanrcJavaMajorVersion, nil
	}

	mavenJavaVersion, err := readJavaVersionFromMavenMetadata(appPath)
	if err != nil {
		return "", fmt.Errorf("unable to read Java version from Maven metadata\n%w", err)
	}

	if len(mavenJavaVersion) > 0 {
		mavenJavaMajorVersion := extractMajorVersion(mavenJavaVersion)
		j.Logger.Body(faint.Sprintf("Using Java version %s extracted from MANIFEST.MF", mavenJavaMajorVersion))
		return mavenJavaMajorVersion, nil
	}

	j.Logger.Body(faint.Sprintf("Using buildpack default Java version %s", version))
	return version, nil
}

func readJavaVersionFromSDKMANRCFile(appPath string) (string, error) {
	components, err := ReadSDKMANRC(filepath.Join(appPath, ".sdkmanrc"))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	for _, component := range components {
		if component.Type == "java" {
			return component.Version, nil
		}
	}

	return "", nil
}

func readJavaVersionFromMavenMetadata(appPath string) (string, error) {
	manifest, err := NewManifest(appPath)
	if err != nil {
		return "", err
	}

	javaVersion, ok := manifest.Get("Build-Jdk-Spec")
	if !ok {
		javaVersion, _ = manifest.Get("Build-Jdk")
	}

	return javaVersion, nil
}

func extractMajorVersion(version string) string {
	versionParts := strings.Split(version, ".")

	if versionParts[0] == "1" {
		return versionParts[1]
	}

	return versionParts[0]
}
