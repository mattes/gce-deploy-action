package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func formatLog(command string, params map[string]string, value string) string {
	if len(params) > 0 {
		p := []string{}
		for k, v := range params {
			p = append(p, fmt.Sprintf("%v=%v", k, v))
		}
		sort.Strings(p)
		return fmt.Sprintf("::%v %v::%v\n", command, strings.Join(p, ","), value)
	}

	return fmt.Sprintf("::%v::%v\n", command, value)
}

func Log(command string, params map[string]string, value string) {
	fmt.Fprint(os.Stdout, formatLog(command, params, value))
}

func LogSetEnv(name, value string) {
	Log("set-env", map[string]string{"name": name}, value)
}

func LogSetOutput(name, value string) {
	Log("set-output", map[string]string{"name": name}, value)
}

func LogAddSystemPath(path string) {
	Log("add-path", nil, path)
}

// You must create a secret named `ACTIONS_STEP_DEBUG` with the value `true`
// to see the debug messages set by this command in the log.
func LogDebug(message string, params map[string]string) {
	Log("debug", params, message)
}

func LogWarning(message string, params map[string]string) {
	Log("warning", params, message)
}

func LogError(message string, params map[string]string) {
	Log("error", params, message)
}

func LogAddMask(value string) {
	Log("add-mask", nil, value)
}

func Infof(format string, a ...interface{}) {
	elapsed := time.Now().Sub(startTime)

	fmt.Fprintf(os.Stdout, "%02d:%02d:%02d] %v\n",
		int(elapsed.Hours()), int(elapsed.Minutes()), int(elapsed.Seconds()),
		fmt.Sprintf(format, a...))
}

func Fatalf(format string, a ...interface{}) {
	LogError(fmt.Sprintf(format, a...), nil)
	os.Exit(1)
}
