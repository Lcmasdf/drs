package pkg

import (
	"bufio"
	"net"
)

//rtsp 连接 C->S
type RtspServerSession struct {
	// tcp 连接
	conn net.Conn

	sm *ServerStatusMachine

	id string

	seq int64

	sdp *SDP
}

func NewRtspServerSession(conn net.Conn) *RtspServerSession {
	sm := &ServerStatusMachine{}
	return &RtspServerSession{
		conn: conn,
		sm:   sm,
	}
}

func (rss *RtspServerSession) handleRequest(r *Request) *Response {
	return nil
}

func (rss *RtspServerSession) run() {
	r := bufio.NewReader(rss.conn)
	for {

	}
}
