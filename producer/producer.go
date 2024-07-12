package main

type Comment struct {
	Text string `form:"text" json:"text"`
}

func main() {
	app := fiber.New()
	api := app.Group("/api/vi")
	ap.Post("/comments", createComment)
	app.Listen(":3000")
}
