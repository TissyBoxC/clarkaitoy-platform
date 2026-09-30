// Package prompt_guard protects model prompts.
package prompt_guard

// Guard sanitizes or rejects unsafe model prompts.
type Guard interface {
	Apply(systemPrompt string, userInput string) (string, error)
}
