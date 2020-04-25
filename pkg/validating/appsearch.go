package validating

import (
	"github.com/pkg/errors"
)

// ValidatedAppSearchRequest is a concrete representation of a validated crawl request
// for an AppSearch engine
type ValidatedAppSearchRequest struct {
	URL      string `json:"url"`
	OnDomain bool   `json:"onDomain"`
	Engine   string `json:"engine"`
	Domain   string `json:"domain,omitempty"`
}

// ValidateAppSearchRequest takes a CrawlRequest and returns a ValidatedAppSearchRequest
// or a slice of errors that came out of validations
func ValidateAppSearchRequest(r *CrawlRequest) (v *ValidatedAppSearchRequest, e []error) {
	if r.Engine == "" {
		message := errors.New("missing Engine in request")
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

	v = &ValidatedAppSearchRequest{
		URL:      r.URL,
		OnDomain: r.OnDomain,
		Engine:   r.Engine,
		Domain:   r.Domain,
	}

	return v, nil
}
