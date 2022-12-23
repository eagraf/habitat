package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	gostream "github.com/libp2p/go-libp2p-gostream"
	p2phttp "github.com/libp2p/go-libp2p-http"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/rs/zerolog/log"
)

func LibP2PHTTPProxy(host host.Host, redirectURL *url.URL) {
	listener, _ := gostream.Listen(host, p2phttp.DefaultP2PProtocol)
	defer listener.Close()

	handler := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = redirectURL.Scheme
			req.URL.Host = redirectURL.Host
			req.URL.Path = redirectURL.Path + req.URL.Path
		},
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
		},
		ModifyResponse: func(res *http.Response) error {
			return nil
		},
		ErrorHandler: func(rw http.ResponseWriter, r *http.Request, err error) {
			log.Error().Err(err).Msgf("libp2p reverse proxy request forwarding error. request to: %s", r.URL.String())
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		},
	}

	server := http.Server{
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Msgf("LibP2P proxy listening on %s", host.Addrs()[0])
	err := server.Serve(listener)
	log.Fatal().Err(err)
}
