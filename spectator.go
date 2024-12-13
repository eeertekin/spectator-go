package spectator

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var collection []string
var c *http.Client
var m sync.Mutex

func Start(interval int) {
	c = &http.Client{
		Timeout: time.Second,
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/tmp/spectatord/sock")
			},
		},
	}

	go watch(interval)
}

func Inc(key string, value int64) {
	m.Lock()
	defer m.Unlock()
	collection = append(collection, fmt.Sprintf("%s+=%d", key, value))
}

func Dec(key string, value int64) {
	m.Lock()
	defer m.Unlock()
	collection = append(collection, fmt.Sprintf("%s-=%d", key, value))
}

func Set(key string, value int64) {
	m.Lock()
	defer m.Unlock()
	collection = append(collection, fmt.Sprintf("%s=%d", key, value))
}

func Push() error {
	if len(collection) == 0 {
		return nil
	}
	m.Lock()
	payload := strings.Join(collection, "\n")
	collection = collection[:0]
	m.Unlock()

	req, err := http.NewRequest("POST", "http://localhost/collect", strings.NewReader(payload))
	if err != nil {
		log.Printf("spectator::push> %s\n", err)
		return err
	}

	_, err = c.Do(req)
	return err
}

func watch(interval int) {
	for range time.NewTicker(time.Duration(interval) * time.Second).C {
		Push()
	}
}