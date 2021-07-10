package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"github.com/eagraf/habitat/structs/ctl"
	"github.com/rs/zerolog/log"
)

const (
	ListenerHost = "localhost"
	ListenerPort = "2040"
)

func main() {
	// TODO make port number configurable
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ListenerHost, ListenerPort))
	if err != nil {
		log.Fatal().Msgf("habitat service listener failed to start: %s", err)
	}
	defer listener.Close()

	log.Info().Msgf("habitat service listening on %s:%s", ListenerHost, ListenerPort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Msgf("error accepting message: %s", err)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) error {
	buf, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return err
	}

	fmt.Println(string(buf))

	marshalled, err := json.Marshal(&ctl.Response{Status: 0})
	if err != nil {
		return err
	}

	// base64 encode to make sure newlines are not present in bytes sent
	encoded := base64.StdEncoding.EncodeToString(marshalled)

	msg := append([]byte(encoded), '\n')

	_, err = conn.Write(msg)
	if err != nil {
		return err
	}

	conn.Close()

	return nil
}
