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
		{SETUP, INIT}:       m.SetupInitHandler,
		{TEARDOWN, INIT}:    m.TeardownInitHandler,
		{SETUP, READY}:      m.SetupReadyHandler,
		{PLAY, READY}:       m.PlayReadyHandler,
		{SETUP, PLAYING}:    m.SetupPlayingHandler,
		{PAUSE, PLAYING}:    m.PausePlayingHandler,
		{TEARDOWN, PLAYING}: m.TeardownPlayingHandler,
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
