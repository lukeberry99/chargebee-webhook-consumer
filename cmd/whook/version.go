package main

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}
