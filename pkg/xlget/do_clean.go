// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"fmt"
	"os"
	"strings"
)

func (c *Client) doClean(r *Response) error {
	for _, file := range r.downloadFiles {
		//only removes if downloaded
		if strings.HasPrefix(file, r.tempDir) {
			c.logger.Debugf("cleaning file '%s'", file)
			err := os.Remove(file)
			if err != nil {
				return fmt.Errorf("removing downloaded file: %v", err)
			}
		}
	}
	return nil
}
