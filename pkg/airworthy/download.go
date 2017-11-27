package airworthy

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/jetstack/airworthy/pkg/cacertificates"
)

func (a *Airworthy) HTTPClient() *http.Client {
	if a.httpClient == nil {
		rootCAs, err := cacertificates.Roots()
		if err != nil {
			panic(fmt.Sprint("error gettting root CAs: ", err))
		}

		httpTransport := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		}

		a.httpClient = &http.Client{
			Transport: httpTransport,
		}
	}

	return a.httpClient
}

func (a *Airworthy) Download(url string) (*Progress, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.HTTPClient().Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 400 {
		return nil, fmt.Errorf("unexpected response code %d for %s (GET)", resp.StatusCode, url)
	}

	return &Progress{
		Reader: resp.Body,
		Size:   resp.ContentLength,
	}, nil
}
