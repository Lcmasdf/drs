package pkg

type state int

//SM状态
const (
	INIT state = iota
	READY
	PLAYING
)

type method int

const (
	OPTIONS method = iota
	DESCRIBE
	SETUP
	TEARDOWN
	PLAY
	PAUSE
)

var Method2String = map[method]string{
	OPTIONS:  "OPTIONS",
	DESCRIBE: "DESCRIBE",
	SETUP:    "SETUP",
	TEARDOWN: "TEARDOWN",
	PLAY:     "PLAY",
	PAUSE:    "PAUSE",
}

var Method2method = map[string]method{
	"OPTION":   OPTIONS,
	"DESCRIBE": DESCRIBE,
	"SETUP":    SETUP,
	"TEARDOWN": TEARDOWN,
	"PLAY":     PLAY,
	"PAUSE":    PAUSE,
}

type TransitionFunc func(r *Request) *Response
type MethodStateTupple struct {
	Method method
	State  state
}

type ServerStatusMachine struct {
	st              state
	transitionTable map[MethodStateTupple]TransitionFunc

	SetupInitHandler       TransitionFunc
	TeardownInitHandler    TransitionFunc
	SetupReadyHandler      TransitionFunc
	PlayReadyHandler       TransitionFunc
	SetupPlayingHandler    TransitionFunc
	PausePlayingHandler    TransitionFunc
	TeardownPlayingHandler TransitionFunc
	OptionsHandler         TransitionFunc
	DescribeHandler        TransitionFunc
}

func (m *ServerStatusMachine) Init() {
	m.transitionTable = map[MethodStateTupple]TransitionFunc{
		{SETUP, INIT}:       m.SetupInit,
		{TEARDOWN, INIT}:    m.TeardownInit,
		{SETUP, READY}:      m.SetupReady,
		{PLAY, READY}:       m.PlayReady,
		{SETUP, PLAYING}:    m.SetupPlaying,
		{PAUSE, PLAYING}:    m.PausePlaying,
		{TEARDOWN, PLAYING}: m.TeardownPlaying,
	}
}

func (m *ServerStatusMachine) Request(r *Request) *Response {
	f, ok := m.transitionTable[MethodStateTupple{Method2method[r.M], m.st}]

	if ok {
		//需要更改状态机状态
		return f(r)
	} else {
		//不需要更改状态机状态
		switch r.M {
		case Method2String[OPTIONS]:
			return m.OptionsHandler(r)
		case Method2String[DESCRIBE]:
			return m.DescribeHandler(r)
		default:
			return nil
		}
	}
}

func (m *ServerStatusMachine) SetupInit(r *Request) *Response {
	resp := m.SetupInitHandler(r)
	if statusCodeMatch3xx(resp.StatusCode) {
		m.st = INIT
	} else if statusCodeMatch4xx(resp.StatusCode) {
		// m.st no change
	} else if statusCodeMatch2xx(resp.StatusCode) {
		m.st = READY
	}

	return resp
}

func (m *ServerStatusMachine) TeardownInit(r *Request) *Response {
	resp := m.TeardownInitHandler(r)
	if statusCodeMatch2xx(resp.StatusCode) {
		m.st = INIT
	} else {
		// m.st no change
	}

	return resp
}

func (m *ServerStatusMachine) SetupReady(r *Request) *Response {
	resp := m.SetupReadyHandler(r)
	if statusCodeMatch3xx(resp.StatusCode) {
		m.st = INIT
	} else if statusCodeMatch4xx(resp.StatusCode) {
		// m.st no change
	} else if statusCodeMatch2xx(resp.StatusCode) {
		m.st = READY
	}
	return resp
}

func (m *ServerStatusMachine) PlayReady(r *Request) *Response {
	resp := m.PlayReadyHandler(r)
	if statusCodeMatch3xx(resp.StatusCode) {
		m.st = INIT
	} else if statusCodeMatch4xx(resp.StatusCode) {
		// m.st no change
	} else if statusCodeMatch2xx(resp.StatusCode) {
		m.st = PLAYING
	}
	return resp
}

func (m *ServerStatusMachine) SetupPlaying(r *Request) *Response {
	resp := m.SetupPlayingHandler(r)
	if statusCodeMatch3xx(resp.StatusCode) {
		m.st = INIT
	} else if statusCodeMatch4xx(resp.StatusCode) {
		// m.st no change
	} else if statusCodeMatch2xx(resp.StatusCode) {
		m.st = PLAYING
	}
	return resp
}

func (m *ServerStatusMachine) PausePlaying(r *Request) *Response {
	resp := m.PausePlayingHandler(r)
	if statusCodeMatch3xx(resp.StatusCode) {
		m.st = INIT
	} else if statusCodeMatch4xx(resp.StatusCode) {
		// m.st no change
	} else if statusCodeMatch2xx(resp.StatusCode) {
		m.st = READY
	}
	return resp
}

func (m *ServerStatusMachine) TeardownPlaying(r *Request) *Response {
	resp := m.TeardownPlayingHandler(r)
	if statusCodeMatch2xx(resp.StatusCode) {
		m.st = INIT
	} else {
		// m.st no change
	}
	return resp
}

func statusCodeMatch2xx(statusCode string) bool {
	return statusCode[0] == '2'
}

func statusCodeMatch3xx(statusCode string) bool {
	return statusCode[0] == '3'
}

func statusCodeMatch4xx(statusCode string) bool {
	return statusCode[0] == '4'
}
