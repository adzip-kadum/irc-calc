package version

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

var (
	Project   string = "test"
	Version   string
	GitCommit string
	GitBranch string
	BuildTS   string
	Semver    *semver.Version
)

func init() {
	if Project == "" {
		panic("version.Project is empty")
	}

	if Version == "" {
		Version = "v0.0.0-unknown"
	} else if Version[0] != 'v' {
		Version = fmt.Sprintf("v0.0.0-%s", Version)
	}

	var err error
	if Semver, err = semver.NewVersion(Version); err != nil {
		panic(errors.Wrap(err, Version))
	}
}
