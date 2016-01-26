package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
)

// Interviewer asks a question and expects an answer
type Interviewer func(question string, sensitive bool) string

type registeredQuestion struct {
	value     string
	sensitive bool
	regex     *regexp.Regexp
}

var (
	questionsMu sync.Mutex
	questions   = make([]registeredQuestion, 0)
)

// AddQuestion registers a question the SteamCMD utility may ask
// to the user. The provided regex is used to match the output line
// from the SteamCMD utility, the regex must compile (causing a panic otherwise).
func AddQuestion(regex, q string, sensitive bool) {
	questionsMu.Lock()
	defer questionsMu.Unlock()
	questions = append(questions, registeredQuestion{
		value:     q,
		sensitive: sensitive,
		regex:     regexp.MustCompile(regex),
	})
}

type interviewer struct {
	fn Interviewer
	w  io.Writer
}

// getQ matches the given bytes b to added questions
// when matched the corresponding question is returned
// else nil is returned
func (i *interviewer) getQ(b []byte) *registeredQuestion {
	questionsMu.Lock()
	defer questionsMu.Unlock()
	for _, q := range questions {
		if len(q.regex.Find(b)) > 0 {
			return &q
		}
	}
	return nil
}

// Run starts an infinite loop reading from tty
// and matching read buffers to the list of questions
// writing back the answers.
// One important thing though; if a matched line
// spans multiple buffers it is not matched!
func (i *interviewer) Run(tty *os.File) {
	go func() {
		buf := make([]byte, 32*1024)
		for {
			nr, er := tty.Read(buf)
			if nr > 0 {
				//				fmt.Println(" <" + string(buf[0:nr]) + "> ")
				if q := i.getQ(buf[0:nr]); q != nil {
					tty.Write([]byte(i.fn(q.value, q.sensitive) + "\n"))
				}
			}
			if er == io.EOF {
				break
			}
			if er != nil {
				fmt.Println(er)
				break
			}
		}
	}()
}
