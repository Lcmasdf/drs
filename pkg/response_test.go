package pkg

import (
	"bytes"
	"fmt"
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

func TestStatusLine_Gen(t *testing.T) {
	Convey("Given a StatusLine struct", t, func() {
		m := &StatusLine{
			RTSPVersion:  "RTSP/1.0",
			StatusCode:   "200",
			ReasonPhrase: "OK",
		}

		Convey("When calling gen()", func() {
			result := m.gen()

			Convey("The result should be correct", func() {
				expected := []byte("RTSP/1.0 200 OK\n")
				So(bytes.Equal([]byte(result), expected), ShouldBeTrue)
			})
		})
	})
}

func TestResponseMessages(t *testing.T) {
	// Initialize a new ResponseMessages instance to use for testing.
	rm := &ResponseMessages{}

	Convey("Given an empty ResponseMessages object", t, func() {

		Convey("When we add a new message", func() {
			header, content := "Header 1", "This is the first message."
			rm.AddMessage(header, content)

			Convey("Then the gen method returns the correct output", func() {
				expected := fmt.Sprintf("%s: %s\n", header, content)
				So(rm.gen(), ShouldEqual, expected)
			})

			Convey("And we add another message with a different header", func() {
				header2, content2 := "Header 2", "This is the second message."
				rm.AddMessage(header2, content2)

				Convey("Then the gen method returns both messages in the correct order", func() {
					expected := fmt.Sprintf("%s: %s\n%s: %s\n", header, content, header2, content2)
					So(rm.gen(), ShouldEqual, expected)
				})
			})

			Convey("And we add another message with the same header", func() {
				content2 := "This is another message with the same header."
				rm.AddMessage(header, content2)

				Convey("Then the gen method returns only the latest message for that header", func() {
					expected := fmt.Sprintf("%s: %s\n", header, content2)
					So(rm.gen(), ShouldNotEqual, expected)
				})
			})
		})
	})

}
