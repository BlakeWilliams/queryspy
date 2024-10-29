package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/blakewilliams/guesswho/mysql"
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
