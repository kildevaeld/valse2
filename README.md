
```golang

import (
    "github.com/kildevaeld/valse2"
    "github.com/kildevaeld/valse2/httpcontext"
)

server := valse2.New()

server.Get("/", func (ctx *httpcontext.Context) error {
    return ctx.JSON(map[string]interface{}{
        "Hello": "world"
    })
}).Listen(":3000")


```