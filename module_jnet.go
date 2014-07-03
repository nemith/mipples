package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/nemith/mipples/jnet"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

func init() {
	module.Register("jnet", &JNetModule{})
}

type JNetConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JNetModule struct {
	config *JNetConfig
}

var prRegexp = regexp.MustCompile(`(?i)(pr[0-9]+)`)

func (m *JNetModule) Init(i *Irc, config json.RawMessage) {
	i.AddMatch(prRegexp, m.prHandler)

	err := json.Unmarshal(config, &m.config)
	if err != nil {
		panic(err)
	}
}

func (m *JNetModule) prHandler(conn *irc.Conn, msg *Privmsg, match []string) {
	jnet := jnet.NewJNet(m.config.Username, m.config.Password)
	pr, err := jnet.GetPR(match[1])
	if err != nil {
		return
	}

	url, err := tinyURL(pr.URL)
	if err != nil {
		// Cannot get TinyURL. Return the ugly one.
		url = pr.URL
	}

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s (%s) [%s] %s", pr.Number, pr.Status, pr.Severity, pr.Title))

	if pr.ResolvedIn != "" {
		buf.WriteString(fmt.Sprintf(" Fixed in: %s", pr.ResolvedIn))
	}
	buf.WriteString(" - ")
	buf.WriteString(url)

	msg.Respond(conn, buf.String())
}

func tinyURL(longURL string) (string, error) {
	client := http.Client{}
	url := url.URL{
		Scheme:   "http",
		Host:     "tinyurl.com",
		Path:     "api-create.php",
		RawQuery: url.Values{"url": {longURL}}.Encode(),
	}
	resp, err := client.Get(url.String())
	if err != nil {
		return "", err
	}
	shortURL, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(shortURL), nil
}