// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import (
	"fmt"
	"io"
	"sync"

	"zikichombo.org/sound"
)

// IO provides a generic minimal interface for an audio/sound processor.
// Implementations must be safe for use in multiple goroutines, but
// may assume that the Run() method is called at most once.
type IO interface {

	// InForm returns the sample rate and number of channels of the
	// input.
	InForm() sound.Form

	// OutForm returns the sample rate and number of channels of the
	// output of the node.
	OutForm() sound.Form

	// SetInput sets the input to s.  If cs is empty, then s must map channels
	// one-to-one with InForm() and have the same sample rate.  If cs is not empty,
	// then when cs[i] == j, then the j'th channel in IO.InForm() corresponds to
	// the i'th channel of s.
	//
	// SetInput panics if any c in cs is out of bounds w.r.t. InForm().Channels().
	//
	// SetInput returns a non-nil error if the channel and sample rates of
	// IO.InForm() and s are not compatible, or if there is a channel in IO
	// which would map to more than one sound.Source as a result of the call to
	// SetInput.
	//
	SetInput(s sound.Source, cs ...int) error

	// AddOutput, if successful, causes the object implementing IO to direct a copy of
	// its output to the destination d.
	//
	// If cs is empty, the d must have the same sample rate and number of
	// channels as IO.  If cs is not empty, it specifies IO channels which
	// are mapped to d: cs[i] = j <=> the j'th channel in IO is mapped to
	// the i'th channel in d.
	//
	// AddOutput panics if any c in cs is out of bounds w.r.t. OutForm().Channels()
	//
	// AddOutput returns a non-nil error if the channel and sample rates of
	// IO.OutForm() and d are not compatible.
	//
	AddOutput(d sound.Sink, cs ...int) error

	// Output returns the output of the node as a sound.Source.
	// If cs is empty, then the resulting source has valve equal to
	// IO.OutForm().  Otherwise, cs lists a sequence of channels in IO
	// to be used in the result.  If cs[i] = j then the i'th channel of
	// the resulting source is the j'th output channel of IO.
	//
	// If any c in cs is out of bounds w.r.t. OutForm().Channels(), then Output panics.
	//
	// Every non-panicking call to Output generates a distinct new sound.Source which
	// can be used independently in different goroutines.
	Output(cs ...int) sound.Source

	// Run runs the IO plug.  Run blocks until it returns.  It will return a non-nil
	// error if something other than io.EOF ended its inputs.  Upon return, all
	// Sources going into the node and Sinks going out have been Close()d.
	Run() error
}

type node struct {
	mu             sync.Mutex
	iForm, oForm   sound.Form
	iBlock, oBlock *Block
	icCounts       []int
	ocCounts       []int

	ins   []*conn
	outs  []*conn
	iPkts []packet
	oPkts []packet
	inC   chan *packet
	prC   chan *packet
	oC    chan *packet
	odC   chan *packet
	doneC chan struct{}
	proc  Processor
}

// New creates a new plug mapping input of channels and sampling frequency
// iForm to output oForm, using the Processor proc
func New(iForm, oForm sound.Form, proc Processor) IO {
	res := &node{
		icCounts: make([]int, iForm.Channels()),
		ocCounts: make([]int, oForm.Channels()),
		ins:      make([]*conn, 0, 2),
		outs:     make([]*conn, 0, 2),
		iPkts:    make([]packet, 0, 2),
		oPkts:    make([]packet, 0, 2),
		oC:       make(chan *packet),
		odC:      make(chan *packet),
		inC:      make(chan *packet),
		prC:      make(chan *packet),
		doneC:    make(chan struct{}),
		iForm:    iForm,
		oForm:    oForm,
		iBlock:   &Block{SampleRate: iForm.SampleRate(), Channels: iForm.Channels()},
		oBlock:   &Block{SampleRate: oForm.SampleRate(), Channels: oForm.Channels()},
		proc:     proc}
	return res
}

// InForm implements T giving the input frequency and number of channels.
func (n *node) InForm() sound.Form {
	return n.iForm
}

// OutForm implements T giving the output frequency and number of channels.
func (n *node) OutForm() sound.Form {
	return n.oForm
}

// Output implement T.
func (n *node) Output(cs ...int) sound.Source {
	n.mu.Lock()
	defer n.mu.Unlock()
	ov := n.oForm
	if len(cs) != 0 {
		ov = sound.NewForm(ov.SampleRate(), len(cs))
		for i := range n.ocCounts {
			n.ocCounts[i]++
		}
	} else {
		for i := range n.ocCounts {
			n.ocCounts[i]++
		}
	}
	conn := newConn(n.oC, n.odC, n.doneC)
	m := len(n.outs)
	n.outs = append(n.outs, conn)
	n.oPkts = append(n.oPkts, packet{})
	pkt := &n.oPkts[m]
	pkt.init(n.oForm, cs...)
	pkt.src, pkt.snk = sound.Pipe(ov)
	return pkt.src
}

// AddOutput implements IO.
func (n *node) AddOutput(d sound.Sink, cs ...int) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if d.SampleRate() != n.oForm.SampleRate() {
		return fmt.Errorf("frequency mismatch: got %s not %s\n", d.SampleRate(), n.iForm.SampleRate())
	}
	if len(cs) == 0 && d.Channels() != n.oForm.Channels() {
		return fmt.Errorf("channel mismatch: got %d not %d\n", d.Channels(), n.oForm.Channels())
	}
	if len(cs) != 0 && d.Channels() != len(cs) {
		return fmt.Errorf("channel mismatch: got %d not %d\n", d.Channels(), len(cs))
	}
	if len(cs) == 0 {
		for i := range n.ocCounts {
			n.ocCounts[i]++
		}
	} else {
		for _, c := range cs {
			n.ocCounts[c]++
		}
	}
	conn := newConn(n.oC, n.odC, n.doneC)
	m := len(n.outs)
	n.outs = append(n.outs, conn)
	n.oPkts = append(n.oPkts, packet{})
	pkt := &n.oPkts[m]
	pkt.init(n.oForm, cs...)
	pkt.snk = d
	pkt.src = nil
	return nil
}

// SetInput implements IO.
func (n *node) SetInput(src sound.Source, cs ...int) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if src.SampleRate() != n.iForm.SampleRate() {
		return fmt.Errorf("frequency mismatch: got %s not %s\n", src.SampleRate(), n.iForm.SampleRate())
	}
	if err := n.ckInputsUnique(cs...); err != nil {
		return err
	}
	conn := newConn(n.inC, n.prC, n.doneC)
	m := len(n.ins)
	n.ins = append(n.ins, conn)
	n.iPkts = append(n.iPkts, packet{})
	pkt := &n.iPkts[m]
	pkt.init(n.iForm, cs...)
	pkt.src = src
	return nil
}

// Run implements T running the plug.
func (n *node) Run() error {
	defer func() {
		close(n.doneC)
		for i := range n.oPkts {
			n.oPkts[i].snk.Close()
		}
		for i := range n.iPkts {
			n.iPkts[i].src.Close()
		}
	}()
	if err := n.checkConns(); err != nil {
		return err
	}
	n.serve()
	var err error
	for {
		err = n.process()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (n *node) process() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	proc := n.proc
	iC := n.iForm.Channels()
	oC := n.oForm.Channels()
	iFrms, oFrms := proc.NextFrames()
	iBlock, oBlock := n.iBlock, n.oBlock

	// ensure buffers are allocated as per request from proc.
	iBlock.Samples = buffer(n.iBlock.Samples, iC, iFrms)
	iBlock.Frames = iFrms
	oBlock.Samples = buffer(n.oBlock.Samples, oC, oFrms)
	oBlock.Frames = oFrms

	// trigger receives on all inputs
	for i := range n.ins {
		pkt := &n.iPkts[i]
		pkt.err = nil
		pkt.n = iFrms
		pkt.samples = buffer(pkt.samples, pkt.nC, pkt.n)
		n.inC <- pkt
	}

	// read all input into iBlock
	nFrms := -1
	for i := range n.ins {
		_ = i
		pkt := <-n.prC
		if pkt.err != nil {
			return pkt.err
		}
		m := pkt.put(iBlock)
		if nFrms == -1 {
			nFrms = m
		}
		if m != nFrms {
			panic("wilma!")
		}
	}
	iBlock.Frames = nFrms

	// actually finally process
	switch proc.ChannelMode() {
	case MonoMode:
		// save channels and samples members and restore them later
		// each channel will i/oBlock with appropriately modified members
		// for call to MonoMode Process().
		ic := iBlock.Channels
		isl := iBlock.Samples
		oc := oBlock.Channels
		osl := oBlock.Samples
		for i := 0; i < iC; i++ {
			iStart := i * nFrms
			iEnd := iStart + nFrms
			iBlock.Samples = isl[iStart:iEnd]
			oStart := i * nFrms
			oEnd := oStart + nFrms
			oBlock.Samples = osl[oStart:oEnd]
			if err := proc.Process(oBlock, iBlock); err != nil {
				return err
			}
		}
		iBlock.Channels = ic
		iBlock.Samples = isl
		oBlock.Channels = oc
		oBlock.Samples = osl

	case FullMode:
		err := proc.Process(oBlock, iBlock)
		if err != nil {
			return err
		}
	default:
		panic("wilma!")
	}
	// send out the outputs
	for i := range n.oPkts {
		pkt := &n.oPkts[i]
		pkt.get(oBlock)
		n.oC <- pkt
	}
	// and make sure they are done, reporting any errors.
	for i := range n.oPkts {
		_ = i
		pkt := <-n.odC
		if pkt.err != nil {
			return pkt.err
		}
	}
	return nil
}

func (n *node) serve() {
	for _, iConn := range n.ins {
		go iConn.serve()
	}
	for _, oConn := range n.outs {
		go oConn.serve()
	}
}

func (n *node) ckInputsUnique(cs ...int) error {
	// check all input channels have at most one source
	if len(cs) == 0 {
		for c, ct := range n.icCounts {
			if ct > 0 {
				return fmt.Errorf("channel %d already has an input", c)
			}
		}
		for c := range n.icCounts {
			n.icCounts[c] = 1
		}
	} else {
		for _, c := range cs {
			if n.icCounts[c] > 0 {
				return fmt.Errorf("channel %d already has an input", c)
			}
		}
		for _, c := range cs {
			n.icCounts[c] = 1
		}
	}
	return nil
}

func (n *node) ckOutputsUnique(cs ...int) error {
	// check all input channels have at most one source
	if len(cs) == 0 {
		for c, ct := range n.ocCounts {
			if ct > 0 {
				return fmt.Errorf("channel %d already has an output", c)
			}
		}
		for c := range n.ocCounts {
			n.ocCounts[c] = 1
		}
	} else {
		for _, c := range cs {
			if n.ocCounts[c] > 0 {
				return fmt.Errorf("channel %d already has an output", c)
			}
		}
		for _, c := range cs {
			n.ocCounts[c] = 1
		}
	}
	return nil
}

func (n *node) checkConns() error {
	// check local connectivity
	for i, ct := range n.icCounts {
		if ct == 0 {
			return dce(true, i)
		}
	}
	for i, ct := range n.ocCounts {
		if ct == 0 {
			return dce(false, i)
		}
	}
	return nil
}
