package domain

import "fmt"

// ErrorKind unterscheidet fachliche Fehler, damit der Handler passende
// HTTP-Statuscodes wählen kann.
type ErrorKind int

const (
	KindInvalid ErrorKind = iota + 1
	KindNotFound
	KindConflict
)

// Error ist ein fachlicher Fehler mit einer für Benutzer lesbaren Meldung.
type Error struct {
	Kind ErrorKind
	Msg  string
}

func (e *Error) Error() string { return e.Msg }

func Invalid(format string, args ...any) error {
	return &Error{Kind: KindInvalid, Msg: fmt.Sprintf(format, args...)}
}

func NotFound(format string, args ...any) error {
	return &Error{Kind: KindNotFound, Msg: fmt.Sprintf(format, args...)}
}

func Conflict(format string, args ...any) error {
	return &Error{Kind: KindConflict, Msg: fmt.Sprintf(format, args...)}
}
