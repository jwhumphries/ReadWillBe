// Package views holds shared types and templ-rendered page templates used
// by the readwillbe HTTP handlers.
package views

// ManualReading is a single row in the manual plan-creation draft form.
type ManualReading struct {
	ID      string
	Date    string
	Content string
}
