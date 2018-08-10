// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import (
	"testing"

	"zikichombo.org/sound"
)

func TestPacketPut(t *testing.T) {
	pkt := packet{}
	v := sound.StereoCd()
	pkt.init(v, 1, 0)
	N := 8
	pkt.samples = make([]float64, N*v.Channels())
	for i := range pkt.samples {
		pkt.samples[i] = float64(i)
	}
	pkt.n = N
	blk := &Block{}
	blk.SampleRate = v.SampleRate()
	blk.Frames = N
	blk.Channels = v.Channels()
	blk.Samples = make([]float64, N*blk.Channels)
	pkt.put(blk)
	for i := N; i < 2*N; i++ {
		if blk.Samples[i-N] != pkt.samples[i] {
			t.Errorf("%d got %f not %f\n", i, blk.Samples[i-N], pkt.samples[i])
		}
	}
	for i := 0; i < N; i++ {
		if blk.Samples[i+N] != pkt.samples[i] {
			t.Errorf("%d got %f not %f\n", i, blk.Samples[i+N], pkt.samples[i])
		}
	}
}

func TestPacketGet(t *testing.T) {
	pkt := packet{}
	v := sound.StereoCd()
	pkt.init(v, 1, 0)
	N := 8
	pkt.samples = make([]float64, N*v.Channels())
	blk := &Block{}
	blk.SampleRate = v.SampleRate()
	blk.Frames = N
	blk.Channels = v.Channels()
	blk.Samples = make([]float64, N*blk.Channels)
	for i := range blk.Samples {
		blk.Samples[i] = float64(i)
	}
	pkt.n = N
	pkt.get(blk)
	for i := N; i < 2*N; i++ {
		if blk.Samples[i-N] != pkt.samples[i] {
			t.Errorf("%d got %f not %f\n", i, blk.Samples[i-N], pkt.samples[i])
		}
	}
	for i := 0; i < N; i++ {
		if blk.Samples[i+N] != pkt.samples[i] {
			t.Errorf("%d got %f not %f\n", i, blk.Samples[i+N], pkt.samples[i])
		}
	}
}
