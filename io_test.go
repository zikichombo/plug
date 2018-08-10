// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import (
	"io"
	"testing"

	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/gen"
	"zikichombo.org/sound/ops"
)

func TestIOMonoChain(t *testing.T) {
	valve := sound.NewForm(44100*freq.Hertz, 1)
	u0 := New(valve, valve, PassThrough)
	u1 := New(valve, valve, PassThrough)
	u2 := New(valve, valve, PassThrough)
	u2.SetInput(u1.Output())
	u1.SetInput(u0.Output())
	u0.SetInput(ops.Limit(gen.Noise(), 44100))
	out := u2.Output()
	go u0.Run()
	go u1.Run()
	go u2.Run()
	buf := make([]float64, 1024)
	ttl := 0
	for {
		n, err := out.Receive(buf)
		ttl += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			break
		}
	}
	if ttl != 44100 {
		t.Errorf("got %d not 44100", ttl)
	}
}

func TestIOMonoMultiChanChain(t *testing.T) {
	valve := sound.NewForm(44100*freq.Hertz, 3)
	u0 := New(valve, valve, PassThrough)
	u1 := New(valve, valve, PassThrough)
	u2 := New(valve, sound.MonoCd(), ToMono)
	u2.SetInput(u1.Output())
	u1.SetInput(u0.Output())
	u0.SetInput(ops.Limit(gen.Noise(), 44100), 0)
	u0.SetInput(ops.Limit(gen.Noise(), 44100), 1)
	u0.SetInput(ops.Limit(gen.Noise(), 44100), 2)
	out := u2.Output()
	go u0.Run()
	go u1.Run()
	go u2.Run()
	buf := make([]float64, 1024)
	ttl := 0
	for {
		n, err := out.Receive(buf)
		ttl += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			break
		}
	}
	if ttl != 44100 {
		t.Errorf("got %d not 44100", ttl)
	}
}

func TestIOMultiOut(t *testing.T) {
	valve := sound.NewForm(44100*freq.Hertz, 1)
	u0 := New(valve, valve, PassThrough)
	u0.SetInput(ops.Limit(gen.Noise(), 44100))
	out := u0.Output(0, 0) // duplicate channel 0 in channels 0,1 of output
	go u0.Run()
	buf := make([]float64, 1024)
	ttl := 0
	for {
		n, err := out.Receive(buf)
		ttl += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			break
		}
	}
	if ttl != 44100 {
		t.Errorf("got %d not 44100", ttl)
	}
}

func TestIOMultiAddOut(t *testing.T) {
	valve := sound.StereoCd()
	u0 := New(valve, valve, PassThrough)
	if err := u0.SetInput(ops.Limit(gen.Noise(), 44100), 0); err != nil {
		t.Fatal(err)
	}
	if err := u0.SetInput(ops.Limit(gen.Noise(), 44100), 1); err != nil {
		t.Fatal(err)
	}
	src0, snk0 := sound.Pipe(valve)
	src1, snk1 := sound.Pipe(valve)
	if err := u0.AddOutput(snk0); err != nil {
		t.Fatal(err)
	}
	if err := u0.AddOutput(snk1); err != nil {
		t.Fatal(err)
	}
	u1 := New(sound.NewForm(44100*freq.Hertz, 4), sound.MonoCd(), ToMono)
	if err := u1.SetInput(src0, 0, 2); err != nil {
		t.Fatal(err)
	}
	if err := u1.SetInput(src1, 1, 3); err != nil {
		t.Fatal(err)
	}
	out := u1.Output()
	go u0.Run()
	go u1.Run()
	buf := make([]float64, 1024)
	ttl := 0
	for {
		n, err := out.Receive(buf)
		ttl += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			break
		}
	}
	if ttl != 44100 {
		t.Errorf("got %d not 44100", ttl)
	}
}

func TestIODiamond(t *testing.T) {
	valve := sound.NewForm(44100*freq.Hertz, 2)
	u0 := New(valve, valve, PassThrough)
	u0.SetInput(ops.Limit(gen.Noise(), 44100), 0)
	u0.SetInput(ops.Limit(gen.Noise(), 44100), 1)
	mono := sound.NewForm(44100*freq.Hertz, 1)
	ua := New(mono, mono, PassThrough)
	ua.SetInput(u0.Output(0))
	ub := New(mono, mono, PassThrough)
	ub.SetInput(u0.Output(1))
	u1 := New(valve, valve, PassThrough)
	u1.SetInput(ua.Output(), 0)
	u1.SetInput(ub.Output(), 1)
	out := u1.Output()
	go u0.Run()
	go ua.Run()
	go ub.Run()
	go u1.Run()
	buf := make([]float64, 1024)
	ttl := 0
	for {
		n, err := out.Receive(buf)
		ttl += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			break
		}
	}
	if ttl != 44100 {
		t.Errorf("got %d not 44100", ttl)
	}
}
