package start

import "fmt"

// BuildAccessURLs returns clickable URLs for the UI.
// If enableHTTP is true, both HTTPS and HTTP URLs are returned.
func BuildAccessURLs(httpsPort, httpPort int, bind string, enableHTTP bool) []string {
	host := bind
	if host == "127.0.0.1" || host == "0.0.0.0" {
		host = "localhost"
	}

	urls := []string{
		fmt.Sprintf("https://%s:%d", host, httpsPort),
	}
	if enableHTTP {
		urls = append(urls, fmt.Sprintf("http://%s:%d", host, httpPort))
	}
	return urls
}
