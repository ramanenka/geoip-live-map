package main

import (
	"context"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
	maxminddb "github.com/oschwald/maxminddb-golang"
)

type broadcaster struct {
	mu sync.Mutex
	cs []chan<- interface{}
}

func (b *broadcaster) sub(c chan<- interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cs = append(b.cs, c)
}

func (b *broadcaster) usub(c chan<- interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.cs[:0]
	for _, x := range b.cs {
		if x != c {
			n = append(n, x)
		}
	}
	b.cs = n
}

func (b *broadcaster) pub(v interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, c := range b.cs {
		c <- v
	}
}

func (b *broadcaster) close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, c := range b.cs {
		close(c)
	}
}

type mmrecord struct {
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
}

var b *broadcaster

func main() {
	logFilename := os.Getenv("LOG_FILENAME")
	if logFilename == "" {
		log.Fatal("LOG_FILENAME env variable is required")
	}

	gdb, err := maxminddb.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatalf("failed to open maxmind db: %v", err)
	}
	defer gdb.Close()

	t, err := tail.TailFile(logFilename, tail.Config{
		Follow:   true,
		Location: &tail.SeekInfo{Whence: os.SEEK_END},
	})
	if err != nil {
		log.Fatalf("failed to tail %s: %v", "ips.log", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws)
	mux.HandleFunc("/", index)
	httpListenOn := os.Getenv("HTTP_LISTEN_ON")
	if httpListenOn == "" {
		httpListenOn = ":8080"
	}
	server := http.Server{Addr: httpListenOn, Handler: mux}

	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()

		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	b = &broadcaster{}

	go func() {
		wg.Add(1)
		defer wg.Done()

		ipregexp := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)

		for line := range t.Lines {
			ipstr := ipregexp.FindString(line.Text)
			if ipstr == "" {
				log.Printf("failed to find IP addres in: %s", line.Text)
				continue
			}
			ip := net.ParseIP(ipstr)

			res := mmrecord{}
			if err := gdb.Lookup(ip, &res); err != nil {
				log.Printf("failed to lookup ip %s location: %v", ipstr, err)
				continue
			}
			b.pub([]float64{res.Location.Latitude, res.Location.Longitude})
		}
	}()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
	log.Println("interrupted, shutting down the server")
	if err := t.Stop(); err != nil {
		log.Println(err)
	}
	if err := server.Shutdown(context.Background()); err != nil {
		log.Println(err)
	}
	wg.Wait()
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := struct{ Host string }{Host: r.Host}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

var upgrader = websocket.Upgrader{}

func ws(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer conn.Close()

	c := make(chan interface{})
	b.sub(c)
	defer b.usub(c)

	for v := range c {
		if err := conn.WriteJSON(v); err != nil {
			log.Println(err)
			break
		}
	}
}
