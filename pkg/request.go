package pkg

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"

	"gitlab.bj.sensetime.com/nebula/common/log"
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
	log.Info(requestLine)

	if err := m.RequestLine.parse(requestLine); err != nil {
		return err
	}

	for {
		data, err := trd.ReadLine()
		log.Info(data)
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

	// m.Seq = seq
	m.Seq, err = strconv.ParseInt(seq, 10, 64)
	if err != nil {
		return err
	}

	return nil
}
