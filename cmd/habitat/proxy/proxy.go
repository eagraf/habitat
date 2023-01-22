package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/configuration"
	"github.com/rs/zerolog/log"
)

type RuleSet map[string]Rule

var Hostname string

func init() {
	Hostname = compass.Hostname()
}

type Server struct {
	Rules RuleSet
}

func NewServer() *Server {
	return &Server{
		Rules: make(RuleSet),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, rule := range s.Rules {
		if rule.Match(r.URL) {
			rule.Handler().ServeHTTP(w, r)
			return
		}
	}
	// No rules matched
	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) Start(host string) {
	log.Fatal().Err(http.ListenAndServe(host, s)).Msg("reverse proxy server failed")
}

func (r RuleSet) Add(name string, rule Rule) error {
	if _, ok := r[name]; ok {
		return fmt.Errorf("rule name %s is already taken", name)
	}
	r[name] = rule
	return nil
}

func (r RuleSet) Remove(name string) error {
	if _, ok := r[name]; !ok {
		return fmt.Errorf("rule %s does not exist", name)
	}
	delete(r, name)
	return nil
}

type Rule interface {
	Match(url *url.URL) bool
	Handler() http.Handler
}

type FileServerRule struct {
	Matcher string
	Path    string
}

func (r *FileServerRule) Match(url *url.URL) bool {
	// TODO make this work with actual glob strings
	// For now, just match based off of base path
	return strings.HasPrefix(url.Path, r.Matcher)
}

func (r *FileServerRule) Handler() http.Handler {
	return &FileServerHandler{
		Prefix: r.Matcher,
		Path:   r.Path,
	}
}

type FileServerHandler struct {
	Prefix string
	Path   string
}

func (h *FileServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try to remove prefix
	oldPath := r.URL.Path
	r.URL.Path = strings.TrimPrefix(oldPath, h.Prefix)

	if oldPath == r.URL.Path {
		// Something weird happened
		w.Write([]byte("unable to remove url path prefix"))
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.FileServer(http.Dir(h.Path)).ServeHTTP(w, r)
}

type RedirectRule struct {
	Matcher         string
	ForwardLocation *url.URL
}

func (r *RedirectRule) Match(url *url.URL) bool {
	// TODO make this work with actual glob strings
	// For now, just match based off of base path
	return strings.HasPrefix(url.Path, r.Matcher)
}

func (r *RedirectRule) Handler() http.Handler {
	host, port, _ := net.SplitHostPort(r.ForwardLocation.Host)
	target := r.ForwardLocation.Host
	if host == "0.0.0.0" {
		target = fmt.Sprintf("%s:%s", Hostname, port)
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = r.ForwardLocation.Scheme
			req.URL.Host = target
			req.URL.Path = strings.TrimPrefix(req.URL.Path, r.Matcher) // TODO this needs to be fixed when globs are implemented
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
			log.Error().Err(err).Msgf("reverse proxy request forwarding error. request to %s", r.URL.String())
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		},
	}
}

func GetRuleFromConfig(config *configuration.ProxyRule, appPath string) (Rule, error) {
	switch config.Type {
	case configuration.ProxyRuleFileServer:
		return &FileServerRule{
			Matcher: config.Matcher,
			Path:    filepath.Join(appPath, config.Target),
		}, nil
	case configuration.ProxyRuleRedirect:
		targetURL, err := url.Parse(config.Target)
		if err != nil {
			return nil, fmt.Errorf("error parsing url for RedirectRule: %s", err)
		}
		return &RedirectRule{
			Matcher:         config.Matcher,
			ForwardLocation: targetURL,
		}, nil
	default:
		return nil, fmt.Errorf("no proxy rule type %s", config.Type)
	}
}
