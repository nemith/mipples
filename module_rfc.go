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
	FetchInterval time.Duration `json:"fetch_interval"`
}

type RFCModule struct {
	config *RFCConfig
}

var rfcRegexp = regexp.MustCompile(`(?i)\b(RFC\d+)\b`)

func (m *RFCModule) Init(c *irc.Conn, config json.RawMessage) {
	err := json.Unmarshal(config, &m.config)
	if err != nil {
		panic(err)
	}

	go rfcFetchLoop(m.config.FetchInterval)
	c.HandleBG("PRIVMSG", NewMatchAllHandler(rfcRegexp, m.rfcHandler))
}

func (m *RFCModule) rfcHandler(conn *irc.Conn, msg *Privmsg, matches [][]string) {
	var rfcs []string
	for _, match := range matches {
		rfcs = append(rfcs, rfc.TrimDocID(strings.ToUpper(match[1])))
	}

	var rfcObjects []RFC
	db.Where("doc_id in (?)", rfcs).Find(&rfcObjects)

	if len(rfcObjects) < 1 {
		log.WithFields(logrus.Fields{
			"rfc": rfcs,
		}).Info("RFC Record not found in db")
		return
	}

	for _, rfcObject := range rfcObjects {
		msg.Respond(conn, "%s: %s (%s %d) [%s]  - %s",
			rfcObject.DocId, rfcObject.Title, rfcObject.Month, rfcObject.Year,
			rfcObject.Status, rfcObject.Url)
	}
}

func rfcFetchLoop(interval time.Duration) {
	for {
		rfcFetcher()
		time.Sleep(interval * time.Minute)
	}
}

type RFC struct {
	Id       int64
	DocId    string `sql:"not null;unique"`
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

func (k RFC) TableName() string {
	return "rfcs"
}

func rfcFetcher() error {
	log.Debug("Fetching RFC Index")

	rfcIndex, err := rfc.FetchAndParseIndex()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to fetch and parse rfc index xml")
	}

	tx := db.Begin()
	tx.DropTable(RFC{})
	tx.CreateTable(RFC{})

	for _, rfcEntry := range rfcIndex.RFCEntries {
		log.WithFields(logrus.Fields{
			"rfc": rfcEntry,
		}).Debug("Inserting RFC into database")
		rfcRow := RFC{
			DocId:    string(rfcEntry.DocID),
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
