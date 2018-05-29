package lua

import (
	"github.com/kildevaeld/valse2"
	"github.com/stevedonovan/luar"
)

func execute(ctx *valse2.Context, ch chan *VM, id int, middleware bool) bool {
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

func route(id int, ch chan *VM) valse2.RequestHandler {

	return func(ctx *valse2.Context) error {

		execute(ctx, ch, id, false)

		return nil
	}
}

func middleware(id int, ch chan *VM) valse2.MiddlewareHandler {
	return func(next valse2.RequestHandler) valse2.RequestHandler {
		return func(ctx *valse2.Context) error {

			if execute(ctx, ch, id, true) {
				return next(ctx)
			}
			return nil
		}
	}
}
