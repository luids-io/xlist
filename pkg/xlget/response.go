// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"time"

	"github.com/luids-io/api/xlist"
)

// State type defines available states for Response
type State int

// List of States
const (
	Ready State = iota
	Downloading
	Uncompressing
	Converting
	Deploying
	Cleaning
	Finished
)

// Response stores information about download and conversion
type Response struct {
	ID         string
	Done       chan struct{}
	Start, End time.Time
	Account    map[xlist.Resource]int
	Updated    bool
	Output     string
	Hash       string

	request *Request
	stop    chan bool
	err     error
	//internal info
	status        State
	tempDir       string
	downloadFiles []string
	sourceFiles   []string
	converted     string
}

// Err returns stored error
func (r *Response) Err() error {
	return r.err
}

// Status returns current state value
func (r *Response) Status() State {
	return r.status
}

// IsComplete returns true if completed
func (r *Response) IsComplete() bool {
	return r.status == Finished
}

// Cancel allows cancel process
func (r *Response) Cancel() bool {
	if r.status != Finished {
		r.stop <- true
		return true
	}
	return false
}

// Wait for finish
func (r *Response) Wait() error {
	if r.status != Finished {
		<-r.Done
	}
	return r.err
}
