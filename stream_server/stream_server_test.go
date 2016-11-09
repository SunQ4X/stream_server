package stream_server

import (
	"fmt"
	"testing"
)

func TestStreamServer(t *testing.T) {
	opts := NewOptions()
	fmt.Println("options:", opts)
	NewStreamServer(opts).Run()
}
