package render

import(
	"io"
)

var(
	HTMLContentType = "text/html; charset=utf-8"
	MsgPackContentType = "application/msgpack; charset=utf-8"
	JSONContentType = "application/json; charset=utf-8"
	JSONPContentType = "application/javascript; charset=utf-8"
	JSONAsciiContentType = "application/json"
	ProtoBufContentType = "application/x-protobuf"
	PlainContentType = "text/plain; charset=utf-8"
	XMLContentType = "application/xml; charset=utf-8"
	YAMLContentType = "application/x-yaml; charset=utf-8"
	ImagePNGContentType = "Content-Type: image/png"
	ImageJPEGContentType = "Content-Type: image/jpeg"
	ImageGIFContentType = "Content-Type: image/gif"
	ImageIconContentType = "Content-Type: image/x-icon"
	VideoMPEG4ContentType = "Content-Type: video/mpeg4"
	OctetStreamContentType = "Content-Type: application/octet-stream"
)

type RenderIO interface{
	io.Writer
	SetContentType(contentType string)
	SetContentLength(len int)
}

type Render interface {
	Render(renderIO RenderIO) error
}
