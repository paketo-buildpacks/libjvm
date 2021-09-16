package helper

import (
	"fmt"
	"os"
	"strings"

	"github.com/paketo-buildpacks/libpak/bard"
)

type NMT struct {
	Logger bard.Logger
}

func (n NMT) Execute() (map[string]string, error) {

	if s, ok := os.LookupEnv("BPL_JAVA_NMT_ENABLED"); ok && strings.ToLower(s) == "false" {
		n.Logger.Info("Disabling Java Native Memory Tracking")
		return nil, nil
	}
	level := "summary"
	if s, ok := os.LookupEnv("BPL_JAVA_NMT_LEVEL"); ok && strings.ToLower(s) == "detail" {
		level = "detail"
	}

	n.Logger.Info("Enabling Java Native Memory Tracking")
	var values []string
	if s, ok := os.LookupEnv("JAVA_TOOL_OPTIONS"); ok {
		values = append(values, s)
	}
	values = append(values, "-XX:+UnlockDiagnosticVMOptions", fmt.Sprintf("-XX:NativeMemoryTracking=%s", level), "-XX:+PrintNMTStatistics")

	// NMT_LEVEL_1 Required for Java Native Memory Tracking to work due to bug which is not fixed until Java v18 (https://bugs.openjdk.java.net/browse/JDK-8256844)
	// '1' = PID of Java process in the container. Value for NMT level should match that passed to '-XX:NativeMemoryTracking' in the NMT helper.
	return map[string]string{"NMT_LEVEL_1": level, "JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil

}
