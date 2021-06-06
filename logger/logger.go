package logger

import (
	"fmt"
	"net/http"
)

type writer interface {
	Append(text string) error
}

type emailer interface {
	Email(subject string, body string) error
}

type logger struct {
	systemName   string
	infoWriter   writer
	accessWriter writer
	emailer      emailer
}

func NewLogger(
	systemName string,
	infoWriter writer,
	accessWriter writer,
	emailer emailer,
) *logger {
	return &logger{systemName, infoWriter, accessWriter, emailer}
}

func (l *logger) Info(message string) {
	l.infoWriter.Append(message)
}

func (l *logger) Critical(message string) {
	subject := fmt.Sprintf("Critical message from %s", l.systemName)
	l.infoWriter.Append(fmt.Sprintf("%s: %s", subject, message))
	l.emailer.Email(subject, message)
}

func (l *logger) CriticalError(err error) {
	subject := fmt.Sprintf("Critical error from %s", l.systemName)
	l.infoWriter.Append(fmt.Sprintf("%s: %s", subject, err.Error()))
	l.emailer.Email(subject, err.Error())
}

func (l *logger) WrapHandlerWithAccessLog(mux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.access(r.RemoteAddr, r.RequestURI, r.Method)
		mux.ServeHTTP(w, r)
	})
}

func (l *logger) access(addr string, uri string, method string) {
	l.accessWriter.Append(fmt.Sprintf("%s %s %s", method, addr, uri))
}
