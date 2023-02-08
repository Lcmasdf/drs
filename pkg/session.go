package pkg

import "net"

//rtsp 连接 C->S
type RtspServerSession struct {
	// tcp 连接
	conn net.Conn

	sm *ServerStatusMachine

	id string

	seq int64
}
