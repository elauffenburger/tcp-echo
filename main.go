package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	l := log.New(os.Stdout, "[tcp-echo] ", log.Flags())
	ctx := context.Background()

	addr := os.Getenv("TCP_ECHO_ADDR")
	if addr == "" {
		addr = ":"
	}

	fixedResponse := os.Getenv("TCP_ECHO_RESPONSE")

	srv := newServer(addr, &fixedResponse, l)
	srv.listenAndServe(ctx)
}

type server struct {
	l *log.Logger

	tcpAddr          string
	tcpFixedResponse *string
}

func newServer(tcpAddr string, tcpFixedResponse *string, l *log.Logger) *server {
	return &server{l, tcpAddr, tcpFixedResponse}
}

func (s *server) listenAndServe(ctx context.Context) {
	go s.startTCP(ctx)
	go s.startHTTP(ctx)

	<-ctx.Done()
}

func (s *server) startTCP(_ context.Context) {
	srv, err := net.Listen("tcp", s.tcpAddr)
	if err != nil {
		panic(err)
	}

	s.l.Printf("TCP listening on %s\n", srv.Addr())

	for {
		conn, err := srv.Accept()
		if err != nil {
			s.l.Printf("error accepting connection: %s\n", err)
			continue
		}

		s.l.Println("received connection")

		go func() {
			rdr := bufio.NewReader(conn)

			str, err := rdr.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				s.l.Printf("error reading payload: %s\n", err)
				return
			}

			s.l.Printf("%s", str)

			var response string
			if s.tcpFixedResponse != nil && *s.tcpFixedResponse != "" {
				response = *s.tcpFixedResponse
			} else {
				response = str
			}

			n, err := io.Copy(conn, strings.NewReader(response))
			if err != nil {
				s.l.Printf("error writing payload: %s\n", err)
				return
			}

			if n != int64(len(response)) {
				s.l.Printf("only wrote %d/%d bytes\n", n, len(response))
				return
			}

			if err := conn.Close(); err != nil {
				s.l.Printf("error closing conn: %s\n", err)
				return
			}
		}()
	}
}

func (s *server) startHTTP(_ context.Context) {
	srv := http.NewServeMux()
	srv.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))

	addr := s.httpAddr()
	s.l.Printf("http listening on %s", addr)

	http.ListenAndServe(addr, srv)
}

func (s *server) httpAddr() string {
	addrParts := strings.SplitN(s.tcpAddr, ":", 2)

	var host string
	if len(addrParts) > 0 {
		host = addrParts[0]
	} else {
		host = ""
	}

	return fmt.Sprintf("%s:80", host)
}
