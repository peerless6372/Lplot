// Copyright 2017 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import "net/http"

// CreateTestContext returns a fresh engine and context for testing purposes
func CreateTestContext(w http.ResponseWriter) (c *Context, r *Engine) {
	r = New()
	c = r.allocateContext()
	c.reset()
	c.writermem.reset(w)
	return
}

func CreateNewContext(r *Engine) (c *Context) {
	c = r.pool.Get().(*Context)
	c.reset()
	c.writermem.reset(nil)
	return
}

func RecycleContext(r *Engine, c *Context) {
	r.pool.Put(c)
}
