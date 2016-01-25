package cmd

import (
	"os"
)

type App struct {
	dir string
}

// Remove the Steam application.
func (a *App) Remove() error {
	return os.RemoveAll(a.dir)
}