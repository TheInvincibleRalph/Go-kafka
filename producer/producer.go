package main

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/gofiber/fiber"
)

type Comment struct {
	Text string `form:"text" json:"text"`
}

func main() {
	app := fiber.New()
	api := app.Group("/api/vi")         //This creates a new route group with the prefix /api/vi. This is useful for versioning the API.
	ap.Post("/comments", createComment) //This line registers a new POST route "/comments" within the /api/vi group. It specifies that the createComment function should handle requests to this route.
	app.Listen(":3000")
}

func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func PushCommentToQueue(topic string, message []byte) error {
	brokersUrl := []string{"localhost:29092"}
	producer, err := ConnectProducer(brokersUrl)
	if err != nil {
		return err
	}
	defer producer.Close()
}

func createComment(c *fiber.Ctx) error {
	cmt := new(Comment)
	//error handling
	if err := c.BodyParser(cmt); err != nil { //BodyParser is a works just like json.NewDecoder in net/http or mux framework
		log.Println(err)
		c.Status(400).JSON(&fiber.Map{ //fiber.Map is a type alias for map[string]interface{} provided by the Fiber framework. It is used to create a generic map with string keys and values of any type. This map can then be easily serialized into JSON using Fiber’s c.JSON method.
			"success": false,
			"message": err,
		})
		return err
	}
	cmtInBytes, err := json.Marshal(cmt)
	PushCommentToQueue("comments", cmtInBytes)

	err = c.JSON(&fiber.Map{
		"success": true,
		"message": "Comment pushed successfully",
		"comment": cmt,
	})

	//error handling
	if err != nil {
		c.Status(500).JSON(&fiber.Map{
			"success": false,
			"message": "Error creating product",
		})
		return err
	}
}
