package version

import (
	"log"

	"github.com/Masterminds/semver/v3"
)

type versionService struct{}

func NewVersionService() Version {
	return &versionService{}
}

// CompareVersions implements Version.
func (v *versionService) CompareVersions(remoteVersion string) bool {
	_newVersion, err := semver.StrictNewVersion(remoteVersion)
	if err != nil {
		log.Printf("ERROR: invalid remote version %s - %s", remoteVersion, err.Error())
		return false
	}

	_curVersion, err := semver.StrictNewVersion(CurrentVersion)
	if err != nil {
		log.Printf("ERROR: invalid current version %s - %s", CurrentVersion, err.Error())
		return false
	}
	// log.Printf("\nCurrent: v%s \nRemote: v%s\nLatest?: %d\n", _curVersion.String(), _newVersion.String(), _newVersion.Compare(_curVersion))

	return _newVersion.GreaterThan(_curVersion)
}
