package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter URL: ")
	link, _ := reader.ReadString('\n')

	for {
		proxyList, _ := ioutil.ReadFile("data/prox.txt")
		proxies := strings.Split(string(proxyList), "\n")

		var validProxies []string

		var wg sync.WaitGroup

		for _, proxy := range proxies {
			proxy = strings.Trim(proxy, "\r")
			if len(proxy) > 0 {
				wg.Add(1)
				go func(proxy string) {
					defer wg.Done()
					proxy = addHTTPPrefix(proxy)
					proxyURL, err := url.Parse(proxy)
					if err != nil {
						fmt.Println(err)
						return
					}
					httpClient := &http.Client{
						Transport: &http.Transport{
							Proxy: http.ProxyURL(proxyURL),
							TLSClientConfig: &tls.Config{
								InsecureSkipVerify: true,
							},
						},
						Timeout: time.Second * 10,
					}

					resp, err := httpClient.Get(strings.TrimSpace(link))
					if err != nil {
						fmt.Println("Error:", err)
						proxy = switchProtocol(proxy)
						return
					}
					resp.Body.Close()
					fmt.Println("GET request sent through proxy:", proxy)
					validProxies = append(validProxies, proxy)
				}(proxy)
			}
		}
		wg.Wait()

		// Update the proxies slice to only include the valid proxies
		proxies = validProxies
	}
}

func addHTTPPrefix(proxy string) string {
	if !strings.HasPrefix(proxy, "http://") && !strings.HasPrefix(proxy, "https://") {
		proxy = "http://" + proxy
	}
	return proxy
}

func switchProtocol(proxy string) string {
	if strings.HasPrefix(proxy, "http://") {
		return strings.Replace(proxy, "http://", "https://", 1)
	} else if strings.HasPrefix(proxy, "https://") {
		return strings.Replace(proxy, "https://", "http://", 1)
	}
	return proxy
}
