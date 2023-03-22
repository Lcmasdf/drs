package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
)

// for RequestLine
type Method struct {
	M string
}

func (m *Method) parse(content string) error {
	m.M = content
	return nil
}

type RequestURI struct {
	URI string
}

func (m *RequestURI) parse(content string) error {
	m.URI = content
	return nil
}

type RTSPVersion struct {
	Version string
}

func (m *RTSPVersion) parse(content string) error {
	rtspVersionReg := regexp.MustCompile("RTSP/[0-9].[0-9]")
	if ok := rtspVersionReg.Match([]byte(content)); !ok {
		return fmt.Errorf("rtsp version parse failed: %s", content)
	}
	m.Version = content
	return nil
}

//for request
type RequestLine struct {
	Method
	RequestURI
	RTSPVersion
}

func (m *RequestLine) parse(content string) error {
	//RFC2326
	//Request-Line = Method SP Request-URI SP RTSP-Version CRLF

	parts := strings.Split(content, " ")
	if len(parts) != 3 {
		return fmt.Errorf("request-line parse failed: %s", content)
	}

	if err := m.Method.parse(parts[0]); err != nil {
		return err
	}

	if err := m.RequestURI.parse(parts[1]); err != nil {
		return err
	}

	if err := m.RTSPVersion.parse(parts[2]); err != nil {
		return err
	}

	return nil
}

type RequestMessages struct {
	messages map[string]string
}

// parseInc 为增量的parse，每次输入一行
func (m *RequestMessages) parseInc(content string) error {
	if m.messages == nil {
		m.messages = make(map[string]string)
	}

	reg := regexp.MustCompile("(.*): (.*)")
	rets := reg.FindStringSubmatch(content)
	if len(rets) != 3 {
		return fmt.Errorf("request message parse failed: %s||", content)
	}
	m.messages[rets[1]] = rets[2]
	return nil
}

func (m *RequestMessages) reset() {
	m.messages = nil
}

func (m *RequestMessages) GetMessage(msgType string) (string, bool) {
	msg, ok := m.messages[msgType]
	return msg, ok
}

type Request struct {
	RequestLine
	RequestMessages

	Seq int64
}

func (m *Request) GenRequest(conn net.Conn) error {
	return m.Parse(*textproto.NewReader(bufio.NewReader(conn)))
}

func (m *Request) Parse(trd textproto.Reader) error {
	requestLine, err := trd.ReadLine()
	if err != nil {
		return err
	}
	fmt.Println(requestLine)

	if err := m.RequestLine.parse(requestLine); err != nil {
		return err
	}

	for {
		data, err := trd.ReadLine()
		fmt.Println(data)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF {
			return err
		}

		if data == "" {
			break
		}

		if err := m.parseInc(data); err != nil {
			return err
		}
	}
	seq, ok := m.messages["CSeq"]
	if !ok {
		return fmt.Errorf("no cseq")
	}

	m.Seq, err = strconv.ParseInt(seq, 10, 64)
	if err != nil {
		return err
	}

	return nil
}

type Transport struct {
	Items []*TransportItem
}

func parseTransport(b []byte) (*Transport, error) {
	ret := &Transport{
		Items: make([]*TransportItem, 0),
	}
	items := bytes.Split(b, []byte(","))
	for _, item := range items {
		t, err := parseTransportItem(item)
		if err != nil {
			return nil, err
		}
		ret.Items = append(ret.Items, t)
	}
	return ret, nil
}

func genTransport(trans *Transport) ([]byte, error) {
	ret := make([]byte, 0)
	for _, v := range trans.Items {
		itemB, err := genTransportItem(v)
		if err != nil {
			return nil, err
		}
		ret = append(ret, itemB...)
	}
	return ret, nil
}

type TransportItem struct {
	Protocol       string
	Profile        string
	LowerTransport string
	Cast           string
	Parameter      map[string][]byte

	//RTP
	//for multicast
	Port1       int
	Port2       int
	ClientPort1 int
	ClientPort2 int
	ServerPort1 int
	ServerPort2 int
	Ssrc        string
}

func parseTransportItem(b []byte) (*TransportItem, error) {
	ret := &TransportItem{
		Parameter: make(map[string][]byte),
	}

	// RTP/AVP;unicast;client_port=3456-3457;mode="PLAY"
	parts := bytes.Split(b, []byte(";"))
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid transportitem %s", string(b))
	}

	specs := bytes.Split(parts[0], []byte("/"))
	ret.Protocol = string(specs[0])
	ret.Profile = string(specs[1])
	if len(specs) == 2 {
		ret.LowerTransport = "UDP"
	} else {
		ret.LowerTransport = string(specs[2])
	}

	ret.Cast = string(parts[1])

	for i := 2; i < len(parts); i++ {
		index := bytes.Index(parts[i], []byte("="))
		if index == -1 {
			return nil, fmt.Errorf("invalid transportitem %s", string(b))
		}

		var err error
		switch string(parts[i][:index]) {
		case "port":
			ret.Port1, ret.Port2, err = transportRtpPortConv(parts[i][index+1:])
			if err != nil {
				return nil, fmt.Errorf("invalid transportitem %s", string(b))
			}
		case "client_port":
			ret.ClientPort1, ret.ClientPort2, err = transportRtpPortConv(parts[i][index+1:])
			if err != nil {
				return nil, fmt.Errorf("invalid transportitem %s", string(b))
			}
		case "server_port":
			ret.ServerPort1, ret.ServerPort2, err = transportRtpPortConv(parts[i][index+1:])
			if err != nil {
				return nil, fmt.Errorf("invalid transportitem %s", string(b))
			}
		case "ssrc":
			ret.Ssrc = string(parts[i][index+1:])
		default:
			ret.Parameter[string(parts[i][:index])] = parts[i][index+1:]
		}
	}

	return ret, nil
}

func transportRtpPortConv(p []byte) (int, int, error) {
	index := bytes.Index(p, []byte("-"))
	if index == -1 {
		return -1, -1, fmt.Errorf("invalid rtp port: %s", p)
	}

	p1, err := strconv.ParseInt(string(p[:index]), 10, 64)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid rtp port: %s", p)
	}

	p2, err := strconv.ParseInt(string(p[index+1:]), 10, 64)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid rtp port: %s", p)
	}

	return int(p1), int(p2), nil
}

func genTransportItem(t *TransportItem) ([]byte, error) {
	// bytes.Join()
	parts := make([][]byte, 0)

	//RTP/AVP
	p1 := fmt.Sprintf("%s/%s", t.Protocol, t.Profile)
	if t.LowerTransport != "" {
		p1 = fmt.Sprintf("%s/%s", p1, t.LowerTransport)
	}
	parts = append(parts, []byte(p1))
	// unicast/multicast
	parts = append(parts, []byte(t.Cast))
	//port=3456-3457
	if t.Port1 != 0 {
		portStr := fmt.Sprintf("port=%d-%d", t.Port1, t.Port2)
		parts = append(parts, []byte(portStr))
	}
	//client_port=18276-18277
	if t.ClientPort1 != 0 {
		clientPortStr := fmt.Sprintf("client_port=%d-%d", t.ClientPort1, t.ClientPort2)
		parts = append(parts, []byte(clientPortStr))
	}
	//server_port=40658-40659
	if t.ServerPort1 != 0 {
		serverPortStr := fmt.Sprintf("server_port=%d-%d", t.ServerPort1, t.ServerPort2)
		parts = append(parts, []byte(serverPortStr))
	}
	//ssrc
	if t.Ssrc != "" {
		ssrcStr := fmt.Sprintf("ssrc=%s", t.Ssrc)
		parts = append(parts, []byte(ssrcStr))
	}
	for k, v := range t.Parameter {
		paraStr := fmt.Sprintf("%s=%s", k, v)
		parts = append(parts, []byte(paraStr))
	}
	return bytes.Join(parts, []byte(";")), nil
}
