// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import "zikichombo.org/sound/freq"

// Block represents one block of data.
type Block struct {
	Samples    []float64
	Frames     int    // setable by processor
	Channels   int    // read only, static w.r.t. IO lifecycle
	SampleRate freq.T // read only, static w.r.t. IO lifecycle
}
