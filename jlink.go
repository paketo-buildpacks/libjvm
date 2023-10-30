package libjvm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/heroku/color"
	"github.com/magiconair/properties"
	"github.com/paketo-buildpacks/libjvm/count"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/effect"
)

type JLink struct {
	LayerContributor  libpak.LayerContributor
	Logger            bard.Logger
	ApplicationPath   string
	Executor          effect.Executor
	CertificateLoader CertificateLoader
	Metadata          map[string]interface{}
	JavaVersion       string
	Args              []string
	UserConfigured    bool
}

func NewJLink(applicationPath string, exec effect.Executor, args []string, certificateLoader CertificateLoader, metadata map[string]interface{}, userConfigured bool) (JLink, error) {
	expected := map[string]interface{}{"jlink-args": args}
	if md, err := certificateLoader.Metadata(); err != nil {
		return JLink{}, fmt.Errorf("unable to generate certificate loader metadata\n%w", err)
	} else {
		for k, v := range md {
			expected[k] = v
		}
	}
	contributor := libpak.NewLayerContributor(
		"JLink",
		expected,
		libcnb.LayerTypes{
			Build:  IsBuildContribution(metadata),
			Cache:  IsBuildContribution(metadata),
			Launch: IsLaunchContribution(metadata),
		})

	return JLink{
		LayerContributor:  contributor,
		Executor:          exec,
		CertificateLoader: certificateLoader,
		Metadata:          metadata,
		ApplicationPath:   applicationPath,
		Args:              args,
		UserConfigured:    userConfigured,
	}, nil
}

func (j JLink) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	j.LayerContributor.Logger = j.Logger

	return j.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {

		if err := os.RemoveAll(layer.Path); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to remove jlink layer dir \n%w", err)
		}

		valid := true
		if j.UserConfigured {
			valid = j.validArgs()
		}
		if !j.UserConfigured || !valid {
			modules, err := j.listJVMModules(layer.Path)
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to retrieve list of JVM modules for jlink\n%w", err)
			}
			j.Args = append(j.Args, "--add-modules", modules)
		}

		j.Args = append(j.Args, "--output", layer.Path)
		if err := j.buildCustomJRE(layer.Path); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to build custom JRE with jlink \n%w", err)
		}

		cacertsPath := filepath.Join(layer.Path, "lib", "security", "cacerts")
		if err := os.Chmod(cacertsPath, 0664); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to set keystore file permissions\n%w", err)
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

			file := filepath.Join(layer.Path, "conf", "security", "java.security")

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

func (j JLink) Name() string {
	return j.LayerContributor.Name
}

func (j *JLink) buildCustomJRE(layerPath string) error {

	if err := j.Executor.Execute(effect.Execution{
		Command: filepath.Join(filepath.Dir(layerPath), "jdk", "bin", "jlink"),
		Args:    j.Args,
		Stdout:  j.Logger.Logger.InfoWriter(),
		Stderr:  j.Logger.Logger.InfoWriter(),
	}); err != nil {
		return fmt.Errorf("unable to run jlink\n%w", err)
	}
	return nil
}

func (j *JLink) validArgs() bool {
	jlinkArgs := j.Args[:0]
	var skipNext, modsFound bool
	for _, original := range j.Args {
		if skipNext {
			skipNext = false
			continue
		}
		a := strings.ToLower(original)
		if strings.HasPrefix(a, "--output") {
			j.Logger.Bodyf(color.New(color.Faint, color.Bold).Sprint("WARNING: explicitly specified '--output' option & value will be overridden"))
			skipNext = true
			continue
		}
		if strings.HasPrefix(a, "--add-modules") {
			modsFound = true
		}
		jlinkArgs = append(jlinkArgs, original)
	}
	j.Args = jlinkArgs
	if !modsFound {
		j.Logger.Bodyf(color.New(color.Faint, color.Bold).Sprint("WARNING: jlink args provided but no modules specified, default JVM modules will be added"))
	}
	return modsFound
}

func (j *JLink) listJVMModules(layerPath string) (string, error) {
	var mods []string
	buf := &bytes.Buffer{}
	if err := j.Executor.Execute(effect.Execution{
		Command: filepath.Join(filepath.Dir(layerPath), "jdk", "bin", "java"),
		Args:    []string{"--list-modules"},
		Stdout:  buf,
		Stderr:  j.Logger.Logger.InfoWriter(),
	}); err != nil {
		return "", fmt.Errorf("unable to list modules\n%w", err)
	}
	m := strings.Split(strings.TrimSpace(buf.String()), "\n")
	for _, mod := range m {
		if strings.HasPrefix(mod, "java.") {
			if strings.Contains(mod, "@") {
				mod = strings.Split(mod, "@")[0]
			}
			mods = append(mods, mod)
		}
	}
	modList := strings.Join(mods, ",")
	return modList, nil
}
