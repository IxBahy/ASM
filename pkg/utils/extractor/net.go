package extractor

import "strings"

func ExtractDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")

	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

func ExtractHostAndPort(url string) (host string, port string) {
	port = ""

	if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
		port = "443"
	} else if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
		port = "80"
	}

	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	if idx := strings.Index(url, ":"); idx != -1 {
		host = url[:idx]
		port = url[idx+1:]
	} else {
		host = url
	}
	return host, port
}
