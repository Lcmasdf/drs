package pkg

import "fmt"

// var methods []string = []string{"DESCRIBE", "SETUP", "TEARDOWN", "PLAY", "PAUSE"}

type StatusLine struct {
	RTSPVersion  string
	StatusCode   string
	ReasonPhrase string
}

func (m *StatusLine) gen() string {
	//Status-Line = RTSP-Version SP Status-Code SP Reason-Phrase CRLF
	return fmt.Sprintf("%s %s %s\n", m.RTSPVersion, m.StatusCode, m.ReasonPhrase)
}

func (m *CSeq) gen() string {
	return fmt.Sprintf("CSeq: %d\n", m.Seq)
}

type ResponseMessages struct {
	messages map[string]string
}

func (m *ResponseMessages) AddMessage(header, content string) {
	if m.messages == nil {
		m.messages = make(map[string]string)
	}

	m.messages[header] = content
}

func (m *ResponseMessages) gen() string {
	ret := ""
	for k, v := range m.messages {
		ret += fmt.Sprintf("%s: %s\n", k, v)
	}
	return ret
}

type Response struct {
	StatusLine
	CSeq
	ResponseMessages
}

func (m *Response) Gen() string {
	ret := ""
	ret += m.StatusLine.gen()
	ret += m.CSeq.gen()
	ret += m.ResponseMessages.gen()
	ret += "\n"
	return ret
}
