package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/blakewilliams/guesswho/mysql"
	"github.com/blakewilliams/guesswho/tracer"
)

type queryReader struct{}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "6033"
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", port))
	if err != nil {
		log.Error("could not listen on port", "err", err, "port", port)
		os.Exit(1)
	}

	log.Info("listening", "host", "127.0.0.1", "port", port)

	ctx := context.Background()

	history := &tracer.History{Logger: log}
	go history.Process(ctx)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Error("error accepting connection", "err", err)
			}
			log.Info("accepted connection", "remote_addr", conn.RemoteAddr().String())
			go processClient(ctx, conn, log, history)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Info("shutting down")
}

func processClient(ctx context.Context, conn net.Conn, log *slog.Logger, history *tracer.History) {
	defer conn.Close() // Ensure the connection is closed when done
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.Error("panic", "err", err)
				return
			}

			log.Error("panic", "r", r)
			cancel()
		}
	}()

	proxy, err := mysql.NewProxy(conn, "tcp", "127.0.0.1:3306")
	if err != nil {
		panic(err)
	}
	defer proxy.Close()

	proxy.Logger = log
	proxy.Handle(mysql.ComQuery, func(p mysql.Packet) bool {
		query := string(p.Payload())

		query = strings.TrimSpace(query)
		if !strings.HasPrefix(query, "gw") {
			history.Queries <- query
			return true
		}

		log.Info("received command", "query", query)

		if strings.HasSuffix(query, "dump") {
			err := history.Dump()
			if err != nil {
				proxy.ReplyClientOK(p, err.Error())
			}
		}

		proxy.ReplyClientOK(p, "Dump complete")

		return false
	})

	proxy.Handle(mysql.ComStmtPrepare, func(p mysql.Packet) bool {
		history.Queries <- string(p.Payload()[1:])
		return true
	})

	err = proxy.Run(ctx)
	if err != nil {
		log.Error("proxy run failed", "err", err)
	}
}
