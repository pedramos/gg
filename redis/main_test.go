package main

import (
	"context"
	"fmt"
	"testing"

	"plramos.win/gg/redis/client"
)

func TestServer(t *testing.T) {
	server, err := NewServer(Config{})
	if err != nil {
		t.Errorf("failed to create new server from default config: %s", err)
	}
	go server.Start()
	err = server.WaitReady()
	if err != nil {
		t.Errorf("failed to start server: %s", err)
	}

	serverTests := map[string]func(*testing.T){
		"testSet": func(*testing.T) {
			ctx := context.Background()
			c, err := client.New("localhost" + server.ListenAddr)
			if err != nil {
				t.Errorf("failed to create client: %s", err)
			}
			err = c.Set(ctx, "leader", "Charlie")
			if err != nil {
				t.Errorf("failed to execute set cmd: %s", err)
			}
			val, err := c.Get(ctx, "leader")
			if err != nil {
				t.Errorf("failed to execute get cmd: %s", err)
			}
			if fmt.Sprintf("%v", val) != "Charlie" {
				t.Errorf("exected `Charlie` and got %s", fmt.Sprintf("%v", val))
			}

		},
	}

	for test, f := range serverTests {
		t.Run(test, f)
	}
}
