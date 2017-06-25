package gateway

import (
	"go.uber.org/zap"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/services/meta"
	"github.com/influxdata/influxdb/services/udp"
)

type Service interface {
	Open() (err error)
	Close() error
}

type Gateway struct {
	Services     []Service
	PointsWriter interface {
		WritePoints(database, retentionPolicy string, consistencyLevel models.ConsistencyLevel, points []models.Point) error
	}
	MetaClient interface {
		CreateDatabase(name string) (*meta.DatabaseInfo, error)
	}
	Logger zap.Logger
}

type FakeMetaClient struct{}

func (f *FakeMetaClient) CreateDatabase(name string) (*meta.DatabaseInfo, error) {
	return nil, nil
}

func New(c Config, log zap.Logger) (*Gateway, error) {
	pointsWriter, err := NewSender(c.Sender)
	pointsWriter.Logger = log
	if err != nil {
		return nil, err
	}
	gateway := &Gateway{
		MetaClient:   &FakeMetaClient{},
		PointsWriter: pointsWriter,
		Logger:       log,
	}
	for _, conf := range c.Sender.UDPs {
		gateway.AppendUDPService(conf)
	}

	return gateway, nil
}

func (g *Gateway) AppendUDPService(conf udp.Config) {
	if !conf.Enabled {
		return
	}
	srv := udp.NewService(conf)
	srv.PointsWriter = g.PointsWriter
	srv.MetaClient = g.MetaClient
	srv.Logger = g.Logger
	g.Services = append(g.Services, srv)
}

func (g *Gateway) Open() (err error) {
	for _, s := range g.Services {
		err = s.Open()
		if err != nil {
			return
		}
	}
	return nil
}

func (g *Gateway) Close() error {
	for _, s := range g.Services {
		err := s.Close()
		if err != nil {
			g.Logger.Error(err.Error())
			continue
		}
	}
	return nil
}
