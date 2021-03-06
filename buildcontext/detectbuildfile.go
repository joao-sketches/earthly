package buildcontext

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/earthly/earthly/domain"
	gwclient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/pkg/errors"
)

// detectBuildFile detects whether to use Earthfile, build.earth or Dockerfile.
func detectBuildFile(target domain.Target, localDir string) (string, error) {
	if target.Target == DockerfileMetaTarget {
		return filepath.Join(localDir, "Dockerfile"), nil
	}
	earthfilePath := filepath.Join(localDir, "Earthfile")
	_, err := os.Stat(earthfilePath)
	if os.IsNotExist(err) {
		buildEarthPath := filepath.Join(localDir, "build.earth")
		_, err := os.Stat(buildEarthPath)
		if os.IsNotExist(err) {
			return "", fmt.Errorf(
				"No Earthfile nor build.earth file found for target %s", target.String())
		} else if err != nil {
			return "", errors.Wrapf(err, "stat file %s", buildEarthPath)
		}
		return buildEarthPath, nil
	} else if err != nil {
		return "", errors.Wrapf(err, "stat file %s", earthfilePath)
	}
	return earthfilePath, nil
}

func detectBuildFileInRef(ctx context.Context, target domain.Target, ref gwclient.Reference, subDir string) (string, error) {
	if target.Target == DockerfileMetaTarget {
		return path.Join(subDir, "Dockerfile"), nil
	}
	earthfilePath := path.Join(subDir, "Earthfile")
	exists, err := fileExists(ctx, ref, earthfilePath)
	if err != nil {
		return "", err
	}
	if exists {
		return earthfilePath, nil
	}
	buildEarthPath := path.Join(subDir, "build.earth")
	exists, err = fileExists(ctx, ref, buildEarthPath)
	if err != nil {
		return "", err
	}
	if exists {
		return buildEarthPath, nil
	}
	return "", errors.Errorf("no build file found in %s", subDir)
}

func fileExists(ctx context.Context, ref gwclient.Reference, fpath string) (bool, error) {
	dir, file := path.Split(fpath)
	fstats, err := ref.ReadDir(ctx, gwclient.ReadDirRequest{
		Path:           dir,
		IncludePattern: file,
	})
	if err != nil {
		return false, errors.Wrapf(err, "cannot read dir %s", dir)
	}
	for _, fstat := range fstats {
		name := path.Base(fstat.GetPath())
		if name == file && !fstat.IsDir() {
			return true, nil
		}
	}
	return false, nil
}
