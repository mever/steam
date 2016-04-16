package cmd

import (
	"github.com/mever/steam"
	"os"
)

type App struct {
	id  steam.AppId
	dir string
}

// Remove the Steam application.
func (a *App) Remove() error {
	return os.RemoveAll(a.dir)
}
