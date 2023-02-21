package pkg

import (
	"fmt"
	"net"
)

type Server struct {
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", ":8554")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		session := NewRtspServerSession(conn)
		session.Init()
		go session.Run()
	}
}
