package eventlog

// DEPRECATED: will soon be removed
type Log interface {
	Append(events []Event) error
}
