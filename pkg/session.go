package pkg

import (
	"fmt"
	"math/rand"
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

	sessionId string

	seq int64

	sdp sdp.SDP

	transport *TransportItem
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
	rss.sm.PlayReadyHandler = rss.PlayReadyHandler
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

	//TODO: transport select
	// for test, use index 0 transport item
	rss.transport = t.Items[0]

	// TODO: get pair of udp ports
	// use 30001-30002
	rss.transport.ServerPort1 = 30001
	rss.transport.ServerPort2 = 30002

	// gen ssrc
	rss.transport.Ssrc = genSsrc()

	// gen session
	rss.sessionId = genRandomSessionId()
	session, err := genSession(&Session{
		SessionId: rss.sessionId,
		Timeout:   60,
	})
	if err != nil {
		ret.StatusCode = "500"
		ret.ReasonPhrase = err.Error()
		return ret
	}
	ret.AddMessage("Session", string(session))

	transResp := &Transport{
		Items: []*TransportItem{
			rss.transport,
		},
	}

	transRespByte, err := genTransport(transResp)
	if err != nil {
		ret.StatusCode = "500"
		ret.ReasonPhrase = err.Error()
		return ret
	}

	ret.AddMessage("Transport", string(transRespByte))

	return ret
}

func genRandomSessionId() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Int63())
}

func genSsrc() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%08x", rand.Uint32())
}

func (rss *RtspServerSession) PlayReadyHandler(r *Request) *Response {
	rss.seq = r.Seq
	resp := &Response{
		StatusLine: StatusLine{
			RTSPVersion:  r.Version,
			StatusCode:   "200",
			ReasonPhrase: "OK",
		},
	}

	resp.AddMessage("CSeq", fmt.Sprintf("%d", rss.seq))
	sessionStr, err := genSession(&Session{
		SessionId: rss.sessionId})
	if err != nil {
		resp.StatusCode = "500"
		resp.ReasonPhrase = err.Error()
		return resp
	}

	resp.AddMessage("CSeq", fmt.Sprintf("%d", rss.seq))

	resp.AddMessage("Session", string(sessionStr))

	rg, _ := genRange(&Range{
		Npt: &RangeNpt{
			NptStartTime: "0.000000",
		},
	})

	resp.AddMessage("Range", string(rg))

	rtpInfo := &RTPInfo{
		Items: []*RTPInfoItem{
			{
				Url:     "trackID=0",
				Seq:     "59394",
				RtpTime: "1271981258",
			},
			{
				Url:     "trackID=1",
				Seq:     "9530",
				RtpTime: "35281645",
			},
			{
				Url:     "trackID=2",
				Seq:     "30356",
				RtpTime: "35302471",
			},
		},
	}

	resp.AddMessage("RTP-Info", string(genRTPInfo(rtpInfo)))

	return resp
}
