package cmd

import (
	"errors"
	"github.com/juju/utils/exec"
	"github.com/kr/pty"
	"github.com/mever/steam"
	"io"
	"os"
	exe "os/exec"
	"strings"
)

// Default Steam Console Client (SteamCMD) directory.
const DefaultDir = "/opt/steam/cmd"

// Default application directory.
const DefaultAppsDir = "/opt/steam/apps"

type Client struct {
	SccDir  string
	AppsDir string

	Stdout io.Writer
	Stderr io.Writer

	AuthUser string
	AuthPw   string
}

// completeConfig fills in the blanks for required parameters for SteamCMD
func (c *Client) completeConfig() {
	if c.SccDir == "" {
		c.SccDir = DefaultDir
	}

	if c.AppsDir == "" {
		c.AppsDir = DefaultAppsDir
	}

	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}

	if c.Stderr == nil {
		c.Stderr = os.Stderr
	}
}

// GetApp returns an installed Steam application. If it returns
// nil the app is not installed.
func (c *Client) GetApp(appId steam.AppId) *App {
	c.completeConfig()

	appDir := c.getAppDir(appId)
	var err error
	_, err = os.Stat(appDir)
	if os.IsNotExist(err) {
		return nil
	} else {
		return &App{dir: appDir}
	}
}

// InstallApp installs the app indicated by the provided Steam app id. When
// during the installation process steam has questions for you it will
// call the interviewer with a question.
func (c *Client) InstallApp(app steam.AppId, i steam.Interviewer) (err error) {
	c.completeConfig()

	_, err = os.Stat(c.SccDir)
	if os.IsNotExist(err) {
		err = c.installClient()
	}

	return c.installApp(app, i)
}

func (c *Client) newInterview(i steam.Interviewer) *interviewer {
	return &interviewer{w: c.Stdout, fn: i}
}

func (c *Client) getAppDir(app steam.AppId) string {
	return c.AppsDir + "/" + app.Id()
}

func (c *Client) installApp(app steam.AppId, i steam.Interviewer) error {
	appDir := c.getAppDir(app)
	cmd := c.buildCmd("+force_install_dir", appDir, "+app_update", app.Id(), "validate")

	tty, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	defer tty.Close()

	interview := c.newInterview(i)
	defer interview.fn("", false)

	interview.Run(tty)
	return cmd.Wait()
}

// buildCmd builds a command with the Steam Console Client
func (c *Client) buildCmd(a ...string) *exe.Cmd {
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
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
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
