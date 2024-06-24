package client

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/tidwall/resp"
)

type Client struct{ net.Conn }

func New(addr string) (Client, error) {
	c, err := net.Dial("tcp", addr)
	return Client{Conn: c}, err
}

func (c Client) Set(ctx context.Context, key, val string) error {
	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	wr.WriteArray([]resp.Value{
		resp.StringValue("set"),
		resp.StringValue(key),
		resp.StringValue(val),
	})

	if _, err := buf.WriteTo(c); err != nil {
		return fmt.Errorf("failed to write to %s: %w", c.RemoteAddr(), err)
	}

	_, isOK, err := c.ReadOK()
	if err != nil {
		return fmt.Errorf("set operation did not return ok: %s", err)
	}
	if !isOK {
		return fmt.Errorf("set operation failled")
	}
	// fmt.Println("Set done")
	return nil
}

func (c Client) Get(ctx context.Context, key string) (string, error) {
	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	wr.WriteArray([]resp.Value{
		resp.StringValue("get"),
		resp.StringValue(key),
	})

	if _, err := buf.WriteTo(c); err != nil {
		return "", fmt.Errorf("failed to write to %s: %w", c.RemoteAddr(), err)
	}
	rd := resp.NewReader(c)
	v, _, err := rd.ReadValue()
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", key, err)
	}

	return fmt.Sprintf("%s", v), nil
}

func (c *Client) ReadOK() ([]byte, bool, error) {
	okbuff := make([]byte, 5)
	_, err := c.Read(okbuff)
	if err != nil {
		return nil, false, fmt.Errorf("reading OK from server: %w", err)
	}
	if slices.Compare(okbuff, []byte("+OK\r\n")) == 0 {
		return okbuff, true, nil
	} else {
		return okbuff, false, nil
	}
}
