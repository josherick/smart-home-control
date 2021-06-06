package logger

import (
	"fmt"
	"os"
	"time"
)

type dateFileWriterPrefix string

func NewFileWriter(prefix string) dateFileWriterPrefix {
	return dateFileWriterPrefix(prefix)
}

func (p dateFileWriterPrefix) Filename(t time.Time) string {
	return fmt.Sprintf(
		"%s-%d-%02d-%02d",
		p,
		t.Year(),
		t.Month(),
		t.Day(),
	)
}

func (p dateFileWriterPrefix) Append(text string) error {
	t := time.Now()
	filename := p.Filename(t)
	file, err := os.OpenFile(
		filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()
	line := fmt.Sprintf("[%s] %s\n", t.Format(time.RFC3339), text)
	_, err = file.WriteString(line)
	return err
}
