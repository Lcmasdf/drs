package pkg

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResponse(t *testing.T) {
	Convey("test response gen", t, func() {
		resp := Response{}
		resp.StatusCode = "200"
		resp.ReasonPhrase = "OK"
		resp.RTSPVersion = "RTSP/1.0"
		resp.AddMessage("Public", "DESCRIBE, SETUP, TEARDOWN, PLAY, PAUSE")
		respContent := "RTSP/1.0 200 OK\nCSeq: 100\nPublic: DESCRIBE, SETUP, TEARDOWN, PLAY, PAUSE\n\n"
		So(respContent, ShouldEqual, resp.Gen())
	})
}
