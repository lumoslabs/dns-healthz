package healthz

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/miekg/dns"
)

type Healthz struct {
	probes []*Probe
	mutex  sync.Mutex
	status []*Status
}

func NewFromConfig(path string) (*Healthz, error) {
	c, er := ReadConfig(path)
	if er != nil {
		return nil, er
	}
	return New(c.Probes), nil
}

func New(p []*Probe) *Healthz {
	st := make([]*Status, 0, len(p))
	for i := range p {
		st = append(st, newStatus(p[i].Name, 0, "", nil))
	}
	return &Healthz{
		probes: p,
		status: st,
	}
}

func (h *Healthz) AddProbe(p ...*Probe) {
	if h.probes == nil {
		h.probes = make([]*Probe, 0)
	}
	h.probes = append(h.probes, p...)

	if h.status == nil {
		h.status = make([]*Status, 0)
	}
	for i := range p {
		h.status = append(h.status, newStatus(p[i].Name, 0, "", nil))
	}
}

func (h *Healthz) Start(ctx context.Context) {
	for _, p := range h.probes {
		glog.V(3).Infof(`Starting probe '%s' {"Server": "%s"}`, p.Name, p.GetAddress())
		go probe(ctx, h, p)
	}
}

func (h *Healthz) Probes() []*Probe {
	return h.probes
}

func (h *Healthz) Status(tag string) *Status {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	for _, s := range h.status {
		if tag == s.Tag {
			return s
		}
	}
	return statusNotFound(tag)
}

func (h *Healthz) AllStatus() []*Status {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.status
}

func (h *Healthz) setStatus(s *Status) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	for i := range h.status {
		if s.Tag == h.status[i].Tag {
			h.status[i] = s
		}
	}
}

func (h *Healthz) String() string {
	p := make([]string, 0)
	for i := range h.probes {
		p = append(p, h.probes[i].String())
	}
	return fmt.Sprintf(`Probes: [%s]`, strings.Join(p, ", "))
}

func probe(ctx context.Context, h *Healthz, p *Probe) {
	ticker := time.NewTicker(p.GetInterval())
	defer func() {
		glog.V(3).Infof(`Shutting down probe '%s' {"Server": "%s"}`, p.Name, p.GetAddress())
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deadline, cancel := context.WithTimeout(ctx, p.GetTimeout())
			query := p.GetQuery()
			if glog.V(4) {
				q := ""
				if query.Question != nil && len(query.Question) > 0 {
					q = fmt.Sprintf(`%s,%s`, query.Question[0].Name, dns.TypeToString[query.Question[0].Qtype])
				}
				glog.Infof(`Running probe '%s' {"Server": "%s", "Query": "%s"}`, p.Name, p.GetAddress(), q)
			}
			in, er := dns.ExchangeContext(deadline, query, p.GetAddress())
			cancel()

			if er != nil {
				h.setStatus(newStatus(p.Name, dns.RcodeServerFailure, "", er))
			} else {
				ans := ""
				if in.Answer != nil && len(in.Answer) > 0 {
					ans = strings.Replace(in.Answer[0].String(), "\t", " ", -1)
				}
				h.setStatus(newStatus(p.Name, in.Rcode, ans, er))
			}
		}
	}
}
