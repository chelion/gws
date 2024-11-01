package render

import (
	"encoding/xml"
	"fmt"
	"testing"
)

type TestBuffer struct {
	str string
}

func (tb *TestBuffer) Write(p []byte) (n int, err error) {
	tb.str = tb.str + string(p) + "\n"
	return len(p), nil
}
func (tb *TestBuffer) SetContentType(contentType string) {
	tb.str += contentType + "\n"
}
func (tb *TestBuffer) SetContentLength(len int) {
	tb.str += "ContentLength:" + fmt.Sprintf("%d", len) + "\n"
}
func (tb *TestBuffer) Print() {
	fmt.Println(tb.str)
}
func (tb *TestBuffer) Reset() {
	tb.str = ""
}

type TestM map[string]interface{}

// MarshalXML allows type M to be used with xml.Marshal.
func (v TestM) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range v {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func TestRender(t *testing.T) {

	testBuffer := &TestBuffer{}
	IndentedJSON{Data: TestM{
		"Name": "chelion",
	}}.Render(RenderIO(testBuffer))
	testBuffer.Print()
	testBuffer.Reset()

	Data{Data: []byte("123456"), ContentType: "user data"}.Render(RenderIO(testBuffer))
	testBuffer.Print()
	testBuffer.Reset()

	MsgPack{Data: TestM{
		"Name": "chelion",
	}}.Render(RenderIO(testBuffer))
	testBuffer.Print()
	testBuffer.Reset()

	String{Format: "len:%d,content:%s", Data: []interface{}{16, "user data"}}.Render(RenderIO(testBuffer))
	testBuffer.Print()
	testBuffer.Reset()

	HTML{}

	XML{Data: TestM{
		"Name": "chelion",
	}}.Render(RenderIO(testBuffer))
	testBuffer.Print()
	testBuffer.Reset()

	YAML{Data: TestM{
		"Name": "chelion",
	}}.Render(RenderIO(testBuffer))
	testBuffer.Print()
	testBuffer.Reset()
}
