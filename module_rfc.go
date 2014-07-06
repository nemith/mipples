package main

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
	"github.com/nemith/mipples/rfc"
	"regexp"
	"strings"
	"time"
)

func init() {
	module.Register("rfc", &RFCModule{})
}

type RFCConfig struct {
	// Number of minutes between refreshing of the database
	FetchInterval int `json:"fetch_interval"`
}

type RFCModule struct{}

var rfcRegexp = regexp.MustCompile(`^(?i)(RFC\d+)$`)

func (k *RFCModule) Init(c *irc.Conn, config json.RawMessage) {
	go rfcFetchLoop(5)
	c.HandleBG("PRIVMSG", NewMatchHandler(rfcRegexp, k.rfcHandler))
}

func (k *RFCModule) rfcHandler(conn *irc.Conn, msg *Privmsg, match []string) {
	rfcObject := &RFC{}
	rfcNum := rfc.TrimDocID(strings.ToUpper(match[1]))
	db.Where(RFC{DocID: rfcNum}).First(rfcObject)

	if rfcObject == nil {
		return
	}

	msg.Respond(conn, "%s - %s [%s] (%s %d) - %s",
		rfcObject.DocID, rfcObject.Title, rfcObject.Status,
		rfcObject.Month, rfcObject.Year, rfcObject.Url)
}

func rfcFetchLoop(interval time.Duration) {
	for {
		rfcFetcher()
		time.Sleep(interval * time.Minute)
	}
}

type RFC struct {
	Id       int64
	DocID    string `sql:"not null;unique"`
	Title    string
	Abstract string
	Month    string
	Year     int
	Status   string
	Url      string
	//	Obsoletes   RFC
	//	ObsoletedBy RFC
	//	UpdatedBy   RFC
}

func rfcFetcher() error {
	log.Debug("Fetching RFC Index")
	xml, err := rfc.FetchRFCIndex()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to retreive rfc index")
		return err
	}
	defer xml.Close()

	rfcIndex, err := rfc.ParseRFCIndex(xml)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to parse rfc index xml")
	}

	tx := db.Begin()
	tx.DropTable(RFC{})
	tx.CreateTable(RFC{})

	for _, rfcEntry := range rfcIndex.RFCEntries {
		log.WithFields(logrus.Fields{
			"rfc": rfcEntry,
		}).Debug("Fetching RFC Index")
		rfcRow := RFC{
			DocID:    string(rfcEntry.DocID),
			Title:    rfcEntry.Title,
			Abstract: string(rfcEntry.Abstract),
			Month:    rfcEntry.Month,
			Year:     rfcEntry.Year,
			Status:   rfcEntry.CurrentStatus,
			Url:      rfcEntry.HTMLURL(),
		}
		tx.Save(&rfcRow)
	}

	tx.Commit()

	return nil
}
