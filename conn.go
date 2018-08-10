// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

type conn struct {
	iC    chan *packet
	oC    chan *packet
	doneC chan struct{}
}

func newConn(iC, oC chan *packet, doneC chan struct{}) *conn {
	res := &conn{}
	res.iC = iC
	res.oC = oC
	res.doneC = doneC
	return res
}

func (c *conn) serve() {
	iC := c.iC
	oC := c.oC
	doneC := c.doneC
	var m int
	var err error
	for {
		select {
		case <-doneC:
			return
		case pkt := <-iC:
			if pkt.snk == nil {
				m, err = pkt.src.Receive(pkt.samples)
			} else {
				err = pkt.snk.Send(pkt.samples)
			}
			pkt.n = m
			pkt.err = err
			select {
			case oC <- pkt:
			case <-doneC:
				return
			}
		}
	}
}
