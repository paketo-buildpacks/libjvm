package helper

import (
	"fmt"
	"os"
	"strings"

	"github.com/paketo-buildpacks/libpak/bard"
)

type JMX struct {
	Logger bard.Logger
}

func (j JMX) Execute() (map[string]string, error) {
	if _, ok := os.LookupEnv("BPL_JMX_ENABLED"); !ok {
		return nil, nil
	}

	port := "5000"
	if s, ok := os.LookupEnv("BPL_JMX_PORT"); ok {
		port = s
	}

	j.Logger.Infof("JMX enabled on port %s", port)

	var values []string
	if s, ok := os.LookupEnv("JAVA_TOOL_OPTIONS"); ok {
		values = append(values, s)
	}

	values = append(values,
		"-Djava.rmi.server.hostname=127.0.0.1",
		"-Dcom.sun.management.jmxremote.authenticate=false",
		"-Dcom.sun.management.jmxremote.ssl=false",
		fmt.Sprintf("-Dcom.sun.management.jmxremote.port=%s", port),
		fmt.Sprintf("-Dcom.sun.management.jmxremote.rmi.port=%s", port),
	)

	return map[string]string{"JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil
}
