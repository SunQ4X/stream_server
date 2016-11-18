package media

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/stream_server/util"
)

type MediaSession struct {
	streamName string
	serialNum  string
	chNo       int
	streamType string
	timeval    util.TimeVal
	ipAddress  string
}

var (
	MapMutex   sync.Mutex
	SessionMap map[string]*MediaSession
)

func init() {
	SessionMap = make(map[string]*MediaSession)
}

func NewMediaSession(streamName string) (*MediaSession, error) {
	temps := strings.SplitN(streamName, "&", 3)
	if 3 != len(temps) {
		return nil, errors.New("invalid stream name format")
	}

	chNo, err := strconv.Atoi(temps[1])
	if err != nil {
		return nil, errors.New("invalid stream name format")
	}

	sess := &MediaSession{
		streamName: streamName,
		serialNum:  temps[0],
		chNo:       chNo,
		streamType: temps[2],
	}

	util.GetCurrentTimeVal(&sess.timeval)
	sess.ipAddress, _ = util.GetLocalIPAddress()

	MapMutex.Lock()
	SessionMap[streamName] = sess
	MapMutex.Unlock()

	return sess, nil
}

func LookupMediaSession(streamName string) (*MediaSession, bool) {
	MapMutex.Lock()
	v, ok := SessionMap[streamName]
	MapMutex.Unlock()
	return v, ok
}

func (sess *MediaSession) GenerateSDPDescription() string {
	sdpFmt := "v=0\r\n" +
		"o=- %d %d IN IP4 %s\r\n" +
		"s=%s\r\n" +
		"t=0 0"

	return fmt.Sprintf(sdpFmt,
		sess.timeval.Sec,
		sess.timeval.Usec,
		sess.ipAddress,
		sess.streamName)
}
