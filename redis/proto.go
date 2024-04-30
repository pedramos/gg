package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tidwall/resp"
)

type Command interface {
	Stmt() string
	Exec(StorageEngine) error
}

const (
	setCmdStmt = "set"
	getCmdStmt = "get"
)

type SetCommand struct{ key, val string }

func (c SetCommand) Stmt() string                { return setCmdStmt }
func (c SetCommand) Exec(kv StorageEngine) error { kv.Set(c.key, c.val); return nil }

type GetCommand struct{ key string }

func (c GetCommand) Stmt() string                { return getCmdStmt }
func (c GetCommand) Exec(kv StorageEngine) error { kv.Get(c.key); return nil }

func parseCommand(raw []byte) (Command, error) {
	rd := resp.NewReader(bytes.NewReader(raw))

	for {
		val, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading value from msg: %w", err)
		}
		if val.Type() == resp.Array {
			switch v := val.Array(); v[0].String() {
			case setCmdStmt:
				if len(v) != 3 {
					return nil, fmt.Errorf("malformed set cmd: %v", v)
				}
				cmd := SetCommand{
					key: v[1].String(),
					val: v[2].String(),
				}
				return cmd, nil
			case getCmdStmt:
				if len(v) != 2 {
					return nil, fmt.Errorf("malformed get cmd: %v", v)
				}
				cmd := GetCommand{key: v[1].String()}
				return cmd, nil
			default:
				return nil, fmt.Errorf("could not identify cmd `%s`", v[0].String())
			}

		}
	}
	return nil, fmt.Errorf("BUG! parsing command in parseCommand function")
}
