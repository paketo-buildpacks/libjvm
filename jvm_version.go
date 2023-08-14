package libjvm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak/v2"
	"github.com/paketo-buildpacks/libpak/v2/bard"
)

type JVMVersion struct {
	Logger bard.Logger
}

func NewJVMVersion(logger bard.Logger) JVMVersion {
	return JVMVersion{Logger: logger}
}

func (j JVMVersion) GetJVMVersion(appPath string, cr libpak.ConfigurationResolver) (string, error) {
	version, explicit := cr.Resolve("BP_JVM_VERSION")
	if explicit {
		f := color.New(color.Faint)
		j.Logger.Body(f.Sprintf("Using Java version %s from BP_JVM_VERSION", version))
		return version, nil
	}

	sdkmanrcJavaVersion, err := readJavaVersionFromSDKMANRCFile(appPath)
	if err != nil {
		return "", fmt.Errorf("unable to read Java version from SDMANRC file\n%w", err)
	}

	if len(sdkmanrcJavaVersion) > 0 {
		sdkmanrcJavaMajorVersion := extractMajorVersion(sdkmanrcJavaVersion)
		f := color.New(color.Faint)
		j.Logger.Body(f.Sprintf("Using Java version %s extracted from .sdkmanrc", sdkmanrcJavaMajorVersion))
		return sdkmanrcJavaMajorVersion, nil
	}

	mavenJavaVersion, err := readJavaVersionFromMavenMetadata(appPath)
	if err != nil {
		return "", fmt.Errorf("unable to read Java version from Maven metadata\n%w", err)
	}

	if len(mavenJavaVersion) > 0 {
		mavenJavaMajorVersion := extractMajorVersion(mavenJavaVersion)
		f := color.New(color.Faint)
		j.Logger.Body(f.Sprintf("Using Java version %s extracted from MANIFEST.MF", mavenJavaMajorVersion))
		return mavenJavaMajorVersion, nil
	}

	f := color.New(color.Faint)
	j.Logger.Body(f.Sprintf("Using buildpack default Java version %s", version))
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
