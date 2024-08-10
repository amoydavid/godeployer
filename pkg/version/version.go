package version

import (
	"fmt"
	"runtime"
)

var (
	// Version 是应用程序的语义版本
	Version = "0.1.0"

	// GitCommit 是构建时的 Git commit hash
	GitCommit = "unknown"

	// BuildDate 是构建的日期和时间
	BuildDate = "unknown"
)

// Info 包含版本信息的结构体
type Info struct {
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
	OS        string
	Arch      string
}

// GetVersionInfo 返回完整的版本信息
func GetVersionInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String 返回格式化的版本信息字符串
func (i Info) String() string {
	return fmt.Sprintf("Version: %s\nGit Commit: %s\nBuild Date: %s\nGo Version: %s\nOS/Arch: %s/%s",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.OS, i.Arch)
}
