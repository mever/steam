package cmd

import (
	"github.com/juju/utils/exec"
	"github.com/mever/steam"
	"os"
	"strings"
	"errors"
	exe "os/exec"
	"strconv"
)

type Client struct {
	SccDir         string
	AppsDir        string

	needsGuardCode bool
	AuthUser       string
	AuthPw         string
	AuthGuardCode  string
}

func (c *Client) checkConfig() {
	if c.SccDir == "" {
		c.SccDir = DefaultDir
	}

	if c.AppsDir == "" {
		c.AppsDir = DefaultAppsDir
	}
}

func (c *Client) Install(app steam.AppId) (err error) {
	c.checkConfig()

	_, err = os.Stat(c.SccDir)
	if os.IsNotExist(err) {
		err = c.installClient()
	}

	return c.installApp(app)
}

func (c *Client) installApp(app steam.AppId) error {
	appId := strconv.Itoa(int(app))
	gameDir := c.AppsDir + "/" + appId
	cmd := c.exec("+force_install_dir", gameDir, "+app_update", appId, "validate")
	return cmd.Run()
}

// exec Executes a command with the Steam Console Client
func (c *Client) exec(a ...string) *exe.Cmd {
	args := make([]string, 0, 10)
	if "" == c.AuthUser {
		args = append(args, "+login", "anonymous")
	} else {
		args = append(args, "+login", c.AuthUser, c.AuthPw)
	}
	args = append(args, a...)
	args = append(args, "+quit")

	cmd := exe.Command("./steamcmd.sh", args...)
	cmd.Dir = c.SccDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// installClient installs the Steam Console Client
func (c *Client) installClient() error {
	return c.runScript(

		// increase bash verbosity and safety
		"set -ux",

		// add 32 bit support to Ubuntu as SteamCMD is a 32 bit program
		"dpkg --add-architecture i386",
		"apt-get update",
		"apt-get install -y libc6:i386 libncurses5:i386 libstdc++6:i386",

		// get steam console client: SteamCMD
		"wget http://media.steampowered.com/client/steamcmd_linux.tar.gz",
		"tar -xvzf steamcmd_linux.tar.gz",
		"rm steamcmd_linux.tar.gz",
	)
}

// runScript executes the given lines in bash
func (c *Client) runScript(lines ...string) error {
	p := exec.RunParams{
		Commands: strings.Join(lines, "\n"),
	}

	os.MkdirAll(c.SccDir, 0755)
	p.WorkingDir = c.SccDir
	if err := p.Run(); err != nil {
		return err
	}

	if r, err := p.Wait(); err != nil {
		return err
	} else {
		os.Stderr.Write(r.Stderr)
		os.Stdout.Write(r.Stdout)
		if r.Code != 0 {
			return errors.New("Script failed")
		}

		return nil
	}
}