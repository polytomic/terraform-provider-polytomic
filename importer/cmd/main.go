package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	Execute(buildVersion(version, commit, date, builtBy))
}

func buildVersion(version, commit, date, builtBy string) string {
	result := version + ","
	if commit != "" {
		result = fmt.Sprintf("%s commit: %s,", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s built at: %s,", result, date)
	}
	if builtBy != "" {
		result = fmt.Sprintf("%s built by: %s,", result, builtBy)
	}
	result = fmt.Sprintf("%s goos: %s, goarch: %s", result, runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s module version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}
