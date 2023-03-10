package pkg

import (
	"fmt"
	"net"
	"strings"
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

func (rss *RtspServerSession) Init() {
	rss.sm.OptionsHandler = rss.OptionsHandler
}

func (rss *RtspServerSession) Run() {
	for {
		req := &Request{}
		err := req.GenRequest(rss.conn)
		if err != nil {
			fmt.Println("gen request failed", err.Error())
			break
		}
		fmt.Println(req)

		resp := rss.sm.Request(req)
		fmt.Println(resp)
		data := resp.Gen()

		_, err = rss.conn.Write([]byte(data))
		if err != nil {
			fmt.Println("write line failed ", err.Error())
		}
	}
}

func (rss *RtspServerSession) OptionsHandler(r *Request) *Response {
	// resp := &Response{}
	rss.seq = r.Seq

	respMessage := ResponseMessages{}
	respMessage.AddMessage("Public", strings.Join(methods, ","))

	ret := &Response{
		StatusLine: StatusLine{
			RTSPVersion:  r.Version,
			StatusCode:   "200",
			ReasonPhrase: "OK",
		},
	}
	ret.AddMessage("CSeq", fmt.Sprintf("%d", rss.seq))
	return ret
}
