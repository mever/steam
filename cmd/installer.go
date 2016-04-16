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
	running     bool
	interviewer Interviewer

	AnonymousLogin bool
	Questions      chan *Question
	Answers        chan string
}

var (
	ErrAlreadyRunning = errors.New("We're aleady running")
)

func (i *Installer) Running() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.running
}

func (i *Installer) Install(appId steam.AppId) error {
	if err := i.start(); err == nil {
		go i.getClient().InstallApp(appId, i.interviewer)
		return nil
	} else {
		return err
	}
}

func (i *Installer) Update(app *App) (err error) {
	if err := i.start(); err == nil {
		go i.getClient().UpdateApp(app, i.interviewer)
		return nil
	} else {
		return err
	}
}

func (i Installer) start() error {
	i.mu.Lock()
	running := i.running
	if !running {
		i.running = true
	}
	i.mu.Unlock()
	if running {
		return ErrAlreadyRunning
	} else {
		return nil
	}
}

func (i *Installer) getClient() *Client {
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
			i.running = false
			i.mu.Unlock()
			return ""
		} else {
			i.Questions <- &Question{Message: q, Sensitive: sensitive}
		}
		return <-i.Answers
	}

	c := Client{}
	if !i.AnonymousLogin {
		c.AuthUser = i.interviewer("What is your Steam username?", false)
		c.AuthPw = i.interviewer("What is your Steam password?", true)
	}

	return &c
}
