package main

import (
	"fmt"
	"os"

	"github.com/kildevaeld/strong"

	"go.uber.org/zap"

	"github.com/kildevaeld/valse2"

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

	server.Get("/", func(ctx *valse2.Context, next valse2.RequestHandler) error {

		return next(ctx)
	}, func(ctx *valse2.Context) error {
		return ctx.HTML("<h1>Hello, World</h1>")
	})

	server.Get("/world/:greeting", func(ctx *valse2.Context) error {
		return ctx.HTML(fmt.Sprintf("<h1>Hello %s</h1>", ctx.Params().ByName("greeting")))
	}).Get("/error", func(ctx *valse2.Context) error {
		return strong.NewHTTPError(strong.StatusUnauthorized, "test")
	})

	return server.Listen(":3000")

}
