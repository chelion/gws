package render

import(
	"github.com/chelion/gws/fasthttp"
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

type Render interface {
	Render(ctx *fasthttp.RequestCtx) error
	WriteContentType(ctx *fasthttp.RequestCtx)
}


func writeContentType(ctx *fasthttp.RequestCtx, value string) {
	ctx.Response.Header.SetContentType(value)
}
