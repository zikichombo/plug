// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package plug

import (
	"sync"

	"zikichombo.org/sound"
)

// Graph is a data type supporting operations for a set
// of interconnected IOs.
//
// The zero value for Graph is an empty graph.
//
// It is not necessary to use the graph api to create and
// interact with I/O plugs directly.  However, Graph facilitates
// some operations when there are many I/O plugs.
type Graph struct {
	nodes []IO
}

// Run runs the graph and returns an error channel
// on which all nodes in the graph report errors.
//
// Usage:
//
//  for e := range g.Run() {
//    // report/handle error
//  }
//
func (g *Graph) Run() <-chan error {
	c := make(chan error)
	var wg sync.WaitGroup

	for _, n := range g.nodes {
		wg.Add(1)
		go func() {
			err := n.Run()
			if err != nil {
				c <- err
			}
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	return c
}

// New creates a new I/O plug.
func (g *Graph) New(iForm, oForm sound.Form, proc Processor) IO {
	n := New(iForm, oForm, proc)
	g.nodes = append(g.nodes, n)
	return n
}

// CheckConnectivity checks whether the graph is fully connected and
// acyclic.
func (g *Graph) CheckConnectivity() error {
	for _, n := range g.nodes {
		// will need to deal with this check with other implementations eventually...
		if err := n.(*node).checkConns(); err != nil {
			return err
		}
	}
	// TBD: cycle check
	return nil
}
