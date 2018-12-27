package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kildevaeld/dict"
	"github.com/kildevaeld/strong"

	"go.uber.org/zap"

	"github.com/kildevaeld/valse2"
	"github.com/kildevaeld/valse2/httpcontext"
	"github.com/kildevaeld/valse2/middlewares/cache"
	mpanic "github.com/kildevaeld/valse2/middlewares/panic"
	"github.com/kildevaeld/valse2/mountables/filesystem"

	system "github.com/kildevaeld/go-system"
)

func main() {
	if err := system.Run(wrappedMain); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

func wrappedMain(kill system.KillChannel) error {

	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(log)

	server := valse2.NewWithOptions(&valse2.Options{
		Debug: true,
	})

	//server.Use(logger.Logger())

	go func() {
		<-kill
		server.Close()
	}()

	server.Use(mpanic.New())
	//server.Use(logger.Logger())

	server.Get("/", func(ctx *httpcontext.Context, next httpcontext.HandlerFunc) error {

		return next(ctx)
	}, cache.NewCacheControl(nil), func(ctx *httpcontext.Context) error {
		return ctx.HTML("<h1>Hello, World</h1>")
	})

	server.Get("/world/:greeting", func(ctx *httpcontext.Context) error {
		return ctx.HTML(fmt.Sprintf("<h1>Hello %s</h1>", ctx.Params().ByName("greeting")))
	}).Get("/error", func(ctx *httpcontext.Context) error {
		return strong.NewHTTPError(strong.StatusUnauthorized, "test")
	}).Get("/panic", func(ctx *httpcontext.Context) error {
		panic("oh ooohhh")
		return nil
	})

	group := valse2.NewGroup()
	group.Use(func(ctx *httpcontext.Context, next httpcontext.HandlerFunc) error {
		fmt.Println("hello from middleware")
		defer println("hello from middleware again")
		return next(ctx)
	}).Get("/", func(ctx *httpcontext.Context) error {
		return ctx.Text(fmt.Sprintf("Ob la di ob la da %s", ctx.Request().URL))
	})

	server.Mount("/test/group", group)

	server.Mount("/api/blog", valse2.NewRest("blog").Create(func(ctx *httpcontext.Context) error {

		return ctx.JSON(dict.Map{})
	}).Get(func(ctx *httpcontext.Context, id string) error {
		return ctx.JSON(dict.Map{
			"hello": "world " + id,
		})
	}))

	server.Mount("/files", filesystem.New(http.Dir("./")), cache.NewCacheControl(&cache.CacheControl{
		Debug:   false,
		Private: false,
		MaxAge:  20000000,
	}))

	return server.Listen(":4000")

}
