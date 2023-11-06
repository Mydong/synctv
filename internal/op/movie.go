package op

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/synctv-org/synctv/internal/model"
	"github.com/synctv-org/synctv/utils"
	"github.com/zijiren233/livelib/av"
	"github.com/zijiren233/livelib/container/flv"
	"github.com/zijiren233/livelib/protocol/hls"
	rtmpProto "github.com/zijiren233/livelib/protocol/rtmp"
	"github.com/zijiren233/livelib/protocol/rtmp/core"
	rtmps "github.com/zijiren233/livelib/server"
)

type Movie struct {
	*model.Movie
	lock    sync.RWMutex
	channel *rtmps.Channel
}

func (m *Movie) Channel() (*rtmps.Channel, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.channel, m.init()
}

func genTsName() string {
	return utils.SortUUID()
}

func (m *Movie) init() (err error) {
	if err = m.Movie.Validate(); err != nil {
		return
	}
	switch {
	case m.Base.Live && m.Base.RtmpSource:
		if m.channel == nil {
			m.channel = rtmps.NewChannel()
			m.channel.InitHlsPlayer(hls.WithGenTsNameFunc(genTsName))
		}
	case m.Base.Live && m.Base.Proxy:
		u, err := url.Parse(m.Base.Url)
		if err != nil {
			return err
		}
		switch u.Scheme {
		case "rtmp":
			if m.channel == nil {
				m.channel = rtmps.NewChannel()
				m.channel.InitHlsPlayer(hls.WithGenTsNameFunc(genTsName))
				go func() {
					for {
						if m.channel.Closed() {
							return
						}
						cli := core.NewConnClient()
						if err = cli.Start(m.Base.Url, av.PLAY); err != nil {
							cli.Close()
							time.Sleep(time.Second)
							continue
						}
						if err := m.channel.PushStart(rtmpProto.NewReader(cli)); err != nil {
							cli.Close()
							time.Sleep(time.Second)
						}
					}
				}()
			}
		case "http", "https":
			if m.channel == nil {
				m.channel = rtmps.NewChannel()
				m.channel.InitHlsPlayer(hls.WithGenTsNameFunc(genTsName))
				go func() {
					for {
						if m.channel.Closed() {
							return
						}
						r := resty.New().R()
						for k, v := range m.Base.Headers {
							r.SetHeader(k, v)
						}
						// r.SetHeader("User-Agent", UserAgent)
						resp, err := r.Get(m.Base.Url)
						if err != nil {
							time.Sleep(time.Second)
							continue
						}
						if err := m.channel.PushStart(flv.NewReader(resp.RawBody())); err != nil {
							time.Sleep(time.Second)
						}
						resp.RawBody().Close()
					}
				}()
			}
		default:
			return errors.New("unsupported scheme")
		}
	}
	return nil
}

func (m *Movie) Terminate() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.terminate()
}

func (m *Movie) terminate() {
	if m.channel != nil {
		m.channel.Close()
		m.channel = nil
	}
}

func (m *Movie) Update(movie model.BaseMovie) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.terminate()
	m.Movie.Base = movie
	return m.init()
}
