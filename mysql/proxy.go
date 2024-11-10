package mysql

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
)

type Proxy struct {
	client net.Conn
	mysql  net.Conn
	Logger *slog.Logger

	clientCapabilities clientCapabilities
	handlers           map[int]func(Packet) bool
}

func NewProxy(client net.Conn, network string, address string) (*Proxy, error) {
	mysqlConn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		client:   client,
		mysql:    mysqlConn,
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		handlers: make(map[int]func(Packet) bool),
	}, nil
}

func (p *Proxy) Close() error {
	p.client.Close()
	return nil
}

func (p *Proxy) Run(ctx context.Context) error {
	if p.Logger == nil {
		p.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	err := p.Handshake()
	if err != nil {
		return err
	}

	return p.ProxyAll()
}

func (p *Proxy) Handshake() error {
	// Handle initial handshake and inform client that SSL is _not_ supported
	authPacket, err := NewAuthPacket(p.mysql)
	if err != nil {
		panic(err)
	}
	authPacket.RemoveSSLSupport()
	p.Logger.Info(
		"connecting to mysql",
		"version", authPacket.MySQLVersion,
		"protocol_version", authPacket.ProtocolVersion,
	)
	authPacket.WriteTo(p.client)

	packet, err := p.ReadPacket(p.client)
	rawPacket := packet.(*rawPacket)

	p.clientCapabilities = clientCapabilitiesFrom(rawPacket.payload[:4])

	if !p.clientCapabilities.Protocol41 {
		p.Logger.Error("client does not support protocol 41")
		return fmt.Errorf("client does not support protocol 41")
	}
	packet.WriteTo(p.mysql)

	return nil
}

func (p *Proxy) ProxyAll() error {
	debug := os.Getenv("DEBUG") == "1"
	errCh := make(chan error, 1)

	go func() {
		for {
			packet, err := p.ReadPacket(p.mysql)
			if err != nil {
				errCh <- err
			}
			if debug {
				p.Logger.Info("cmd from mysql", "cmd", packet.CommandName(), "seq", packet.Seq())
			}
			packet.WriteTo(p.client)
		}
	}()

	go func() {
		for {
			packet, err := p.ReadPacket(p.client)
			if err != nil {
				errCh <- err
			}
			if debug {
				p.Logger.Info("cmd from client", "cmd", packet.CommandName(), "seq", packet.Seq())
			}

			if handler, exists := p.handlers[packet.Command()]; exists {
				if !handler(packet) {
					continue
				}
			}

			packet.WriteTo(p.mysql)
		}
	}()

	return <-errCh
}

func (p *Proxy) Handle(command int, handler func(Packet) bool) {
	if _, exists := p.handlers[command]; exists {
		panic(fmt.Sprintf("handler for command %d already exists", command))
	}

	p.handlers[command] = handler
}

func (p *Proxy) ReplyClientOK(packet Packet, message string) {
	res := NewOKPacket(packet, message, p.clientCapabilities)
	res.WriteTo(p.client)
}

func (p *Proxy) ReadPacket(conn io.Reader) (Packet, error) {
	rawPacket := &rawPacket{capabilities: p.clientCapabilities, isClient: conn == p.client}
	err := rawPacket.ReadFrom(conn)
	if err != nil {
		return nil, err
	}

	switch rawPacket.Command() {
	case ComQuery:
		return newComQueryPacket(rawPacket, p), nil
	default:
		return rawPacket, nil
	}
}
