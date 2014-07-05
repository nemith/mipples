package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
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

func (m *JNetModule) Init(c *irc.Conn, config json.RawMessage) {
	c.HandleBG("PRIVMSG", NewMatchHandler(prRegexp, m.prHandler))

	err := json.Unmarshal(config, &m.config)
	if err != nil {
		panic(err)
	}
}

func (m *JNetModule) prHandler(conn *irc.Conn, msg *Privmsg, match []string) {
	prNumber := match[1]

	tinyUrlChan := make(chan string)
	go func(tinyUrlChan chan string) {
		longUrl := jnet.PRUrl(prNumber)
		shortUrl, err := tinyURL(longUrl)
		if err != nil {
			log.WithFields(logrus.Fields{
				"longUrl":  longUrl,
				"shortUrl": shortUrl,
				"pr":       prNumber,
				"error":    err,
			}).Error("Jnet: Failed to resolve tinyURL")
			tinyUrlChan <- longUrl
		}
		tinyUrlChan <- shortUrl
	}(tinyUrlChan)

	prChan := make(chan *jnet.JNetPR)
	go func(prChan chan *jnet.JNetPR) {
		j := jnet.NewJNet(m.config.Username, m.config.Password)
		pr, err := j.GetPR(prNumber)
		if err != nil {
			log.WithFields(logrus.Fields{
				"pr":    prNumber,
				"error": err,
			}).Error("Jnet: Failed to get PR information")
			prChan <- nil
		}
		prChan <- pr
	}(prChan)

	// Wait for reponses
	shortUrl := <-tinyUrlChan
	pr := <-prChan

	if pr == nil {
		msg.RespondToNick(conn, "Failed to find PR information for %s - %s", prNumber, shortUrl)
	}

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s (%s) [%s] %s", pr.Number, pr.Status, pr.Severity, pr.Title))

	if pr.ResolvedIn != "" {
		buf.WriteString(fmt.Sprintf(" Fixed in: %s", pr.ResolvedIn))
	}
	buf.WriteString(" - ")
	buf.WriteString(shortUrl)

	msg.Respond(conn, buf.String())
}

func tinyURL(longURL string) (string, error) {
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
