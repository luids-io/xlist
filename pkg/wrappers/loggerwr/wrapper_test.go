// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package loggerwr_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/wrappers/loggerwr"
)

func TestWrapper_Check(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true, false}}

	output := &logmockup{}
	cfg := loggerwr.DefaultConfig()
	cfg.Prefix = "mockup"
	logged := loggerwr.New(mockup, output, cfg)

	var tests = []struct {
		name    string
		want    bool
		wantLog string
	}{
		{"10.10.10.1", true, "INFO: mockup"},      //0
		{"10.10.10.2", false, "DEBUG: mockup"},    //1
		{"10.10.10.3", true, "INFO: mockup"},      //2
		{"10.10.10.4", false, "DEBUG: mockup"},    //3
		{"www.google.com", false, "WARN: mockup"}, //4
	}
	for idx, test := range tests {
		resp, _ := logged.Check(context.Background(), test.name, xlist.IPv4)
		if test.want != resp.Result {
			t.Errorf("idx[%v] logged.Check(): want=%v got=%v", idx, test.want, resp)
		}

		if test.wantLog != "" && !strings.Contains(output.last, test.wantLog) {
			t.Errorf("idx[%v] logged.Check(): wantLog=%v got=%v", idx, test.wantLog, output.last)
		} else if test.wantLog == "" && output.last != "" {
			t.Errorf("idx[%v] logged.Check(): wantLog=%v got=%v", idx, test.wantLog, output.last)
		}
	}
}

type logmockup struct {
	last string
}

func (m *logmockup) Debugf(template string, args ...interface{}) {
	m.update("DEBUG", template, args...)
}

func (m *logmockup) Infof(template string, args ...interface{}) {
	m.update("INFO", template, args...)
}

func (m *logmockup) Warnf(template string, args ...interface{}) {
	m.update("WARN", template, args...)
}

func (m *logmockup) Errorf(template string, args ...interface{}) {
	m.update("ERROR", template, args...)
}

func (m *logmockup) Fatalf(template string, args ...interface{}) {
	m.update("FATAL", template, args...)
	panic("log fatal")
}

func (m *logmockup) update(level string, template string, args ...interface{}) {
	message := fmt.Sprintf(template, args...)
	m.last = fmt.Sprintf("%s: %s", level, message)
}
