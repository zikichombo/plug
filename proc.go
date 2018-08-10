// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import "fmt"

// ChannelMode indicates how channels are processed.
type ChannelMode int

const (
	// MonoMode implies Process(dst, src *Block) has one channel in src and one in dst.
	MonoMode ChannelMode = iota
	// FullMode implies Process(dst, src *Block) has all input channels in src and dst
	// in channel deinterleaved format.
	FullMode
)

// Proc couples the shape and channel mode of a processing function with the function itself.
//
// The channel mode is static, while the mapping of number of input frames to output frames
// is dynamic, determined by NextFrames(), which is called before every call
// to Process in Full mode and before every sequence of calls to Process which occur
// on a per-channel basis in Mono mode.
type Processor interface {

	// Mode describes the processing mode in which the Process method is invoked.
	ChannelMode() ChannelMode

	// NextFrames returns the desired number of source and destination frames,
	// respectively, for the next processing block.
	NextFrames() (int, int)

	// Process processes samples from src to dst.
	//
	// If ChannelMode() is MonoMode, then the block is processed by calling
	// calling Process for each channel independently, and both src.Channels and
	// dst.Channels == 1. If ChannelMode() is FullMode, then Process is called
	// once with all channels of input and output.
	//
	// Assuming the last call to next frames returned N, M, Process may assume
	// that
	//
	//  1. 1 <= src.Frames <= N
	//  2. dst.Frames == M
	//  3. len(src.Samples) = N * src.Channels
	//  4. len(dst.Samples) = M * dst.Channels
	//  5. src.Samples and dst.Samples are in channel deinterleaved format.
	//
	// In turn, let us denote the value of dst.Frames before the call as M, and after,  M';
	// then process should guarantee that
	//
	// 1. M' contains the real number of outputs written
	// 2. 0 <= M' <= M
	// 3. dst.Samples[:d.Channels*M'] is in channel de-interleaved format.
	Process(dst, src *Block) error
}

// ProcFunc gives the type of a processing function. The semantics of
// ProcFunc are exactly as in Process() in the Processor interface.
type ProcFunc func(dst, src *Block) error

type proc struct {
	mode      ChannelMode
	inFrames  int
	outFrames int
	procFunc  ProcFunc
}

const (
	// default block size of processor in terms of frames
	DefaultInFrames  = 1024
	DefaultOutFrames = 1024
)

// NewProcessor creates a new processor with default frames
// using channel mode "mode".
func NewProcessor(mode ChannelMode, fn ProcFunc) Processor {
	return NewProcessorFrames(mode, fn, DefaultInFrames, DefaultOutFrames)
}

// NewProcessorFrames is like NewProcessor but allows specifying the
// input and output frames.
func NewProcessorFrames(mode ChannelMode, fn ProcFunc, ifrms, ofrms int) Processor {
	return &proc{
		mode:      mode,
		inFrames:  ifrms,
		outFrames: ofrms,
		procFunc:  fn}
}

func (p *proc) Process(dst, src *Block) error {
	return p.procFunc(dst, src)
}

func (p *proc) ChannelMode() ChannelMode {
	return p.mode
}

func (p *proc) NextFrames() (int, int) {
	return p.inFrames, p.outFrames
}

// PassThrough is no-op processor.
var PassThrough = NewProcessor(MonoMode, func(dst, src *Block) error {
	N := src.Frames
	copy(dst.Samples[:N], src.Samples[:N])
	dst.Frames = N
	return nil
})

// ToMono is a mono converter.
var ToMono = NewProcessor(FullMode, func(dst, src *Block) error {
	if dst.Channels != 1 {
		return fmt.Errorf("cannot make mono to %d channel dst", dst.Channels)
	}
	D := float64(src.Channels)
	for f := 0; f < src.Frames; f++ {
		acc := 0.0
		for c := 0; c < src.Channels; c++ {
			i := c*src.Frames + f
			d := src.Samples[i]
			acc += d
		}
		dst.Samples[f] = acc / D
	}
	dst.Frames = src.Frames
	return nil
})
