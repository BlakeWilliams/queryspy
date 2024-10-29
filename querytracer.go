package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/blakewilliams/querytracer/mysql"
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

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("error accepting connection", "err", err)
		}
		go processClient(conn, log)
	}
}

func processClient(conn net.Conn, log *slog.Logger) {
	debug := os.Getenv("DEBUG") == "1"

	defer func() {
		r := recover()

		if err, ok := r.(error); ok {
			log.Error("panic", "err", err)
			return
		}

		log.Error("panic", "r", r)
	}()

	mysqlConn, err := net.Dial("tcp", "127.0.0.1:3306")
	if err != nil {
		panic(err)
	}
	defer conn.Close() // Ensure the connection is closed when done

	// Handle initial handshake and inform client that SSL is _not_ supported
	packet := &mysql.Packet{}
	err = packet.ReadFrom(mysqlConn)
	if err != nil {
		panic(err)
	}
	if debug {
		log.Info("cmd from mysql", "cmd", packet.CommandName())
	}

	payload := packet.RawPayload()
	protocolVersion := packet.RawPayload()[0]

	payload = payload[1:]

	versionEnd := 0
	for i, b := range payload {
		if b == 0x00 {
			versionEnd = i
			break
		}
	}

	version := payload[:versionEnd]
	payload = payload[versionEnd+1:]

	// 4 for thread, 8 for auth-plugin-data, 1 for filler
	payload = payload[4+8+1:]
	capabilities := payload[:2]
	capabilities[1] &^= 0x08

	log.Info(
		"connecting to mysql",
		"version", string(version),
		"protocol_version", int(protocolVersion),
	)

	packet.WriteTo(conn)

	go func() {
		for {
			packet := &mysql.Packet{}
			err = packet.ReadFrom(mysqlConn)
			if err != nil {
				panic(err)
			}
			if debug {
				log.Info("cmd from mysql", "cmd", packet.CommandName())
			}
			packet.WriteTo(conn)
		}
	}()

	for {
		packet := &mysql.Packet{}
		err = packet.ReadFrom(conn)
		if err != nil {
			panic(err)
		}
		if packet.Command() == mysql.ComQuery {
			log.Info("query from client", "query", string(packet.Payload()))
		}
		if debug {
			log.Info("cmd from client", "cmd", packet.CommandName())
		}
		packet.WriteTo(mysqlConn)
	}
}
