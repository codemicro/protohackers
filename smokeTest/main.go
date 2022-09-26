package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	Address string `default:"127.0.0.1:9000"`
}

func main() {
	conf := new(Config)
	arg.MustParse(conf)

	if err := run(conf); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run(conf *Config) error {

	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGINT)

	listener, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return err
	}
	defer listener.Close()

	_, _ = fmt.Fprintf(os.Stderr, "Listening on %s\n", conf.Address)

	connChan := make(chan net.Conn, 256)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Failed to accept connection: %v", err)
				return
			}
			connChan <- conn
		}
	}()

loop:
	for {
		select {
		case <-stopSig:
			break loop
		case conn := <-connChan:
			go func() {
				err := handleConn(conn)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error when handling connection: %s: %v\n", conn.RemoteAddr(), err)
				}
			}()
		}
	}

	return nil
}

func handleConn(conn net.Conn) error {
	defer conn.Close()

	if n, err := io.Copy(conn, conn); err != nil {
		return err
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Copied %d bytes for %s\n", n, conn.RemoteAddr())
	}

	return nil
}
