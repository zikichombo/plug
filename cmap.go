// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import "zikichombo.org/sound"

type cmap struct {
	m []int
	i []int
}

// TBD(wsc) make this work for case len(cs) > v.Channels()
func newCmap(v sound.Form, cs ...int) *cmap {
	nC := v.Channels()
	invC := nC
	res := &cmap{}
	res.m = make([]int, nC)
	if len(cs) == 0 {
		res.i = make([]int, invC)
		for i := range res.m {
			res.m[i] = i
			res.i[i] = i
		}
		return res
	}
	invC = len(cs)
	res.i = make([]int, invC)

	for i := range res.i {
		res.i[i] = -1
	}
	for i := range res.m {
		res.m[i] = -1
	}
	for i, c := range cs {
		res.m[c] = i
		res.i[i] = c
	}
	return res
}

func (m *cmap) mapC(c int) int {
	return m.m[c]
}

func (m *cmap) imapC(c int) int {
	return m.i[c]
}
