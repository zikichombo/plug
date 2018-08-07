// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import "fmt"

// DisconnectedError is an error describing a
// channel and direction (in or out) of
// which has no connection.
type DisconnectedError struct {
	IsInput bool
	Chan    int
}

func (d *DisconnectedError) Error() string {
	dir := "input"
	if !d.IsInput {
		dir = "output"
	}
	return fmt.Sprintf("%s channel %d not connected.", dir, d.Chan)
}

func dce(in bool, c int) *DisconnectedError {
	return &DisconnectedError{
		IsInput: in,
		Chan:    c}
}
