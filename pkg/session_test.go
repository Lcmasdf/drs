package pkg

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenSsrc(t *testing.T) {
	Convey("Given a call to genSsrc", t, func() {
		Convey("When called, it should generate a random 8 digit hex string", func() {
			ssrc := genSsrc()
			So(ssrc, ShouldHaveLength, 8)
			var in string
			_, err := fmt.Sscanf(ssrc, "%x", &in)
			So(err, ShouldBeNil)
		})
	})
}
