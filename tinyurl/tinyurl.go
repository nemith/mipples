package tinyurl

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

func Tinyify(longURL string) (string, error) {
	client := http.Client{}
	apiURL := url.URL{
		Scheme:   "http",
		Host:     "tinyurl.com",
		Path:     "api-create.php",
		RawQuery: url.Values{"url": {longURL}}.Encode(),
	}
	resp, err := client.Get(apiURL.String())
	if err != nil {
		return "", err
	}
	shortURL, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(shortURL), nil
}
