// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func (c *Client) doDeploy(r *Response) (bool, error) {
	c.logger.Debugf("deploying '%s'", r.request.Output)
	if r.request.NoHash {
		//deploy without hashing
		err := os.Rename(r.converted, r.request.Output)
		if err != nil {
			return false, fmt.Errorf("renaming file: %v", err)
		}
		return true, nil
	}
	//compute hash and replace if mismatches
	c.logger.Debugf("computing filehash '%s'", r.converted)
	hash, err := c.computeHash(r, r.converted)
	if err != nil {
		return false, err
	}
	r.Hash = hash
	c.logger.Debugf("computed hash '%s'", hash)
	md5file := fmt.Sprintf("%s.md5", r.request.Output)
	if fileExists(md5file) {
		storedhash, err := ioutil.ReadFile(md5file)
		if err != nil {
			os.Remove(r.converted)
			return false, fmt.Errorf("comparing hash: %v", err)
		}
		if fileExists(r.request.Output) && hash == string(storedhash) {
			c.logger.Debugf("md5 hashes match")
			os.Remove(r.converted)
			os.Chtimes(md5file, time.Now(), time.Now())
			return false, nil
		}
	}
	err = os.Rename(r.converted, r.request.Output)
	if err != nil {
		return false, fmt.Errorf("renaming file: %v", err)
	}
	ioutil.WriteFile(md5file, []byte(hash), os.ModePerm)
	return true, nil
}

func (c *Client) computeHash(r *Response, file string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var hash string
	var err error
	finished := make(chan bool, 0)
	go func() {
		hash, err = hashFile(ctx, file)
		finished <- true
	}()

	select {
	case <-ctx.Done():
		cancel()
		<-finished
		err = ErrCanceled
	case <-finished:
	}
	close(finished)
	return hash, err
}
