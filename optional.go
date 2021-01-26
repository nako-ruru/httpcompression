package httpcompression

import (
	"net/http"
)

// extend returns a http.ResponseWriter that wraps the compressWriter and that
// dynamically exposes some optional interfaces of http.ResponseWriter.
// Currently the supported optional interfaces are http.Hijacker, http.Pusher,
// and http.CloseNotifier.
func extend(cw *compressWriter) http.ResponseWriter {
	switch r := cw.ResponseWriter.(type) {
	case iHijackPushCloseNotifier:
		return cwHijackPushCloseNotifier{cw, r}
	case iPushCloseNotifier:
		return cwPushCloseNotifier{cw, r}
	case iHijackPusher:
		return cwHijackPusher{cw, r}
	case iHijackCloseNotifier:
		return cwHijackCloseNotifier{cw, r}
	case http.CloseNotifier:
		return cwCloseNotifier{cw, r}
	case http.Hijacker:
		return cwHijacker{cw, r}
	case http.Pusher:
		return cwPusher{cw, r}
	default:
		return cw
	}
}

type cwHijacker struct {
	*compressWriter
	http.Hijacker
}

var _ http.Hijacker = cwHijacker{}

type cwCloseNotifier struct {
	*compressWriter
	http.CloseNotifier
}

var _ http.CloseNotifier = cwCloseNotifier{}

type cwPusher struct {
	*compressWriter
	http.Pusher
}

var _ http.Pusher = cwPusher{}

type cwHijackCloseNotifier struct {
	*compressWriter
	iHijackCloseNotifier
}

type iHijackCloseNotifier interface {
	http.Hijacker
	http.CloseNotifier
}

var (
	_ http.Hijacker      = cwHijackCloseNotifier{}
	_ http.CloseNotifier = cwHijackCloseNotifier{}
)

type cwHijackPusher struct {
	*compressWriter
	iHijackPusher
}

type iHijackPusher interface {
	http.Hijacker
	http.Pusher
}

var (
	_ http.Hijacker = cwHijackPusher{}
	_ http.Pusher   = cwHijackPusher{}
)

type cwPushCloseNotifier struct {
	*compressWriter
	iPushCloseNotifier
}

type iPushCloseNotifier interface {
	http.Pusher
	http.CloseNotifier
}

var (
	_ http.Pusher        = cwPushCloseNotifier{}
	_ http.CloseNotifier = cwPushCloseNotifier{}
)

type cwHijackPushCloseNotifier struct {
	*compressWriter
	iHijackPushCloseNotifier
}

type iHijackPushCloseNotifier interface {
	http.Hijacker
	http.Pusher
	http.CloseNotifier
}

var (
	_ http.Hijacker      = cwHijackPushCloseNotifier{}
	_ http.Pusher        = cwHijackPushCloseNotifier{}
	_ http.CloseNotifier = cwHijackPushCloseNotifier{}
)
