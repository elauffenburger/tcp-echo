package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	l := log.New(os.Stdout, "[tcp-echo] ", log.Flags())

	addr := os.Getenv("TCP_ECHO_ADDR")
	if addr == "" {
		addr = ":"
	}

	srv, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	fixedResponse := os.Getenv("TCP_ECHO_RESPONSE")

	l.Printf("listening on %s\n", srv.Addr())

	for {
		conn, err := srv.Accept()
		if err != nil {
			l.Printf("error accepting connection: %s\n", err)
			continue
		}

		l.Println("received connection")

		go func() {
			rdr := bufio.NewReader(conn)

			str, err := rdr.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				l.Printf("error reading payload: %s\n", err)
				return
			}

			l.Printf("%s", str)

			var response string
			if fixedResponse != "" {
				response = fixedResponse
			} else {
				response = str
			}

			n, err := io.Copy(conn, strings.NewReader(response))
			if err != nil {
				l.Printf("error writing payload: %s\n", err)
				return
			}

			if n != int64(len(response)) {
				l.Printf("only wrote %d/%d bytes\n", n, len(response))
				return
			}

			if err := conn.Close(); err != nil {
				l.Printf("error closing conn: %s\n", err)
				return
			}
		}()
	}
}
