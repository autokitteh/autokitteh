package pythonrt

import (
	"encoding/json"
	"fmt"
	"net"
)

// Messages should be in sync with MessageType in ak_runner.py

// py <-> go
type CallbackMessage struct {
	Name string            `json:"name"`
	Args []string          `json:"args"`
	Kw   map[string]string `json:"kw"`
	Data []byte            `json:"data"`
}

func (CallbackMessage) Type() string {
	return "callback"
}

// There's no data in the done message

// py -> go
type ModuleMessage struct {
	Entries []string `json:"entries"`
}

func (ModuleMessage) Type() string { return "module" }

// py <-> go
type ResponseMessage struct {
	Value []byte `json:"value"`
}

func (ResponseMessage) Type() string { return "response" }

// go -> python
type RunMessage struct {
	FuncName string         `json:"func_name"`
	Event    map[string]any `json:"event"`
}

func (RunMessage) Type() string { return "run" }

type DoneMessage struct{}

func (DoneMessage) Type() string { return "done" }

type LogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func (LogMessage) Type() string { return "log" }

type Typed interface {
	Type() string
}

type SubMessage interface {
	CallbackMessage |
		DoneMessage |
		LogMessage |
		ModuleMessage |
		ResponseMessage |
		RunMessage

	Typed
}

func messageType[T SubMessage]() string {
	var m T
	return m.Type()
}

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func zero[T any]() (out T) {
	return
}

func extractMessage[T SubMessage](m Message) (T, error) {
	var sm T
	if m.Type != sm.Type() {
		return zero[T](), fmt.Errorf("message type: expected %q, got %q", sm.Type(), m.Type)
	}

	if err := json.Unmarshal(m.Payload, &sm); err != nil {
		return zero[T](), err
	}

	return sm, nil
}

type Comm struct {
	conn net.Conn
	dec  *json.Decoder
	enc  *json.Encoder
}

func NewComm(conn net.Conn) *Comm {
	c := Comm{
		conn: conn,
		dec:  json.NewDecoder(conn),
		enc:  json.NewEncoder(conn),
	}

	return &c
}

func (c *Comm) Close() error {
	return c.conn.Close()
}

func (c *Comm) Send(msg Typed) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	m := Message{
		Type:    msg.Type(),
		Payload: data,
	}

	return c.enc.Encode(m)
}

func (c *Comm) Recv() (Message, error) {
	var m Message
	if err := c.dec.Decode(&m); err != nil {
		return Message{}, nil
	}

	return m, nil
}
