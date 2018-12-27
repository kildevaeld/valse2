package lua

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kildevaeld/valse2/httpcontext"
	"go.uber.org/zap"

	"github.com/aarzilli/golua/lua"
	"github.com/kildevaeld/valse2"
	"github.com/stevedonovan/luar"
)

type VM struct {
	state *lua.State
	id    int
}

func createRequest(ctx *httpcontext.Context) luar.Map {
	return luar.Map{
		"header": luar.Map{
			"get": func(key string) string {
				return ctx.Request().Header.Get(key)
			},
		},
		"url": func() luar.Map {

			return luar.Map{
				"query": luar.Map{
					"get": func(key string) string {
						return ctx.Request().URL.Query().Get(key)
					},
				},
			}
		},
		"path":   string(ctx.Request().URL.Path),
		"method": string(ctx.Request().Method),
		"body": func() string {
			var (
				err  error
				body io.ReadCloser
				bs   []byte
			)
			if body, err = ctx.Request().GetBody(); err != nil {
				return ""
			}
			if bs, err = ioutil.ReadAll(body); err != nil {
				return ""
			}
			return string(bs)
		},
	}
}

func createResponse(ctx *httpcontext.Context) luar.Map {
	body := bytes.NewBuffer(nil)
	ctx.SetBody(ioutil.NopCloser(body))
	return luar.Map{
		"header": luar.Map{
			"get": func(key string) string {
				return ctx.Header().Get(key)
			},
			"set": func(key, val string) {
				ctx.Header().Set(key, val)
			},
		},
		"write": func(str string) {
			body.Write([]byte(str))
		},
		"setStatus": func(status int) {
			ctx.SetStatusCode(status)
		},
	}
}

type LuaOptions struct {
	Path        string
	StopOnError bool
	LuaFactory  func() *lua.State
	WorkQueue   int
}

type File struct {
	Path    string
	Content string
}

type RouterFactory func(method, path string, id int)

func createLua(options LuaOptions, logger *zap.Logger, files []File, factory RouterFactory) (*lua.State, error) {
	var L *lua.State

	if options.LuaFactory != nil {
		L = options.LuaFactory()
	} else {
		L = luar.Init()
		L.OpenLibs()
	}

	registerExtensions(L)

	L.Register("__create_route", func(state *lua.State) int {
		method := state.ToString(1)
		route := state.ToString(2)
		id := state.ToInteger(3)

		factory(method, route, id)
		return 0
	})

	L.Register("__create_middleware", func(state *lua.State) int {
		id := state.ToInteger(1)
		factory("", "", id)
		return 0
	})

	L.DoString(string(MustAsset("prelude.lua")))

	for _, file := range files {
		logger.Debug("reading file", zap.String("path", file.Path))
		err := L.DoString(file.Content)
		if err != nil {
			if options.StopOnError {
				return nil, err
			}
			logger.Error("could not load file", zap.String("path", file.Path), zap.Error(err))
		}

	}

	return L, nil
}

func getSortedFiles(path string) ([]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".lua" || strings.HasPrefix(file.Name(), "_") {
			continue
		}

		fileNames = append(fileNames, filepath.Join(path, file.Name()))
	}

	sort.Strings(fileNames)

	luaPath := os.Getenv("LUA_PATH")
	if luaPath != "" {
		luaPath += ";"
	}

	luaPath += path + "/?.lua"

	os.Setenv("LUA_PATH", luaPath)

	return fileNames, nil
}

type LuaValse struct {
	o  LuaOptions
	ch chan *lua.State
	s  *valse2.Valse
	id int
}

func (l LuaValse) loadFiles() []File {

	files, err := getSortedFiles(l.o.Path)
	if err != nil {
		zap.L().Fatal("load files", zap.Error(err))
	}
	var out []File
	for _, file := range files {

		bs, err := ioutil.ReadFile(file)
		if err != nil {
			zap.L().Fatal("load files", zap.Error(err))
		}
		out = append(out, File{
			Path:    file,
			Content: string(bs),
		})
	}
	return out
}

func (l *LuaValse) Open() error {
	wn := l.o.WorkQueue
	if wn == 0 {
		wn = 5
	}

	logger := zap.L()

	files := l.loadFiles()

	ch := make(chan *VM, wn+1)
	logger.Debug("Registering lua virtualmchaines", zap.Int("count", wn))
	for i := 0; i < wn; i++ {
		lua, err := createLua(l.o, logger, files, func(method, path string, id int) {
			if id <= l.id {
				return
			}
			l.id = id
			if method == "" {
				logger.Debug("middleware  added", zap.Int("id", id))
				l.s.Use(middleware(id, ch))
			} else {
				logger.Debug("path '%d' added: '%s'", zap.Int("id", id), zap.String("path", path))
				l.s.Route(method, path, func(ctx *httpcontext.Context) error {

					fn := route(id, ch)

					return fn(ctx)
					//return nil
				})
			}
		})

		if err != nil {
			return err
		}

		ch <- &VM{lua, i}
	}

	return nil
}

func (l *LuaValse) Close() {
	for c := range l.ch {
		c.Close()
	}
	close(l.ch)

}
func New(server *valse2.Valse, o LuaOptions) *LuaValse {
	return &LuaValse{o, nil, server, 0}
}
