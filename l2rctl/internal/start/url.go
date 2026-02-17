package start

import "fmt"

// BuildAccessURLs returns clickable URLs for the UI.
// If enableHTTP is true, both HTTPS and HTTP URLs are returned.
func BuildAccessURLs(httpsPort, httpPort int, bind string, enableHTTP bool, user, pass string) []string {
	host := bind
	if host == "127.0.0.1" || host == "0.0.0.0" {
		host = "localhost"
	}

	var userinfo string
	if user != "" && pass != "" {
		userinfo = fmt.Sprintf("%s:%s@", user, pass)
	}

	urls := []string{
		fmt.Sprintf("https://%s%s:%d", userinfo, host, httpsPort),
	}
	if enableHTTP {
		urls = append(urls, fmt.Sprintf("http://%s%s:%d", userinfo, host, httpPort))
	}
	return urls
}
