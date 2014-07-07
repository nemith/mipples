package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nemith/mipples/tinyurl"

	"github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
	rss "github.com/jteeuwen/go-pkg-rss"
)

func init() {
	module.Register("rss", &RSSModule{})
}

type RSSFeedConfig struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Channels []string `json:"channels"`
	Timeout  int      `json:"timeout"`
}

type RSSConfig struct {
	Timeout int              `"json:timeout"`
	Feeds   []*RSSFeedConfig `json:"feeds"`
}

type RSSModule struct {
	config *RSSConfig
}

// RSSLastSeen stores the last RFC seen to be persistant across instances of the
// bot
type RSSLastSeen struct {
	Id   int
	Feed string `sql:"not null;unique"`
	Key  string `sql:"not null"`
}

func (r RSSLastSeen) TableName() string {
	return "rss_lastseen"
}

func (m *RSSModule) Init(c *irc.Conn, config json.RawMessage) {
	db.AutoMigrate(&RSSLastSeen{})

	err := json.Unmarshal(config, &m.config)
	if err != nil {
		panic(err)
	}

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		// Wait a bit to lets things settle.  This should be rewritten
		// to check for channel memebership
		<-time.After(30 * time.Second)
		for _, feed := range m.config.Feeds {
			if feed.URL == "" {
				log.WithFields(logrus.Fields{
					"feed": feed,
				}).Error("RSS: Feed has no URL")
			}

			timeout := feed.Timeout
			if timeout < 1 {
				if m.config.Timeout < 1 {
					log.WithFields(logrus.Fields{
						"feed": feed,
					}).Error("RSS: Feed has no timeout or global timeout")
				}
				timeout = m.config.Timeout
			}

			go pollFeed(c, feed, timeout)
		}
	})
}

func pollFeed(conn *irc.Conn, feedConfig *RSSFeedConfig, timeout int) {
	itemHandler := func(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
		tag := feedConfig.Name
		if feedConfig.Name == "" {
			tag = ch.Title
		}

		// Dirty reverse
		reverseItems := make([]*rss.Item, len(newitems))
		for i, item := range newitems {
			reverseItems[len(newitems)-i-1] = item
		}

		rssLastSeen := RSSLastSeen{}
		db.Where(RSSLastSeen{Feed: feedConfig.URL}).FirstOrInit(&rssLastSeen)

		// This is hack.  Testing for gorm.RecordNotFound doesn't seem to be working with
		// FirstOrInit
		if rssLastSeen.Key == "" {
			log.WithFields(logrus.Fields{
				"feed": feedConfig.URL,
			}).Info("No last seen for feed.  Assuming new feed and not spamming the channel")
		} else {
			for _, item := range reverseItems {
				if rssLastSeen.Key == item.Key() {
					log.WithFields(logrus.Fields{
						"item":        item.Title,
						"lastseenkey": rssLastSeen.Key,
						"item_key":    item.Key(),
					}).Debug("RSS: Already seen this RSS item")
					break
				}

				// Tinyify item's url
				shortURL, err := tinyurl.Tinyify(item.Links[0].Href)
				if err != nil {
					log.WithFields(logrus.Fields{
						"longURL":  item.Links[0].Href,
						"shortURL": shortURL,
						"error":    err,
					}).Error("RSS: Failed to resolve tinyURL")
					shortURL = item.Links[0].Href
				}

				for _, channel := range feedConfig.Channels {
					conn.Privmsg(channel, fmt.Sprintf("[RSS %s] %s - %s", tag, item.Title, shortURL))
				}
			}
		}

		// Write the first item to the last seen database
		if rssLastSeen.Key != reverseItems[0].Key() {
			rssLastSeen.Key = reverseItems[0].Key()
			db.Save(&rssLastSeen)
		}
	}

	feed := rss.New(timeout, true, nil, itemHandler)

	for {
		log.WithFields(logrus.Fields{
			"feed": feedConfig.URL,
		}).Info("RSS: Fetching rss feed")

		if err := feed.Fetch(feedConfig.URL, nil); err != nil {
			log.WithFields(logrus.Fields{
				"feed":  feedConfig.URL,
				"error": err,
			}).Error("RSS: Error retriving feed.")
		}

		log.WithFields(logrus.Fields{
			"feed":              feedConfig.URL,
			"secondsTillUpdate": feed.SecondsTillUpdate(),
		}).Info("RSS: Waiting")

		<-time.After(time.Duration(feed.SecondsTillUpdate()) * time.Second)
	}
}
