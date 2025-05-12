package jsonlog

import (
	"log/slog"
	"os"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/response"
)

func init() { plugin.Register("jsonlog", setup) }

func setup(c *caddy.Controller) error {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	logLevel, err := parseLogLevel(c)
	if err != nil {
		return plugin.Error("jsonlog", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return JsonLogger{
			Next:     next,
			logger:   logger,
			logLevel: logLevel,
		}
	})

	return nil
}

func parseLogLevel(c *caddy.Controller) (response.Class, error) {
	var logLevel response.Class

	for c.Next() {
		args := c.RemainingArgs()

		switch len(args) {
		case 0:
			logLevel = response.All
		case 1:
			cls, err := response.ClassFromString(args[0])
			if err != nil {
				return logLevel, err
			}
			logLevel = cls
		default:
			return logLevel, c.ArgErr()
		}
	}

	return logLevel, nil
}
