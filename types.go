package steam

import (
	"strconv"
)

func AppIdFromString(id string) AppId {
	i, _ := strconv.Atoi(id)
	return AppId(i)
}

type AppId int

func (a AppId) Int() int {
	return int(a)
}

func (a AppId) Id() string {
	return strconv.Itoa(int(a))
}

// Interviewer asks a question and expects an answer
type Interviewer func(question string, sensitive bool) string
