package notify

import (
	"github.com/Songmu/prompter"
)

// Notify interface asks confirmation
type Notify interface {
	AskQuestion(question string) (answer bool)
}

// CliQuestion implements notify
type CliQuestion struct {
}

// AskQuestion prompt a confirmation to the user
func (n CliQuestion) AskQuestion(question string) (answer bool) {
	return prompter.YesNo(question, false)
}
