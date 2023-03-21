package sdp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
)

var (
	sessionKeySequence = []byte{'v', 'o', 's', 'i', 'u', 'e', 'p', 'c', 'b'}
	mediaKeySequence   = []byte{'m', 'i', 'c', 'b', 'k', 'a'}
)

type SDP interface {
	Parse([]byte) error
	Gen() []byte
}

type SDPItem interface {
	SetItem(byte, []byte) error
	Gen() []byte
}

type SDPImpl struct {
	S  *Session
	Ms []*Media
}

var itemRepeated map[byte]interface{} = map[byte]interface{}{
	'b': nil,
	'a': nil,
	'r': nil,
}

func (s *SDPImpl) Parse(data []byte) error {
	reader := textproto.NewReader(bufio.NewReader(bytes.NewBuffer(data)))

	// Session  Media  Media  Media  Media
	var instance SDPItem

	for {
		line, err := reader.ReadLineBytes()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		k, v, err := ParseLine(line)
		if err != nil {
			return err
		}

		if k == 'v' {
			instance = &Session{
				Item: make(map[byte][][]byte),
			}
			s.S = instance.(*Session)
		} else if k == 'm' {
			instance = &Media{
				Item: make(map[byte][][]byte),
			}
			s.Ms = append(s.Ms, instance.(*Media))
		}

		err = instance.SetItem(k, v)
		if err != nil {
			return err
		}

	}
	return nil
}

func (s *SDPImpl) Gen() []byte {
	ret := make([]byte, 0)
	ret = append(ret, s.S.Gen()...)

	for _, m := range s.Ms {
		ret = append(ret, m.Gen()...)
	}

	return ret
}

func ParseLine(line []byte) (byte, []byte, error) {
	index := bytes.Index(line, []byte("="))
	if index != 1 {
		return 0, nil, fmt.Errorf("line parse error, %s", line)
	}
	return line[0], line[2:], nil
	// if index == -1 {
	// }

	// part := bytes.Split(line, []byte("="))
	// if len(part) != 2 {
	// }
	// if len(part[0]) != 1 {
	// 	return 0, nil, fmt.Errorf("line parse error, %s", line)
	// }
	// if bytes.HasPrefix(part[1], []byte(" ")) {
	// 	return 0, nil, fmt.Errorf("line parse error, %s", line)
	// }

	// return part[0][0], part[1], nil
}

type Session struct {
	Item map[byte][][]byte
}

func (s *Session) SetItem(key byte, value []byte) error {
	_, repteated := itemRepeated[key]
	if !repteated {
		s.Item[key] = [][]byte{value}
		return nil
	}

	_, ok := s.Item[key]
	if !ok {
		s.Item[key] = [][]byte{value}
	} else {
		s.Item[key] = append(s.Item[key], value)
	}
	return nil
}

func (s *Session) Gen() []byte {
	ret := make([]byte, 0)

	for _, v := range sessionKeySequence {
		values, ok := s.Item[v]
		if !ok {
			continue
		}

		for _, value := range values {
			ret = append(append(append(ret, v, '='), value...), '\n')
		}
	}

	return ret
}

type Media struct {
	Item map[byte][][]byte
}

func (m *Media) SetItem(key byte, value []byte) error {
	_, repteated := itemRepeated[key]
	if !repteated {
		m.Item[key] = [][]byte{value}
		return nil
	}

	_, ok := m.Item[key]
	if !ok {
		m.Item[key] = [][]byte{value}
	} else {
		m.Item[key] = append(m.Item[key], value)
	}
	return nil
}

func (m *Media) Gen() []byte {
	ret := make([]byte, 0)

	for _, v := range mediaKeySequence {
		values, ok := m.Item[v]
		if !ok {
			continue
		}

		for _, value := range values {
			ret = append(append(append(ret, v, '='), value...), '\n')
		}
	}
	return ret
}

func (m *Media) GetM() (*M, error) {
	ms, ok := m.Item['m']
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	//m not repeated
	return parseM(ms[0])
}

func (m *Media) GetRtpmaps() ([]*Rtpmap, error) {
	//a=rtpmap:96 L8/8000
	ret := make([]*Rtpmap, 0)

	attrs := m.Item['a']
	for _, attr := range attrs {
		if bytes.HasPrefix(attr, []byte("rtpmap")) {
			r, err := parseRtpmap(attr)
			if err != nil {
				return nil, err
			}
			ret = append(ret, r)
		}
	}

	return ret, nil
}

func (m *Media) GetControl() ([]*Control, error) {
	//a=control:trackID=2
	ret := make([]*Control, 0)

	attrs := m.Item['a']
	for _, attr := range attrs {
		if bytes.HasPrefix(attr, []byte("control")) {
			r, err := parseControl(attr)
			if err != nil {
				return nil, err
			}

			ret = append(ret, r)
		}
	}
	return ret, nil
}

//============================item impl=========================

type M struct {
	Media    string
	Port     int
	PortsNum int
	Proto    string
	Fmt      string
}

func parseM(b []byte) (*M, error) {
	ret := &M{}
	var err error

	// s := strings.Split(str, " ")
	parts := bytes.Split(b, []byte(" "))
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid M %s", b)
	}
	ret.Media = string(parts[0])

	if !bytes.Contains(parts[1], []byte("/")) {
		ret.Port, err = strconv.Atoi(string(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid M %s", b)
		}
		ret.PortsNum = 1
	} else {
		ports := bytes.Split(parts[1], []byte("/"))
		if len(ports) != 2 {
			return nil, fmt.Errorf("invalid M %s", b)
		}
		ret.Port, err = strconv.Atoi(string(ports[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid M %s", b)
		}
		ret.PortsNum, err = strconv.Atoi(string(ports[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid M %s", b)
		}
	}

	ret.Proto = string(parts[2])
	ret.Fmt = string(parts[3])

	return ret, nil
}

type Control struct {
	value string
}

func parseControl(b []byte) (*Control, error) {
	return nil, nil
}

type Rtpmap struct {
	PayloadType   int
	EncodingName  string
	ClockRate     int
	EncodingParam int
}

func parseRtpmap(b []byte) (*Rtpmap, error) {
	var err error
	//rtpmap:96 L8/8000
	rtpmap := b[6:]
	//rtmap-value = payload-type SP encoding-name/clock-rate[/encoding-params]
	parts := bytes.Split(rtpmap, []byte(" "))
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ARtp %s", rtpmap)
	}
	ret := &Rtpmap{}
	ret.PayloadType, err = strconv.Atoi(string(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid ARtp %s", rtpmap)
	}

	attrs := bytes.Split(parts[1], []byte("/"))
	if len(attrs) == 2 {
		ret.EncodingName = string(attrs[0])
		ret.ClockRate, err = strconv.Atoi(string(attrs[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid ARtp %s", rtpmap)
		}
	} else if len(attrs) == 3 {
		ret.EncodingName = string(attrs[0])
		ret.ClockRate, err = strconv.Atoi(string(attrs[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid ARtp %s", rtpmap)
		}
		ret.EncodingParam, err = strconv.Atoi(string(attrs[2]))
		if err != nil {
			return nil, fmt.Errorf("invalid ARtp %s", rtpmap)
		}
	} else {
		return nil, fmt.Errorf("invalid ARtp %s", rtpmap)
	}

	return ret, nil
}
