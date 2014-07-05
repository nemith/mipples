package jnet

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type JNet struct {
	Username string
	Password string
	client   *http.Client
}

func NewJNet(username, password string) *JNet {
	j := &JNet{
		Username: username,
		Password: password,
	}
	j.Login()
	return j
}

func (j *JNet) Login() error {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	j.client = &http.Client{
		Jar: cookieJar,
	}

	resp, err := j.client.Get("https://www.juniper.net/customers/support/")
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		panic(err)
	}
	pageGenTime, _ := doc.Find(`input[name="pageGenTime"]`).First().Attr("value")

	resp, err = j.client.PostForm("https://iam-fed.juniper.net/jlogin/SignOn.jsp", url.Values{
		"userid":       {j.Username},
		"PASSWORD":     {j.Password},
		"LOCALE":       {"en_us"},
		"AUTHMETHOD":   {"UserPassword"},
		"pageGenTime":  {pageGenTime},
		"target":       {"http://www.juniper.net/customers/support/"},
		"smauthreason": {"1"},
	})
	if err != nil {
		return err
	}
	return nil
}

type JNetPR struct {
	Number         string
	Title          string
	ReleaseNote    string
	Severity       string
	Status         string
	LastModified   string
	ResolvedIn     string
	OS             string
	Product        string
	FunctionalArea string
	URL            string
}

func getPRTableRow(table *goquery.Selection, column string) string {
	selectorText := fmt.Sprintf(`td:contains("%s")`, column)
	return strings.TrimSpace(table.Find(selectorText).NextFiltered("td").Text())
}

func (j *JNet) GetPR(prNumber string) (*JNetPR, error) {
	prNumber = strings.ToUpper(prNumber)
	query := url.Values{
		"page": {"prcontent"},
		"id":   {prNumber},
	}
	url := url.URL{
		Scheme:   "https",
		Host:     "prsearch.juniper.net",
		Path:     "InfoCenter/index",
		RawQuery: query.Encode(),
	}

	pr := &JNetPR{
		URL: url.String(),
	}

	resp, err := j.client.Get(url.String())
	if err != nil {
		return pr, nil
	}
	doc, err := goquery.NewDocumentFromResponse(resp)

	prTable := doc.Find(`table[summary="prsearch results"]`)

	if prTable.Length() < 1 {
		return nil, fmt.Errorf("No PR found")
	}

	pr.Number = getPRTableRow(prTable, "Number")
	pr.Title = getPRTableRow(prTable, "Title")
	pr.ReleaseNote = getPRTableRow(prTable, "Release Note")
	pr.Severity = getPRTableRow(prTable, "Severity")
	pr.Status = getPRTableRow(prTable, "Status")
	pr.LastModified = getPRTableRow(prTable, "Last Modified")
	pr.ResolvedIn = getPRTableRow(prTable, "Resolved In")
	pr.OS = getPRTableRow(prTable, "Operating System")
	pr.Product = getPRTableRow(prTable, "Product")
	pr.FunctionalArea = getPRTableRow(prTable, "Functional Area")

	return pr, nil
}
