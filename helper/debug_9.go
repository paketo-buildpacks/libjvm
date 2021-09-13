package helper

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/paketo-buildpacks/libpak/bard"
)

type Debug9 struct {
	Logger bard.Logger
}

func (d Debug9) Execute() (map[string]string, error) {
	if _, ok := os.LookupEnv("BPL_DEBUG_ENABLED"); !ok {
		return nil, nil
	}

	var err error

	port := "*:8000" // Java 9+ address format
	if s, ok := os.LookupEnv("BPL_DEBUG_PORT"); ok {
		port = "*:" + s
	}

	suspend := false
	if s, ok := os.LookupEnv("BPL_DEBUG_SUSPEND"); ok {
		suspend, err = strconv.ParseBool(s)
		if err != nil {
			return nil, fmt.Errorf("unable to parse $BPL_DEBUG_SUSPEND\n%w", err)
		}
	}

	s := fmt.Sprintf("Debugging enabled on port %s", port)
	if suspend {
		s = fmt.Sprintf("%s, suspended on start", s)
	}
	d.Logger.Info(s)

	var values []string
	if s, ok := os.LookupEnv("JAVA_TOOL_OPTIONS"); ok {
		values = append(values, s)
	}

	if suspend {
		s = "y"
	} else {
		s = "n"
	}

	values = append(values,
		fmt.Sprintf("-agentlib:jdwp=transport=dt_socket,server=y,address=%s,suspend=%s", port, s))

	return map[string]string{"JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil
}
