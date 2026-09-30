// Package domain contains notification domain types.
package domain

// Message describes one outbound notification.
type Message struct {
	Recipient string
	Channel   string
	Body      string
}
