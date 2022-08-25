package observer

import (
	"bytes"
	"fmt"
)

//NewTextEvent ...
func NewTextEvent(msg string, args ...interface{}) TextMessageEvent {
	return TextMessageEvent{}.AddFmt(msg, args...)
}

//TextMessageEvent simple textual message
type TextMessageEvent struct {
	EventType
	parts []struct {
		message string
		args    []interface{}
	}
}

//AddFmt добавить строку с/без форматирования
func (p TextMessageEvent) AddFmt(s string, args ...interface{}) TextMessageEvent {
	p.parts = append(p.parts, struct {
		message string
		args    []interface{}
	}{message: s, args: args})
	return p
}

//String is a fmt.Stringer contract
func (p TextMessageEvent) String() string {
	buf := bytes.NewBuffer(nil)
	for _, item := range p.parts {
		_, _ = fmt.Fprintf(buf, item.message, item.args...)
	}
	return buf.String()
}
