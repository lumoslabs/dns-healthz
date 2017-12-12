package healthz

import (
	"errors"
	"fmt"
	"time"

	"github.com/miekg/dns"
)

type Status struct {
	Tag    string    `json:"tag"`
	RCode  int       `json:"rcode"`
	Answer string    `json:"answer"`
	ErrMsg string    `json:"error"`
	Time   time.Time `json:"time"`
	err    error
}

func newStatus(tag string, code int, ans string, er error) *Status {
	s := &Status{Tag: tag, RCode: code, Answer: ans, Time: time.Now()}
	if er != nil {
		s.ErrMsg = er.Error()
		s.err = er
	} else if code != 0 {
		s.ErrMsg = dns.RcodeToString[code]
		s.err = errors.New(dns.RcodeToString[s.RCode])
	}
	return s
}

func statusNotFound(tag string) *Status {
	e := fmt.Errorf(`Probe '%s' not found`, tag)
	return &Status{
		Tag:    tag,
		Answer: `Not Found`,
		ErrMsg: e.Error(),
		Time:   time.Now(),
		err:    e,
	}
}

func (s *Status) Error() error {
	return s.err
}

func (s *Status) String() string {
	msg := s.Answer
	if s.err != nil {
		msg = fmt.Sprintf(`%s: %v`, s.Answer, s.err)
	}
	return fmt.Sprintf(`[%s] Status at %s: %d - %s`, s.Tag, s.Time, s.RCode, msg)
}
