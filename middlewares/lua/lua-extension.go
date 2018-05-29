package lua

import (
	"encoding/json"
	"errors"

	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"
)

func registerExtensions(l *lua.State) {

	luar.Register(l, "valse", luar.Map{
		"json": luar.Map{
			"decode": func(str string) (luar.Map, error) {
				var out luar.Map
				if err := json.Unmarshal([]byte(str), &out); err != nil {
					return nil, err
				}
				return out, nil
			},
			"encode": func(state *lua.State) int {

				var err error
				var bs []byte
				var oo interface{}
				t := state.Type(1)
				switch t {
				case lua.LUA_TNIL, lua.LUA_TNONE:
					oo = nil
				case lua.LUA_TTABLE:
					var o luar.Map
					err = luar.LuaToGo(state, 1, &o)
					oo = o
				case lua.LUA_TNUMBER:
					oo = state.ToNumber(1)
				case lua.LUA_TSTRING:
					oo = state.ToString(1)
				case lua.LUA_TBOOLEAN:
					oo = state.ToBoolean(1)
				default:
					err = errors.New("invalid type")
				}

				out := ""
				if err == nil {
					bs, err = json.Marshal(oo)
					if err == nil {
						out = string(bs)
					}
				}
				state.PushString(out)
				if err != nil {
					state.PushString(err.Error())
				} else {
					state.PushNil()
				}

				return 2
			},
		},
	})

}
