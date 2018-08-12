package lua

import (
	"fmt"

	"github.com/kildevaeld/valse2/httpcontext"
	"github.com/stevedonovan/luar"
)

func execute(ctx *httpcontext.Context, ch chan *VM, id int, middleware bool) bool {
	req := createRequest(ctx)
	res := createResponse(ctx)

	vm := <-ch
	defer func() {
		ch <- vm
	}()

	state := vm.state
	state.GetGlobal("router")
	state.GetGlobal("Router")
	state.GetField(-1, "trigger")
	state.PushValue(-3)
	luar.GoToLua(state, id)
	luar.GoToLua(state, req)
	luar.GoToLua(state, res)
	state.Call(4, 1)

	return vm.state.ToBoolean(0)
}

func route(id int, ch chan *VM) httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) error {

		execute(ctx, ch, id, false)
		return nil
	}
}

func middleware(id int, ch chan *VM) httpcontext.MiddlewareHandler {
	return func(next httpcontext.HandlerFunc) httpcontext.HandlerFunc {
		fmt.Printf("HERERERERE\n")
		return func(ctx *httpcontext.Context) error {

			if execute(ctx, ch, id, true) {
				return next(ctx)
			}
			return nil
		}
	}
}
