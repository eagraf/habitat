package main

import (
	"fmt"
	"io/ioutil"
	"net"

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
	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		return err
	}

	fmt.Println(string(buf))

	return nil
}
