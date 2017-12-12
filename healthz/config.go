package healthz

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/miekg/dns"
)

var (
	defaultQuery         = []string{`www.google.com`, "A"}
	defaultCheckInterval = 10 * time.Second
	defaultDialTimeout   = 5 * time.Second
)

type Config struct {
	Probes []*Probe `json:"probes"`
}

type Probe struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	Query         string `json:"query,omitempty"`
	Timeout       string `json:"timeout,omitempty"`
	CheckInterval string `json:"interval,omitempty"`
}

func ReadConfig(path string) (*Config, error) {
	var cfg Config
	data, er := ioutil.ReadFile(path)
	if er != nil {
		return nil, er
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (p *Probe) GetAddress() string {
	bits := strings.SplitN(p.Address, ":", 2)
	if len(bits) == 2 {
		return p.Address
	}
	return fmt.Sprintf(`%s:53`, bits[0])
}

func (p *Probe) GetQuery() *dns.Msg {
	var (
		query = new(dns.Msg)
		q     = make([]string, len(defaultQuery))
	)

	copy(q, defaultQuery)
	if p.Query != "" {
		bits := strings.SplitN(p.Query, ",", 2)
		q[0] = bits[0]
		if len(bits) == 2 {
			q[1] = bits[1]
		}
	}

	return query.SetQuestion(dns.Fqdn(q[0]), dns.StringToType[q[1]])
}

func (p *Probe) GetTimeout() time.Duration {
	if t, er := time.ParseDuration(p.Timeout); er == nil {
		return t
	}
	return defaultDialTimeout
}

func (p *Probe) GetInterval() time.Duration {
	if t, er := time.ParseDuration(p.CheckInterval); er == nil {
		return t
	}
	return defaultCheckInterval
}

func (p *Probe) String() string {
	o, er := json.Marshal(p)
	if er != nil {
		return ""
	}
	return string(o)
}
