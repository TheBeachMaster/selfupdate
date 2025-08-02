package version

var CurrentVersion string = "0.1.0"

type Version interface {
	CompareVersions(remoteVersion string) bool
}
