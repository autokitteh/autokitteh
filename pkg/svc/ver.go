package svc

import "fmt"

type Version struct{ Version, Commit, Date string }

func (v Version) String() string {
	return fmt.Sprintf("%s %s %s", v.Version, v.Commit, v.Date)
}

func (v Version) Any() bool { return v.Version != "" || v.Commit != "" || v.Date != "" }

var version *Version

func SetVersion(v *Version) {
	if v == nil || !v.Any() {
		version = nil
		return
	}

	version = v
}

func GetVersion() *Version { return version }
