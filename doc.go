// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// Package plug provides support for audio processing plugs.
//
// Audio plugs are implemented in a 2-tiered fashion.  There is an io tier,
// which manages all the input/output between processing plugs, and a
// processing tier which handles the actual computation in a standardized
// fashion which enables its use in the io tier.
//
// IO Tier
//
// The i/o tier implements audio I/O based on multi-channel, fixed sample rate
// inputs and outputs.  The number of channels and sample rate may vary between
// the input and output, but the sample rate is fixed across all inputs and
// likewise all outputs.
//
// The main interfaces for the I/O tier is the IO interface.  I/O endpoints,
// such as audio capture sources or audio file sources, or playback speakers
// are expected to take the form of snd.Source or snd.Sink, like in the sio
// package,  and may be attached to processing plugs arbitrarily using
// SetInput, AddOutput.
//
// Processing Tier
//
// The processing tier implements computation behind a processing plug.  The
// main interface in the processing tier is Processor.  The job of the
// processor is
//
//  1. To provide a function which maps input samples to output samples.
//
//  2. To provide shape information to the IO tier so the I/O tier can provide
//  the processing function with the appropriate data (and vice versa).
//
// "Shape" information describes how many frames of input and output are to be
// expected in one processing block and also how the function treats channels.
//
// The number of frames of input and output is dynamic and is queried by IO
// before every processing block, by means of the NextFrames() method.
//
// The channel mode describes whether channels are treated independently.  In
// MonoMode, the processing function maps input channels to output channels one
// at a time and there is a supposition that the number of input channels
// equals the number of output channels and the ith input channel is mapped to
// the ith output channel.  This mode enables the lifting of mono-channel
// processors uniformly to multi-channel data.
//
// The FullMode on the other hand maps all the input data of all the channels
// to all the output data of all the output channels in one step, giving more
// context to the processing function but requiring more memory and putting
// more responsability on the processing function in terms of coordinating what
// data is where.
//
package plug /* import "zikichombo.org/plug" */
