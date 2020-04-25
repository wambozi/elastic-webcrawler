package validating

import (
	"github.com/pkg/errors"
)

// ValidatedElasticsearchRequest is a concrete representation of a validated crawl request
// for an Elasticsearch engine
type ValidatedElasticsearchRequest struct {
	URL      string `json:"url"`
	OnDomain bool   `json:"onDomain"`
	Index    string `json:"index"`
	Domain   string `json:"domain,omitempty"`
}

// ValidateElasticsearchRequest takes a CrawlRequest and returns a ValidatedElasticsearchRequest
// or a slice of errors that came out of validations
func ValidateElasticsearchRequest(r *CrawlRequest) (v *ValidatedElasticsearchRequest, e []error) {
	if r.Index == "" {
		message := errors.New("missing Index in request")
		e = append(e, message)
	}

	if r.URL == "" {
		message := errors.New("missing URL in request")
		e = append(e, message)
	}

	if r.Domain == "" {
		message := errors.New("missing Domain in request")
		e = append(e, message)
	}

	if len(e) > 0 {
		return nil, e
	}

	v = &ValidatedElasticsearchRequest{
		URL:      r.URL,
		OnDomain: r.OnDomain,
		Index:    r.Index,
		Domain:   r.Domain,
	}

	return v, nil
}
