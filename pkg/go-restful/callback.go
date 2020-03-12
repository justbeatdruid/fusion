package restful

// Copyright 2020 Chinamobile Xiongyan. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

// CallbackChain is an object for processing a seires of callback functions which will be called after Router.Function
type CallbackChain struct {
	Callbacks []CallbackFunction
	Index     int
}

// ProcessCallback calls callback functions in order with handled request and response.
// Entities written can be found in response.GetEntities().
func (c *CallbackChain) ProcessCallback(request *Request, response *Response) {
	if c.Index < len(c.Callbacks) {
		c.Index++
		c.Callbacks[c.Index-1](request, response, c)
	}
}

// CallbackFunction defines actions with request and response
type CallbackFunction func(*Request, *Response, *CallbackChain)
