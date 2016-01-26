package cmd

import (
	"github.com/juju/errors"
	"github.com/mever/steam"
	"sync"
)

type Question struct {
	Sensitive bool
	Message   string
}

type Installer struct {
	mu          sync.Mutex
	installing  bool
	interviewer Interviewer

	AnonymousLogin bool
	Questions      chan *Question
	Answers        chan string
}

var (
	ErrAlreadyInstalling = errors.New("We're aleady installing")
)

func (i *Installer) Installing() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.installing
}

func (i *Installer) Install(appId steam.AppId) error {
	i.mu.Lock()
	installing := i.installing
	if !installing {
		i.installing = true
	}
	i.mu.Unlock()
	if installing {
		return ErrAlreadyInstalling
	}

	if i.Questions == nil {
		i.Questions = make(chan *Question, 1)
	}
	if i.Answers == nil {
		i.Answers = make(chan string)
	}

	i.interviewer = func(q string, sensitive bool) string {
		if q == "" {
			close(i.Questions)
			close(i.Answers)
			i.mu.Lock()
			i.installing = false
			i.mu.Unlock()
			return ""
		} else {
			i.Questions <- &Question{Message: q, Sensitive: sensitive}
		}
		return <-i.Answers
	}

	go i.run(appId)
	return nil
}

func (i *Installer) run(appId steam.AppId) {
	c := Client{}
	if !i.AnonymousLogin {
		c.AuthUser = i.interviewer("What is your Steam username?", false)
		c.AuthPw = i.interviewer("What is your Steam password?", true)
	}

	c.InstallApp(appId, i.interviewer)
}
