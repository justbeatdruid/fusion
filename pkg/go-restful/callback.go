package restful

import (
	"sync"
)

// Copyright 2020 Chinamobile Xiongyan. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

// CallbackChain is an object for processing a seires of callback functions which will be called after Router.Function
type CallbackChain struct {
	Callbacks []CallbackFunction
}

// ProcessCallback calls callback functions in order with handled request and response.
// Entities written can be found in response.GetEntities().
func (c *CallbackChain) ProcessCallback(request *Request, response *Response) {
	if len(c.Callbacks) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	for i := range c.Callbacks {
		wg.Add(1)
		go func(idx int) {
			c.Callbacks[idx](request, response)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

// CallbackFunction defines actions with request and response
type CallbackFunction func(*Request, *Response)
