// +build ignore

package main

import (
	"log"
	"net"
	"os"

	"github.com/phuslu/fastdns"
	"github.com/valyala/fasthttp"
)

type FasthttpAdapter struct {
	FastdnsHandler fastdns.Handler
}

func (h *FasthttpAdapter) Handler(ctx *fasthttp.MessageCtx) {
	req := fastdns.AcquireMessage()
	defer fastdns.ReleaseMessage(req)

	err := fastdns.ParseMessage(req, ctx.PostBody(), true)
	if err != nil {
		ctx.Error("bad request", fasthttp.StatusBadMessage)
		return
	}

	rw := fastdns.AcquireMemoryResponseWriter()
	defer fastdns.ReleaseMemoryResponseWriter(rw)

	rw.Data = rw.Data[:0]
	rw.Raddr = ctx.RemoteAddr()
	rw.Laddr = ctx.LocalAddr()

	h.FastdnsHandler.ServeDNS(rw, req)

	ctx.SetContentType("application/dns-message")
	_, _ = ctx.Write(rw.Data)

}

type DNSHandler struct{}

func (h *DNSHandler) ServeDNS(rw fastdns.ResponseWriter, req *fastdns.Message) {
	log.Printf("%s] %s: CLASS %s TYPE %s\n", rw.RemoteAddr(), req.Domain, req.Question.Class, req.Question.Type)
	if req.Question.Type == fastdns.TypeA {
		fastdns.HOST(rw, req, 300, []net.IP{{10, 0, 0, 1}})
	} else {
		fastdns.Error(rw, req, fastdns.RcodeNameError)
	}
}

func main() {
	fasthttp.ListenAndServe(os.Args[1], (&FasthttpAdapter{&DNSHandler{}}).Handler)
}
