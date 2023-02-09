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
	OPTION method = iota
	DESCRIBE
	SETUP
	TEARDOWN
	PLAY
	PAUSE
)

var Method2String = map[method]string{
	OPTION:   "OPTION",
	DESCRIBE: "DESCRIBE",
	SETUP:    "SETUP",
	TEARDOWN: "TEARDOWN",
	PLAY:     "PLAY",
	PAUSE:    "PAUSE",
}

var Method2method = map[string]method{
	"OPTION":   OPTION,
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
}

func (m *ServerStatusMachine) Init() {
	m.transitionTable = map[MethodStateTupple]TransitionFunc{
		{SETUP, INIT}:       m.tSetupInit,
		{TEARDOWN, INIT}:    m.tTeardownInit,
		{SETUP, READY}:      m.tSetupReady,
		{PLAY, READY}:       m.tPlayReady,
		{SETUP, PLAYING}:    m.tSetupPlaying,
		{PAUSE, PLAYING}:    m.tPausePlaying,
		{TEARDOWN, PLAYING}: m.tTeardownPlaying,
	}
}

//状态转移方法
func (m *ServerStatusMachine) tSetupInit(r *Request) *Response {
	m.st = READY
	return nil
}

//状态转移方法
func (m *ServerStatusMachine) tTeardownInit(r *Request) *Response {
	m.st = INIT
	return nil
}

//状态转移方法
func (m *ServerStatusMachine) tSetupReady(r *Request) *Response {
	m.st = READY
	return nil
}

//状态转移方法
func (m *ServerStatusMachine) tPlayReady(r *Request) *Response {
	m.st = PLAYING
	return nil
}

//状态转移方法
func (m *ServerStatusMachine) tSetupPlaying(r *Request) *Response {
	m.st = PLAYING
	return nil
}

//状态转移方法
func (m *ServerStatusMachine) tPausePlaying(r *Request) *Response {
	m.st = READY
	return nil
}

//状态转移方法
func (m *ServerStatusMachine) tTeardownPlaying(r *Request) *Response {
	m.st = INIT
	return nil
}

func (m *ServerStatusMachine) options(r *Request) *Response {
	return nil
}

func (m *ServerStatusMachine) describe(r *Request) *Response {
	return nil
}

func (m *ServerStatusMachine) Request(r *Request) *Response {
	f, ok := m.transitionTable[MethodStateTupple{Method2method[r.M], m.st}]

	if ok {
		//需要更改状态机状态
		return f(r)
	} else {
		//不需要更改状态机状态
		switch r.M {
		case Method2String[OPTION]:
			return m.options(r)
		case Method2String[DESCRIBE]:
			return m.describe(r)
		default:
			return nil
		}
	}
}
