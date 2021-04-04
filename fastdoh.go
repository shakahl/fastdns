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
	msg := fastdns.AcquireMessage()
	defer fastdns.ReleaseMessage(msg)

	err := fastdns.ParseMessage(msg, ctx.PostBody(), true)
	if err != nil {
		ctx.Error("bad request", fasthttp.StatusBadMessage)
		return
	}

	rw := fastdns.AcquireMemoryResponseWriter()
	defer fastdns.ReleaseMemoryResponseWriter(rw)

	rw.Data = rw.Data[:0]
	rw.Raddr = ctx.RemoteAddr()
	rw.Laddr = ctx.LocalAddr()

	h.FastdnsHandler.ServeDNS(rw, msg)

	ctx.SetContentType("application/dns-message")
	_, _ = ctx.Write(rw.Data)

}

type DNSHandler struct{}

func (h *DNSHandler) ServeDNS(rw fastdns.ResponseWriter, msg *fastdns.Message) {
	log.Printf("%s] %s: CLASS %s TYPE %s\n", rw.RemoteAddr(), msg.Domain, msg.Question.Class, msg.Question.Type)
	if msg.Question.Type == fastdns.TypeA {
		fastdns.HOST(rw, msg, []net.IP{{10, 0, 0, 1}}, 300)
	} else {
		fastdns.Error(rw, msg, fastdns.RcodeNameError)
	}
}

func main() {
	fasthttp.ListenAndServe(os.Args[1], (&FasthttpAdapter{&DNSHandler{}}).Handler)
}