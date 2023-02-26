package pkg

import (
	"bufio"
	"bytes"
	"net/textproto"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequestParse(t *testing.T) {
	Convey("test request parse", t, func() {
		rLine := "OPTIONS rtsp://127.0.0.1:7776 RTSP/1.0\nCSeq: 1\nUser-Agent: Lavf57.83.100\n"
		r := Request{}
		err := r.Parse(*textproto.NewReader(bufio.NewReader(bytes.NewReader([]byte(rLine)))))
		So(err, ShouldBeNil)
		So(r.M, ShouldEqual, "OPTIONS")
		So(r.Version, ShouldEqual, "RTSP/1.0")
		So(r.Seq, ShouldEqual, 1)
		agent, ok := r.GetMessage("User-Agent")
		So(ok, ShouldBeTrue)
		So(agent, ShouldEqual, "Lavf57.83.100")
	})
}
