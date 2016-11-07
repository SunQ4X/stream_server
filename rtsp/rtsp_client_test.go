package rtsp

import (
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	var host = "rtsp://192.168.188.86:554/ch01_sub.264"

	client, err := Connect(host)
	if err != nil {
		fmt.Println("Connect to Server Failed", err)
		return
	}

	_, err = client.Options()
	if err != nil {
		fmt.Println("OPTIONS Failed", err)
		return
	}

	_, err = client.Describe()
	if err != nil {
		fmt.Println("DESCRIBE Failed", err)
		return
	}

	_, err = client.Setup("RTP/AVP;unicast;client_port=9000-9001")
	if err != nil {
		fmt.Println("SETUP Failed", err)
		return
	}

	_, err = client.Play()
	if err != nil {
		fmt.Println("PLAY Failed", err)
		return
	}
}
