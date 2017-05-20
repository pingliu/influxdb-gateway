package gateway

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/influxdata/influxdb/models"
)

type Sender struct {
	url        *url.URL
	username   string
	password   string
	userAgent  string
	httpClient *http.Client
	gzip       bool
	precision  string
}

func NewSender(c SenderConfig) (*Sender, error) {
	if c.UserAgent == "" {
		c.UserAgent = "InfluxDB-Gateway"
	}
	u, err := url.Parse(c.Addr)
	if err != nil {
		return nil, err
	} else if u.Scheme != "http" && u.Scheme != "https" {
		m := fmt.Sprintf("Unsupported protocol scheme: %s, your address"+
			" must start with http:// or https://", u.Scheme)
		return nil, errors.New(m)
	}
	return &Sender{
		url:       u,
		username:  c.Username,
		password:  c.Password,
		userAgent: c.UserAgent,
		gzip:      c.Gzip,
		precision: c.Precision,
		httpClient: &http.Client{
			Timeout: time.Duration(c.Timeout) * time.Second,
		},
	}, nil
}

func (s *Sender) WritePoints(database, retentionPolicy string, consistencyLevel models.ConsistencyLevel, points []models.Point) error {
	var b bytes.Buffer
	if s.gzip {
		writer := gzip.NewWriter(&b)
		for _, p := range points {
			if _, err := writer.Write([]byte(p.PrecisionString(s.precision))); err != nil {
				return err
			}
			if _, err := writer.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if err := writer.Flush(); err != nil {
			return err
		}
		if err := writer.Close(); err != nil {
			return err
		}
	} else {
		for _, p := range points {
			if _, err := b.WriteString(p.PrecisionString(s.precision)); err != nil {
				return err
			}
			if err := b.WriteByte('\n'); err != nil {
				return err
			}
		}
	}

	u := s.url
	u.Path = "write"
	req, err := http.NewRequest("POST", u.String(), &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "")
	req.Header.Set("User-Agent", s.userAgent)
	if s.gzip {
		req.Header.Set("Content-Encoding", "gzip")
	}
	if s.username != "" {
		req.SetBasicAuth(s.username, s.password)
	}

	params := req.URL.Query()
	params.Set("db", database)
	params.Set("rp", retentionPolicy)
	params.Set("precision", s.precision)
	params.Set("consistency", strconv.Itoa(int(consistencyLevel)))

	req.URL.RawQuery = params.Encode()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var err = fmt.Errorf(string(body))
		return err
	}

	return nil
}
