package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	cacheTTL = 5 * time.Second
	cache = make(map[string]NsCache)
	done := CacheChecker()
	InputManager(done)
}

func InputManager(done chan bool) {
	for {
		var input string
		fmt.Print("Enter domain (enter ':q' for quit) : ")
		_, err := fmt.Scanln(&input)
		if err != nil {
			return
		}
		if strings.Compare(input, ":q") == 0 {
			done <- true
			return
		}
		NsResolver(input)
	}
}

func NsResolver(domain string) {
	var nsIPs []net.IP
	val, ok := CacheGet(domain)
	if ok {
		fmt.Printf("Got %s IPs from cache\n", domain)
		nsIPs = val.IPs
	} else {
		fmt.Printf("Got %s IPs from resolver\n", domain)
		ips, err := net.LookupIP(domain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		}
		nsIPs = ips
		nsCache := NsCache{
			IPs:       ips,
			CreatedAt: time.Now(),
		}
		CacheSet(domain, nsCache)
	}

	for _, ip := range nsIPs {
		fmt.Printf("%s IN A %s\n", domain, ip.String())
	}
}

type NsCache struct {
	IPs       []net.IP
	CreatedAt time.Time
}

var cache map[string]NsCache

var cacheTTL time.Duration

func CacheSet(domain string, val NsCache) {
	cache[domain] = val
}

func CacheGet(domain string) (NsCache, bool) {
	val, ok := cache[domain]
	return val, ok
}

func CacheChecker() chan bool {
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				CacheInvalidator()
			}
		}
	}()

	return done
}

func CacheInvalidator() {
	for k, v := range cache {
		if v.CreatedAt.Add(cacheTTL).Before(time.Now()) {
			delete(cache, k)
		}
	}
}
