package validating

// CrawlRequest represents the request to the /crawl route
type CrawlRequest struct {
	Index    string `json:"index,omitempty"`
	URL      string `json:"url"`
	OnDomain bool   `json:"onDomain"`
	Engine   string `json:"engine,omitempty"`
	Domain   string `json:"domain,omitempty"`
}
