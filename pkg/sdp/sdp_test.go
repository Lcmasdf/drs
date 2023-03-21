package sdp

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSDPGen(t *testing.T) {
	Convey("test sdp gen", t, func() {
		sdp := &SDPImpl{}
		err := sdp.Parse(MockSDP)
		So(err, ShouldBeNil)
		items := sdp.S.Item['v']
		So(len(items), ShouldEqual, 1)

		for _, m := range sdp.Ms {
			mSession, err := m.GetM()
			So(err, ShouldBeNil)
			if mSession.Media == "video" {
				So(mSession.Proto, ShouldEqual, "RTP/AVP")
			}
		}
	})
}
