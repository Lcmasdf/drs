package pkg

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Lcmasdf/drs/pkg/sdp"
)

//rtsp 连接 C->S
type RtspServerSession struct {
	// tcp 连接
	conn net.Conn

	sm *ServerStatusMachine

	// id string

	seq int64

	sdp sdp.SDP
}

func NewRtspServerSession(conn net.Conn) *RtspServerSession {
	sm := &ServerStatusMachine{}
	// sm.Init()

	//for test
	mockSDP := &sdp.SDPImpl{}
	mockSDP.Parse(sdp.MockSDP)

	return &RtspServerSession{
		conn: conn,
		sm:   sm,
		sdp:  mockSDP,
	}
}

func (rss *RtspServerSession) Init() {
	rss.sm.OptionsHandler = rss.OptionsHandler
	rss.sm.DescribeHandler = rss.DescribeHandler
	rss.sm.SetupInitHandler = rss.SetupInitHandler
	rss.sm.Init()
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
	rss.seq = r.Seq

	ret := &Response{
		StatusLine: StatusLine{
			RTSPVersion:  r.Version,
			StatusCode:   "200",
			ReasonPhrase: "OK",
		},
	}

	ret.AddMessage("Public", strings.Join(methods, ","))
	ret.AddMessage("CSeq", fmt.Sprintf("%d", rss.seq))
	return ret
}

func (rss *RtspServerSession) DescribeHandler(r *Request) *Response {
	rss.seq = r.Seq

	ret := &Response{
		StatusLine: StatusLine{
			RTSPVersion:  r.Version,
			StatusCode:   "200",
			ReasonPhrase: "OK",
		},
	}
	ret.AddMessage("CSeq", fmt.Sprintf("%d", rss.seq))
	ret.AddMessage("Date", time.Now().Format(time.RFC1123))
	ret.AddMessage("Content-Type", "application/sdp")
	ret.AddBody(rss.sdp.Gen())

	return ret
}

func (rss *RtspServerSession) SetupInitHandler(r *Request) *Response {
	rss.seq = r.Seq
	ret := &Response{
		StatusLine: StatusLine{
			RTSPVersion:  r.Version,
			StatusCode:   "200",
			ReasonPhrase: "OK",
		},
	}
	ret.AddMessage("CSeq", fmt.Sprintf("%d", rss.seq))

	transport, ok := r.GetMessage("Transport")
	if !ok {
		ret.StatusCode = "300"
		ret.ReasonPhrase = "transport not found"
		return ret
	}

	//parse transport
	t, err := parseTransport([]byte(transport))
	if err != nil {
		ret.StatusCode = "500"
		ret.ReasonPhrase = err.Error()
		return ret
	}

	for _, item := range t.Items {
		fmt.Println(item.Cast)
		fmt.Println(item.Protocol)
	}

	return ret
}
