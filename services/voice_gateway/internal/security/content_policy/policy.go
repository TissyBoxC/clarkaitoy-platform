// Package content_policy enforces child content rules.
package content_policy

// Decision is the result of one content policy check.
type Decision struct {
	Allowed bool
	Reason  string
}

// Policy checks input and output text.
type Policy interface {
	CheckText(text string) Decision
}
