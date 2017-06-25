package gateway

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/services/udp"
)

const (
	DefaultPrecision   = "ns"
	DefaultTimeout     = 1
	DefaultConsistency = "one"
)

type Sender struct {
	url         *url.URL
	username    string
	password    string
	userAgent   string
	httpClient  *http.Client
	gzip        bool
	precision   string
	consistency string
	Logger      zap.Logger
}

type SenderConfig struct {
	Addr               string       `toml:"addr"`
	Username           string       `toml:"username"`
	Password           string       `toml:"password"`
	UserAgent          string       `toml:"user-agent"`
	Timeout            int          `toml:"timeout"`
	Gzip               bool         `toml:"gzip"`
	InsecureSkipVerify bool         `toml:"insucure-skip-verify"`
	Precision          string       `toml:"precision"`   // ns | s | ms | n
	Consistency        string       `toml:"consistency"` // all | any | one | quorum
	UDPs               []udp.Config `toml:"udp"`
}

func NewSender(c SenderConfig) (*Sender, error) {
	if c.UserAgent == "" {
		c.UserAgent = "InfluxDB-Gateway"
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	if c.Precision == "" {
		c.Precision = DefaultPrecision
	}
	if c.Consistency == "" {
		c.Consistency = DefaultConsistency
	}

	u, err := url.Parse(c.Addr)
	if err != nil {
		return nil, err
	} else if u.Scheme != "http" && u.Scheme != "https" {
		m := fmt.Sprintf("Unsupported protocol scheme: %s, your address"+
			" must start with http:// or https://", u.Scheme)
		return nil, errors.New(m)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.InsecureSkipVerify,
		},
	}

	return &Sender{
		url:         u,
		username:    c.Username,
		password:    c.Password,
		userAgent:   c.UserAgent,
		gzip:        c.Gzip,
		precision:   c.Precision,
		consistency: c.Consistency,
		httpClient: &http.Client{
			Timeout:   time.Duration(c.Timeout) * time.Second,
			Transport: tr,
		},
	}, nil
}

func (s *Sender) WritePoints(database, retentionPolicy string, consistencyLevel models.ConsistencyLevel, points []models.Point) error {
	go func() {
		var b bytes.Buffer
		if s.gzip {
			writer := gzip.NewWriter(&b)
			for _, p := range points {
				if _, err := writer.Write([]byte(p.PrecisionString(s.precision))); err != nil {
					s.Logger.Error(err.Error())
					return
				}
				if _, err := writer.Write([]byte("\n")); err != nil {
					s.Logger.Error(err.Error())
					return
				}
			}
			if err := writer.Flush(); err != nil {
				s.Logger.Error(err.Error())
				return
			}
			if err := writer.Close(); err != nil {
				s.Logger.Error(err.Error())
				return
			}
		} else {
			for _, p := range points {
				if _, err := b.WriteString(p.PrecisionString(s.precision)); err != nil {
					s.Logger.Error(err.Error())
					return
				}
				if err := b.WriteByte('\n'); err != nil {
					s.Logger.Error(err.Error())
					return
				}
			}
		}

		u := s.url
		u.Path = "write"
		req, err := http.NewRequest("POST", u.String(), &b)
		if err != nil {
			s.Logger.Error(err.Error())
			return
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
		params.Set("consistency", s.consistency)

		req.URL.RawQuery = params.Encode()

		resp, err := s.httpClient.Do(req)
		if err != nil {
			s.Logger.Error(err.Error())
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			s.Logger.Error(err.Error())
			return
		}
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			s.Logger.Error(string(body))
			return
		}
	}()

	return nil
}
