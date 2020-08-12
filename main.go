package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mdp/qrterminal/v3"
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

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func TryListenPort(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return ln.Close()
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	var port int
	var host string
	flag.StringVar(&host, "h", "0.0.0.0", "`host` to serve on")
	flag.IntVar(&port, "p", -1, "`port` to listen on, defualt is 'auto'")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: goserver [-h host] [-p port] <path>\n")
		fmt.Fprintf(os.Stderr, "eg: \n")
		fmt.Fprintf(os.Stderr, "    goserver ./\n")
		fmt.Fprintf(os.Stderr, "    goserver -h 127.0.0.1 -p 8080 /path/to/folder/\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		if flag.NFlag() != 0 {
			fmt.Println("Did you forget to provide `path` argument?\n")
		}
		flag.Usage()
		return
	}
	directory := flag.Arg(0)
	http.Handle("/", NoCache(http.FileServer(http.Dir(directory))))
	ports := []int{5000, 8000, 5001, 8001}
	if port != -1 {
		ports = []int{port}
	}
	addr := ""
	for i, p := range ports {
		port = p
		addr = host + ":" + strconv.Itoa(port)
		if TryListenPort(addr) == nil {
			break
		}
		fmt.Printf("port %d is not available\n", port)
		if i == len(ports)-1 {
			log.Fatal("port bind failed")
		}
	}
	log.Printf("Serving %s on %s\n", directory, addr)
	ip := fmt.Sprintf("http://%s:%d/", GetOutboundIP().String(), port)
	qrterminal.Generate(ip, qrterminal.M, os.Stdout)
	fmt.Printf("Scan QR code above or visit %s\nWaiting for connections...", ip)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
}
