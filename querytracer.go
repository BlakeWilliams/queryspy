package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

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

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("error accepting connection", "err", err)
		}
		go processClient(ctx, conn, log, history)
	}
}

func processClient(ctx context.Context, conn net.Conn, log *slog.Logger, history *tracer.History) {
	debug := os.Getenv("DEBUG") == "1"
	defer conn.Close() // Ensure the connection is closed when done

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.Error("panic", "err", err)
				return
			}

			log.Error("panic", "r", r)
		}
	}()

	mysqlConn, err := net.Dial("tcp", "127.0.0.1:3306")
	if err != nil {
		panic(err)
	}
	defer mysqlConn.Close()

	// Handle initial handshake and inform client that SSL is _not_ supported
	authPacket, err := mysql.NewAuthPacket(mysqlConn)
	if err != nil {
		panic(err)
	}
	authPacket.RemoveSSLSupport()
	log.Info(
		"connecting to mysql",
		"version", authPacket.MySQLVersion,
		"protocol_version", authPacket.ProtocolVersion,
	)
	authPacket.WriteTo(conn)

	packet := &mysql.Packet{}
	err = packet.ReadFrom(conn)
	capabilities := binary.LittleEndian.Uint32(packet.RawPayload()[:4])

	if capabilities&mysql.ClientCapabilityClientProtocol41 != mysql.ClientCapabilityClientProtocol41 {
		log.Error("client does not support protocol 41")
		panic("client does not support protocol 41")
	}

	if capabilities&mysql.ClientCapabilitySessionTrack != mysql.ClientCapabilitySessionTrack {
		panic("Add support for non-session track")
	}

	packet.WriteTo(mysqlConn)

	errCh := make(chan error, 1)

	go func() {
		for {
			packet := &mysql.Packet{}
			err = packet.ReadFrom(mysqlConn)
			if err != nil {
				errCh <- err
			}
			if debug {
				log.Info("cmd from mysql", "cmd", packet.CommandName(), "seq", packet.SeqID())
			}
			packet.WriteTo(conn)
		}
	}()

	go func() {
		for {
			packet := &mysql.Packet{}
			err = packet.ReadFrom(conn)
			if err != nil {
				errCh <- err
			}
			if debug {
				log.Info("cmd from client", "cmd", packet.CommandName(), "seq", packet.SeqID())
			}
			if packet.Command() == mysql.ComQuery {
				query := string(packet.Payload())
				sanitized := strings.TrimSpace(query)
				if strings.HasPrefix(sanitized, "gw") {
					if strings.HasSuffix(sanitized, "dump") {
						err := history.Dump()
						if err != nil {
							res := mysql.NewOKPacket(packet, err.Error())
							res.WriteTo(conn)
							continue
						}

						res := mysql.NewOKPacket(packet, "Dump successful")
						res.WriteTo(conn)
						continue
					}

					continue
				}
				history.Queries <- query
			}
			packet.WriteTo(mysqlConn)
		}
	}()

	select {
	case <-ctx.Done():
		log.Error("context canceled", "err", ctx.Err())
	case err = <-errCh:
		log.Error("error processing connection", "err", err)
	}
}
