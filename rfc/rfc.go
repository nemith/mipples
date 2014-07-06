package rfc

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	RFCIndexURL = "http://www.rfc-editor.org/rfc/rfc-index.xml"
)

type DocID string

// For whatever reason the RFC numbers from IETF is zero left padded. Let's
// remove them
func (i *DocID) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp string
	d.DecodeElement(&temp, &start)
	*i = DocID(TrimDocID(temp))
	return nil
}

type Abstract string

// Stupid hack to get innerxml without defining a new struct with a single
// member due to the fact that Abstract field has <p> HTML element
func (a *Abstract) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	temp := struct {
		Text string `xml:",innerxml"`
	}{}
	d.DecodeElement(&temp, &start)
	*a = Abstract(temp.Text)
	return nil
}

type BCPEntry struct {
	XMLName xml.Name `xml:"bcp-entry"`
	DocID   DocID    `xml:"doc-id"`
	IsAlso  []DocID  `xml:"is-also>doc-id"`
}

type FYIEntry struct {
	XMLName xml.Name `xml:"fyi-entry"`
	DocID   DocID    `xml:"doc-id"`
	IsAlso  []DocID  `xml:"is-also>doc-id"`
}

type RFCEntry struct {
	XMLName  xml.Name `xml:"rfc-entry"`
	DocID    DocID    `xml:"doc-id"`
	Title    string   `xml:"title"`
	Abstract Abstract `xml:"abstract"`
	Authors  []string `xml:"author>name"`
	Month    string   `xml:"date>month"`
	Year     int      `xml:"date>year"`
	Formats  []struct {
		FileFormat string `xml:"file-format"`
		CharCount  int    `xml:"char-count"`
		PageCount  int    `xml:"page-count"`
	} `xml:"format"`
	CurrentStatus string   `xml:"current-status"`
	PubStatus     string   `xml:"publication-status"`
	Stream        string   `xml:"stream"`
	ErrataURL     string   `xml:"errata-url"`
	Area          string   `xml:"area"`
	WorkingGroup  string   `xml:"wg_acronym"`
	Draft         string   `xml:"draft"`
	Keywords      []string `xml:"keywords>kw"`
	Obsoletes     []DocID  `xml:"obsoletes>doc-id"`
	ObsoletedBy   []DocID  `xml:"obsoleted-by>doc-id"`
	UpdatedBy     []DocID  `xml:"updated-by>doc-id"`
}

func (rfc *RFCEntry) HTMLURL() string {
	return "http://tools.ietf.org/html/" + strings.ToLower(string(rfc.DocID))
}

func (rfc *RFCEntry) TextURL() string {
	return "http://tools.ietf.org/rfc/" + strings.ToLower(string(rfc.DocID))
}

func (rfc *RFCEntry) PDFURL() string {
	return "http://tools.ietf.org/pdf/" + strings.ToLower(string(rfc.DocID)) + ".pdf"
}

func (rfc *RFCEntry) String() string {
	return fmt.Sprintf("%s: %s [%s %d]", rfc.DocID, rfc.Title, rfc.Month, rfc.Year)
}

type STDEntry struct {
	XMLName xml.Name `xml:"std-entry"`
	DocID   DocID    `xml:"doc-id"`
	Title   string   `xml:"title"`
	IsAlso  []DocID  `xml:"is-also>doc-id"`
}

type RFCIndex struct {
	XMLName    xml.Name    `xml:"rfc-index"`
	BCPEntries []*BCPEntry `xml:"bcp-entry"`
	FYIEntries []*FYIEntry `xml:"fyi-entry"`
	RFCEntries []*RFCEntry `xml:"rfc-entry"`
	STDEntries []*STDEntry `xml:"std-entry"`
}

// TrimDocID removes any zero padding added in an RFC number. For example if
// you have RFC0010 return RFC10
func TrimDocID(input string) string {
	numStart := strings.IndexAny(input, "0123456789")
	return input[:numStart] + strings.TrimLeft(input[numStart:], "0")
}

// FetchIndex will attempt to retreive the XML RFC Index document from
// the ietf website.  Returns an io.ReadCloser of the XML document.
func FetchIndex() (io.ReadCloser, error) {
	client := http.Client{}
	resp, err := client.Get("http://www.rfc-editor.org/rfc/rfc-index.xml")
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ParseIndex takes an io.Reader and will parse the RFC index XML from it
// and return a parsed RFCIndex
func ParseIndex(rawInput io.Reader) (*RFCIndex, error) {
	index := &RFCIndex{}
	decoder := xml.NewDecoder(rawInput)
	if err := decoder.Decode(index); err != nil {
		return nil, err
	}
	return index, nil
}

// FetchAndParseIndex is a convience function to fetch a new copy index from
// the IETF website and parse the XML and return it
func FetchAndParseIndex() (*RFCIndex, error) {
	xmlBody, err := FetchIndex()
	if err != nil {
		return nil, err
	}
	defer xmlBody.Close()

	index, err := ParseIndex(xmlBody)
	if err != nil {
		return nil, err
	}
	return index, nil
}
