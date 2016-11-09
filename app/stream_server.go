package main

import (
	"os"

	"github.com/stream_server/stream_server"
)

func main() {
	opts := stream_server.NewOptions()
	flagSet := opts.LoadFlagSet()
	flagSet.Parse(os.Args[1:])
	opts.StoreFlagSet(flagSet)

	server := stream_server.NewStreamServer(opts)
	server.Run()
}
