package sdp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"strings"
)

type SDP interface {
	Parse([]byte) error
	Gen() []byte
}

type SDPItem interface {
	SetItem(string, string) error
}

type SDPImpl struct {
	S  *Session
	Ms []*Media
}

func (s *SDPImpl) Parse(data []byte) error {
	reader := textproto.NewReader(bufio.NewReader(bytes.NewBuffer(data)))

	// Session  Media  Media  Media  Media
	var instance SDPItem

	for {
		str, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		k, v, err := ParseLine(str)
		if err != nil {
			return err
		}

		if k == "v" {
			instance = &Session{}
			s.S = instance.(*Session)
		} else if k == "m" {
			s.Ms = append(s.Ms, instance.(*Media))
			instance = &Media{}
		}

		err = instance.SetItem(k, v)
		if err != nil {
			return err
		}

	}
	return nil
}

func ParseLine(str string) (string, string, error) {
	strs := strings.Split(str, "=")
	if len(strs) != 2 {
		return "", "", fmt.Errorf("line parse error, %s", str)
	}
	if len(strs[0]) != 1 {
		return "", "", fmt.Errorf("line parse error, %s", str)
	}
	if strings.HasPrefix(strs[1], " ") {
		return "", "", fmt.Errorf("line parse error, %s", str)
	}

	return strs[0], strs[1], nil
}

type Session struct{}

func (s *Session) SetItem(key, value string) error {
	return nil
}

type Media struct{}

func (m *Media) SetItem(key, value string) error {
	return nil
}

//============================item impl=========================

type Version struct{}

type Origin struct{}
