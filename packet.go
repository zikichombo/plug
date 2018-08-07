// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import (
	"zikichombo.org/sound"
)

type packet struct {
	err     error
	n       int
	samples []float64
	cmap    *cmap
	nC      int
	src     sound.Source
	snk     sound.Sink
}

func (p *packet) init(v sound.Form, cs ...int) {
	p.cmap = newCmap(v, cs...)
	p.err = nil
	p.n = 0
	p.samples = p.samples[:0]
	if len(cs) == 0 {
		p.nC = v.Channels()
	} else {
		p.nC = len(cs)
	}
}

func (p *packet) put(dst *Block) int {
	sl := p.samples
	nC := dst.Channels
	frms := p.n
	cmap := p.cmap
	for c := 0; c < nC; c++ {
		cc := cmap.mapC(c)
		if cc == -1 {
			continue
		}
		sStart := cc * frms
		sEnd := sStart + frms
		dStart := c * frms
		dEnd := dStart + frms
		copy(dst.Samples[dStart:dEnd], sl[sStart:sEnd])
	}
	return frms
}

func (p *packet) get(src *Block) {
	cmap := p.cmap
	nC := len(cmap.i)
	frms := src.Frames
	sl := p.samples
	sl = buffer(sl, nC, frms)
	for cc := 0; cc < nC; cc++ /*not really*/ {
		c := cmap.imapC(cc)
		sStart := c * frms
		sEnd := sStart + frms
		dStart := cc * frms
		dEnd := dStart + frms
		copy(sl[dStart:dEnd], src.Samples[sStart:sEnd])
	}
	p.samples = sl
	p.n = frms
}

func buffer(d []float64, c, f int) []float64 {
	N := c * f
	if cap(d) < N {
		tmp := make([]float64, (5*N)/3)
		copy(tmp, d)
		d = tmp
	}
	return d[:N]
}
