package stream_server

import (
	"flag"
)

var (
	RTSP_FLAG    = "rtsp-address"
	RTSP_DEFAULT = "127.0.0.1:554"
	RTSP_USAGE   = "<address>:<port> to listen on for RTSP client"

	HTTP_FLAG    = "http-address"
	HTTP_DEFAULT = "127.0.0.1:80"
	HTTP_USAGE   = "<address>:<port> to listen on for HTTP client"
)

type Options struct {
	RTSPAddress string
	HTTPAddress string
}

func NewOptions() *Options {
	return &Options{
		RTSPAddress: RTSP_DEFAULT,
		HTTPAddress: HTTP_DEFAULT,
	}
}

func (opts *Options) LoadFlagSet() *flag.FlagSet {
	flagSet := flag.NewFlagSet("OPTIONS", flag.ExitOnError)

	flagSet.String(RTSP_FLAG, opts.RTSPAddress, RTSP_USAGE)
	flagSet.String(HTTP_FLAG, opts.HTTPAddress, HTTP_USAGE)

	return flagSet
}

func (opts *Options) StoreFlagSet(f *flag.FlagSet) {
	flagRTSP := f.Lookup(RTSP_FLAG)
	if flagRTSP != nil {
		opts.RTSPAddress = flagRTSP.Value.String()
	}

	flagHTTP := f.Lookup(HTTP_FLAG)
	if flagHTTP != nil {
		opts.HTTPAddress = flagHTTP.Value.String()
	}
}
