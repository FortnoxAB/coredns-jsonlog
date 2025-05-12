package jsonlog

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/pkg/response"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type JsonLogger struct {
	Next plugin.Handler

	logger   *slog.Logger
	logLevel response.Class
}

func (l JsonLogger) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	state := request.Request{W: w, Req: r}

	rrw := dnstest.NewRecorder(w)
	rc, err := plugin.NextOrFailure(l.Name(), l.Next, ctx, rrw, r)

	attrs := []any{
		slog.String("remote.ip", state.IP()),
		slog.String("remote.port", state.Port()),
		slog.Int("id", int(state.Req.Id)),
		slog.String("type", state.Type()),
		slog.String("class", state.Class()),
		slog.String("name", state.Name()),
		slog.String("proto", state.Proto()),
		slog.Int("size", state.Req.Len()),
		slog.Bool("do", state.Do()),
		slog.Int("opcode", state.Req.Opcode),
		slog.String("rcode", dns.RcodeToString[rrw.Rcode]),
		slog.Int("rsize", rrw.Len),
		slog.String("duration", fmt.Sprintf("%.4fms", time.Since(rrw.Start).Seconds()*1000)),
	}

	responseType, _ := response.TypeFromString(dns.RcodeToString[rrw.Rcode])
	responseClass := response.Classify(responseType)

	if responseClass >= l.logLevel {
		if responseClass == response.Error {
			l.logger.Error(state.Name(), slog.Group("coredns", attrs...))
		} else {
			l.logger.Info(state.Name(), slog.Group("coredns", attrs...))
		}
	}

	return rc, err
}

// Name implements the Handler interface.
func (l JsonLogger) Name() string { return "jsonlog" }
