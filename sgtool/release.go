package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"src.sourcegraph.com/sourcegraph/dev/release"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

func init() {
	_, err := CLI.AddCommand("release",
		"release a new version to all users",
		"The release command releases a new Sourcegraph version by building, packaging, and uploading it.",
		&releaseCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

type ReleaseCmd struct {
	SkipPackage      bool `long:"skip-package" description:"skip package step (assumes it has already been run)"`
	SkipDistPackage  bool `long:"skip-dist-package" description:"skip create deb and rpm step (assumes it has already been run)"`
	InspectArtifacts bool `long:"inspect-artifacts" description:"avoids upload, but puts all artifacts in ./selfupdate"`
	Private          bool `long:"private" description:"do not upload src.json files which make this version the latest public version"`

	S3Dir string `long:"s3-dir" description:"S3 base directory to upload release to (default: src)"`

	PackageCmd
}

var releaseCmd ReleaseCmd

func (c *ReleaseCmd) Execute(args []string) error {
	// Check for dependencies before starting.
	if err := requireCmds("make", "go-selfupdate", "cp", "aws"); err != nil {
		return err
	}

	if !c.SkipPackage {
		if err := c.PackageCmd.Execute(nil); err != nil {
			return err
		}
	}
	if !c.SkipDistPackage {
		cmd := exec.Command("make", "package", "VERSION="+c.Args.Version)
		cmd.Dir = "./package"
		if err := execCmd(cmd); err != nil {
			return err
		}
	}

	var selfupdateDir string
	if c.InspectArtifacts {
		selfupdateDir = "selfupdate"
	} else {
		var err error
		selfupdateDir, err = ioutil.TempDir("", "selfupdate")
		if err != nil {
			return err
		}
		defer os.RemoveAll(selfupdateDir)
	}

	const releaseDir = "release"
	if err := execCmd(exec.Command("go-selfupdate", "-o="+selfupdateDir, "-cmd="+sgxcmd.Name, filepath.Join(releaseDir, c.Args.Version), c.Args.Version)); err != nil {
		return err
	}
	distDir := "package/dist/" + c.Args.Version
	if err := execCmd(exec.Command("cp", distDir+"/src.deb", distDir+"/src.rpm", selfupdateDir+"/"+c.Args.Version+"/linux-amd64/")); err != nil {
		return err
	}
	if err := execCmd(exec.Command("cp", distDir+"/src-docker.deb", selfupdateDir+"/"+c.Args.Version+"/linux-amd64/")); err != nil {
		return err
	}

	// Versions like "10.3.100" are considered public by default, while ones with
	// a dash suffix like "10.3.101-hack" are considered private by default.
	privateVersion := len(strings.Split(c.Args.Version, "-")) > 1
	if c.Private || privateVersion {
		matches, err := filepath.Glob(selfupdateDir + "/*/src.json")
		if err != nil {
			return err
		}
		for _, match := range matches {
			if err := os.Remove(match); err != nil {
				return err
			}
		}
	}

	if c.InspectArtifacts {
		return nil
	}

	if c.S3Dir == "" {
		c.S3Dir = release.S3Dir
	}

	syncCmd := exec.Command(
		"aws", "s3", "sync",
		"--acl", "public-read",
		selfupdateDir,
		"s3://"+release.S3Bucket+"/"+c.S3Dir,
	)
	if err := execCmd(syncCmd); err != nil {
		return err
	}

	return nil
}
