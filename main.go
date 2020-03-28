package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

func NoCache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		for _, v := range etagHeaders {
			if r.Header.Get(v) != "" {
				r.Header.Del(v)
			}
		}
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func main() {
	var port int
	var host string
	flag.StringVar(&host, "h", "0.0.0.0", "`host` to serve on")
	flag.IntVar(&port, "p", 80, "`port` to listen on")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: goserver [-h host] [-p port] <path>\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		return
	}
	directory := flag.Arg(0)
	http.Handle("/", NoCache(http.FileServer(http.Dir(directory))))
	log.Printf("Serving %s on %s:%d\n", directory, host, port)
	log.Fatal(http.ListenAndServe(host+":"+strconv.Itoa(port), nil))
}
