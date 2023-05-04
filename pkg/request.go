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

type Session struct {
	SessionId string
	Timeout   uint64
}

func parseSession(b []byte) (*Session, error) {
	index := bytes.Index(b, []byte(";"))
	if index == -1 {
		return &Session{
			SessionId: string(b),
		}, nil
	}

	timeout, err := strconv.ParseUint(string(b[index+1:]), 10, 64)
	if err != nil {
		return nil, err
	}

	return &Session{
		SessionId: string(b[:index]),
		Timeout:   timeout,
	}, nil
}

func genSession(s *Session) ([]byte, error) {
	if s.Timeout == 0 {
		return []byte(s.SessionId), nil
	}

	return []byte(fmt.Sprintf("%s;timeout=%d", s.SessionId, s.Timeout)), nil
}

// frame-level accuracy , relative to the start
type RangeSmpte struct {
	SmpteType      string
	SmpteStartTime string
	SmpteEndTime   string
}

func parseRangeSmpte(b []byte) (*RangeSmpte, error) {
	ret := &RangeSmpte{}
	indexEqual := bytes.Index(b, []byte("="))
	if indexEqual == -1 {
		return nil, fmt.Errorf("invalid range smpte: %s", string(b))
	}
	ret.SmpteType = string(b[:indexEqual])

	indexHorizon := bytes.Index(b, []byte("-"))
	ret.SmpteStartTime = string(b[indexEqual+1 : indexHorizon])
	ret.SmpteEndTime = string(b[indexHorizon+1:])

	return ret, nil
}

func genRangeSmpte(r *RangeSmpte) []byte {
	if len(r.SmpteEndTime) == 0 {
		return []byte(fmt.Sprintf("%s=%s-", r.SmpteType, r.SmpteStartTime))
	} else {
		return []byte(fmt.Sprintf("%s=%s-%s", r.SmpteType, r.SmpteStartTime, r.SmpteEndTime))
	}
}

// absolute position relative to the beginning of the presentation
type RangeNpt struct {
	NptStartTime string
	NptEndTime   string
	Now          bool
}

func parseRangeNpt(b []byte) (*RangeNpt, error) {
	ret := &RangeNpt{}
	indexEqual := bytes.Index(b, []byte("-"))
	if indexEqual == -1 {
		return nil, fmt.Errorf("invalid range npt: %s", string(b))
	}
	if indexEqual == 0 {
		ret.NptEndTime = string(b[indexEqual+1:])
	} else {
		if string(b[:indexEqual]) == "now" {
			ret.Now = true
		} else {
			ret.NptStartTime = string(b[:indexEqual])
		}
		ret.NptEndTime = string(b[indexEqual+1:])
	}
	return ret, nil
}

func genRangeNpt(r *RangeNpt) []byte {
	if r.Now {
		return []byte(fmt.Sprintf("now-%s", r.NptEndTime))
	} else if len(r.NptStartTime) == 0 {
		return []byte(fmt.Sprintf("-%s", r.NptEndTime))
	} else {
		return []byte(fmt.Sprintf("%s-%s", r.NptStartTime, r.NptEndTime))
	}
}

type RangeUtc struct {
	UtcStartTime string
	UtcEndTime   string
}

func parseRangeUtc(b []byte) (*RangeUtc, error) {
	ret := &RangeUtc{}
	indexEqual := bytes.Index(b, []byte("="))
	indexHorizon := bytes.Index(b, []byte("-"))

	ret.UtcStartTime = string(b[indexEqual+1 : indexHorizon])
	ret.UtcEndTime = string(b[indexHorizon+1:])
	return ret, nil
}

func genRangeUtc(r *RangeUtc) []byte {
	return []byte(fmt.Sprintf("clock=%s-%s", r.UtcStartTime, r.UtcEndTime))
}

type Range struct {
	Smpte *RangeSmpte
	Npt   *RangeNpt
	Utc   *RangeUtc
	Time  string
}

func parseTime(b []byte) ([]byte, error) {
	index := bytes.Index(b, []byte("="))
	if index == -1 {
		return nil, fmt.Errorf("invalid time: %s", string(b))
	}
	return b[index+1:], nil
}

func genTime(t string) []byte {
	return []byte(fmt.Sprintf("time=%s", t))
}

func parseRange(b []byte) (*Range, error) {
	ret := &Range{}
	var err error
	index := bytes.Index(b, []byte(";"))
	if index == -1 {
		return nil, fmt.Errorf("invalid range: %s", string(b))
	}
	specifier := b[:index]
	if bytes.HasPrefix(specifier, []byte("smpte")) {
		ret.Smpte, err = parseRangeSmpte(specifier)
		if err != nil {
			return nil, err
		}
	} else if bytes.HasPrefix(specifier, []byte("npt")) {
		ret.Npt, err = parseRangeNpt(specifier)
		if err != nil {
			return nil, err
		}
	} else if bytes.HasPrefix(specifier, []byte("clock")) {
		ret.Utc, err = parseRangeUtc(specifier)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid range: %s", string(b))
	}

	if index+1 < len(b) {
		var timeBytes []byte
		timeBytes, err = parseTime(b[index+1:])
		if err != nil {
			return nil, err
		}
		ret.Time = string(timeBytes)
	}
	return ret, nil
}

func genRange(r *Range) ([]byte, error) {
	ret := make([]byte, 0)
	if r.Smpte != nil {
		ret = append(ret, genRangeSmpte(r.Smpte)...)
	}
	if r.Npt != nil {
		ret = append(ret, genRangeNpt(r.Npt)...)
	}
	if r.Utc != nil {
		ret = append(ret, genRangeUtc(r.Utc)...)
	}
	if len(r.Time) != 0 {
		ret = append(ret, genTime(r.Time)...)
	}
	return ret, nil
}

type RTPInfoItem struct {
	Url     string
	Seq     string
	RtpTime string
}

type RTPInfo struct {
	Items []*RTPInfoItem
}

func parseRTPInfo(b []byte) (*RTPInfo, error) {
	ret := &RTPInfo{}
	items := bytes.Split(b, []byte(","))
	for _, item := range items {
		parts := bytes.Split(item, []byte(";"))
		rtpItem := &RTPInfoItem{}
		for _, part := range parts {
			if bytes.HasPrefix(part, []byte("url")) {
				rtpItem.Url = string(part[4:])
			} else if bytes.HasPrefix(part, []byte("seq")) {
				rtpItem.Seq = string(part[4:])
			} else if bytes.HasPrefix(part, []byte("rtptime")) {
				rtpItem.RtpTime = string(part[8:])
			} else {
				return nil, fmt.Errorf("invalid rtp info: %s", string(b))
			}
		}
		ret.Items = append(ret.Items, rtpItem)
	}
	return ret, nil
}

func genRTPInfo(info *RTPInfo) []byte {
	ret := make([]byte, 0)

	for _, item := range info.Items {
		itemStr := fmt.Sprintf("url=%s", item.Url)
		if len(item.Seq) != 0 {
			itemStr = fmt.Sprintf("%s;seq=%s", itemStr, item.Seq)
		}
		if len(item.RtpTime) != 0 {
			itemStr = fmt.Sprintf("%s;rtptime=%s", itemStr, item.RtpTime)
		}

		ret = append(ret, []byte(itemStr)...)
	}
	return ret
}
