package telegram_bot_framework

import "github.com/google/uuid"

const (
	AnsTypeText    AnswerType = iota
	AnsTypeVariant AnswerType = iota
)

type Prompt struct {
	id     string
	params PromptParams
}

type PromptParams struct {
	// Text is the prompt question that tg user sees
	Text string
	// AnsType is the expected type of prompt answer e.g. text, variant
	AnsType AnswerType
	// Variants contains predefined variants of answer to prompt. sent along with the prompt to user in case ansType is AnsTypeVariant
	Variants []string
	// Validators are executed in the order they were specified in the array 0,1,2...
	Validators []PromptValidator
	// SuccessMessage is shown in case input has passed all validators successfully
	SuccessMessage string
}

type AnswerType int

type PromptValidator struct {
	ErrorMessage string
	Validator    func(input string) bool
}

func NewPrompt(params PromptParams) *Prompt {
	return &Prompt{
		id:     uuid.New().String(),
		params: params,
	}
}

func (p *Prompt) Validate(input string) (string, bool) {
	for _, validator := range p.params.Validators {
		if validator.Validator(input) {
			continue
		}
		return validator.ErrorMessage, false
	}
	return "", true
}
