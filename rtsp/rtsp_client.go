package rtsp

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
)

type Client struct {
	cSeq      int
	conn      net.Conn
	url       string
	sessionId string
}

func Connect(rawUrl string) (*Client, error) {
	URL, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", URL.Host)
	if err != nil {
		return nil, err
	}

	client := &Client{
		cSeq: 0,
		conn: conn,
		url:  rawUrl,
	}
	return client, nil
}

func (c *Client) nextCSeq() string {
	c.cSeq++
	return strconv.Itoa(c.cSeq)
}

func (c *Client) Options() (*Response, error) {
	req, err := NewRequest(OPTIONS, c.url, c.nextCSeq(), "")
	if err != nil {
		return nil, err
	}

	_, err = io.WriteString(c.conn, req.String())
	if err != nil {
		return nil, err
	}

	fmt.Println(">>OPTIONS\r\n", "SEND:\r\n", req.String())
	res, err := ReadResponse(c.conn)
	if err == nil {
		fmt.Println("RECEIVE:\r\n", res.String())
	}

	return res, err
}

func (c *Client) Describe() (*Response, error) {
	req, err := NewRequest(DESCRIBE, c.url, c.nextCSeq(), "")
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/sdp")
	_, err = io.WriteString(c.conn, req.String())
	if err != nil {
		return nil, err
	}

	fmt.Println(">>DESCRIBE\r\n", "SEND:\r\n", req.String())
	res, err := ReadResponse(c.conn)
	if err == nil {
		fmt.Println("RECEIVE:\r\n", res.String())
	}

	return res, err
}

func (c *Client) Setup(transport string) (*Response, error) {
	req, err := NewRequest(SETUP, c.url, c.nextCSeq(), "")
	if err != nil {
		return nil, err
	}

	req.Header.Add("Transport", transport)
	_, err = io.WriteString(c.conn, req.String())
	if err != nil {
		return nil, err
	}

	fmt.Println(">>SETUP\r\n", "SEND:\r\n", req.String())
	res, err := ReadResponse(c.conn)
	if err == nil {
		fmt.Println("RECEIVE:\r\n", res.String())
		c.sessionId = res.Header.Get("Session")
	}

	return res, err
}

func (c *Client) Play() (*Response, error) {
	req, err := NewRequest(PLAY, c.url, c.nextCSeq(), "")
	if err != nil {
		return nil, err
	}

	req.Header.Add("Session", c.sessionId)
	_, err = io.WriteString(c.conn, req.String())
	if err != nil {
		return nil, err
	}

	fmt.Println(">>PLAY\r\n", "SEND:\r\n", req.String())
	res, err := ReadResponse(c.conn)
	if err == nil {
		fmt.Println("RECEIVE:\r\n", res.String())
	}

	return res, err
}

func (c *Client) Teardown() (*Response, error) {
	req, err := NewRequest(TEARDOWN, c.url, c.nextCSeq(), "")
	if err != nil {
		return nil, err
	}

	req.Header.Add("Session", c.sessionId)
	_, err = io.WriteString(c.conn, req.String())
	if err != nil {
		return nil, err
	}

	fmt.Println(">>TEARDOWN\r\n", "SEND:\r\n", req.String())
	res, err := ReadResponse(c.conn)
	if err == nil {
		fmt.Println("RECEIVE:\r\n", res.String())
	}

	return res, err
}

func (c *Client) Close() {
	c.conn.Close()
}
