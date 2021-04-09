package main

/*
*** Attribution - Do not remove ****

Based on work [1] by Michał Łowicki, licensed under CC BY 4.0 [2]

[1] https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c
[2] https://creativecommons.org/licenses/by/4.0/

*/

import (
	"bytes"
	"crypto/tls"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
func main() {

	var proxyAddr string
	flag.StringVar(&proxyAddr, "addr", "localhost:8888", "listen on <ip>:<port>")
	flag.Parse()

	server := &http.Server{
		Addr: proxyAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Printf("===================")
			//log.Printf("original request: %+v\n", r)
			log.Printf("📨\toriginal url: \t %s", string(r.URL.String()))

			origURL := r.RequestURI

			if strings.HasPrefix(origURL, "http://") {

				newURL := strings.Replace(origURL, "http", "https", 1)
				newURL = strings.Replace(newURL, ":80", ":443", 1)

				newBody, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println(err)
				}
				r.Body = ioutil.NopCloser(bytes.NewReader(newBody))

				newR, err := http.NewRequestWithContext(r.Context(), r.Method, newURL, bytes.NewReader(newBody))
				if err != nil {
					log.Println(err)
				}
				newR.Header = make(http.Header)
				for h, val := range r.Header {
					newR.Header[h] = val
				}

				r = newR

				//log.Printf("new request: %+v\n", r)
				log.Printf("✏️\tnew url: \t %s", string(r.URL.String()))
			} else {
				log.Printf("🚫\trequest isn't http, so url not modified")
			}

			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("starting proxy server on %s ...", proxyAddr)
	log.Fatal(server.ListenAndServe())
}
