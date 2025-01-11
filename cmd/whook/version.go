package main

import "github.com/lukeberry99/whook/internal/version"

func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version: version.Version,
		Commit:  version.Commit,
		Date:    version.Date,
	}
}

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}
