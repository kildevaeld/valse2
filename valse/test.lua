function fn(req, res)
    res.write("Hello, World ")
    return true
end

router:get(
    "/test",
    function(req, res)
        local str, err =
            valse.json.encode(
            {
                rapper = "er lige med"
            }
        )
        res.header.set("Content-Type", "application/json")
        res.write(str)
    end
)

router:get(
    "/",
    function(req, res)
        res.header.set("Content-Type", "text/html")

        res.write(
            render_html(
                function()
                    return div {
                        h1 "Hello, World",
                        p "It's the end of the world as we know it"
                    }
                end
            )
        )
    end
)

router:post(
    "/rap",
    function(req, res)
        res.write(res.body())
    end
)

--router:use(fn)
