package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}
type simpleservererver struct {
	addr  string
	proxy *httputil.ReverseProxy
}
type Loadbalancer struct {
	port            string
	roundrobincount int
	servers         []Server
}

func NewLoadbalancer(port string, servers []Server) *Loadbalancer {
	return &Loadbalancer{
		port:            port,
		roundrobincount: 0,
		servers:         servers,
	}
}
func handlerr(err error) {
	if err != nil {
		fmt.Println("error in parse", err.Error())
		os.Exit(1)
	}
}

func newserver(addr string) *simpleservererver {
	server, err := url.Parse(addr)
	handlerr(err)
	return &simpleservererver{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(server),
	}
}
func (s *simpleservererver) Address() string {
	return s.addr
}
func (s *simpleservererver) IsAlive() bool {
	return true
}
func (s *simpleservererver) Serve(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}
func (lb *Loadbalancer) getnextavailableserver() Server {
	server := lb.servers[lb.roundrobincount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundrobincount++
		server = lb.servers[lb.roundrobincount%len(lb.servers)]
	}
	lb.roundrobincount++
	return server
}
func (lb *Loadbalancer) serveproxy(rw http.ResponseWriter, r *http.Request) {
	targetserver := lb.getnextavailableserver()
	fmt.Printf("forwarding request to address %q\n", targetserver.Address())
	targetserver.Serve(rw, r)
}

func main() {
	servers := []Server{newserver("https://www.facebook.com"), newserver("https://www.bing.com"), newserver("https://www.duckduckgo.com")}
	lb := NewLoadbalancer("8000", servers)
	handleredirect := func(rw http.ResponseWriter, r *http.Request) {
		lb.serveproxy(rw, r)
	}
	http.HandleFunc("/", handleredirect)
	fmt.Printf("serving request at %s\n", lb.port)
	http.ListenAndServe()
}
