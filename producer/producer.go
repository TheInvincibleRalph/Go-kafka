package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/gofiber/fiber"
)

type Comment struct {
	Text string `form:"text" json:"text"`
}

// ======================================== CONNECTS THE PRODUCER TO KAFKA BROKERS ========================================
func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true          // Configures the producer to wait for acknowledgments from Kafka brokers.
	config.Producer.RequiredAcks = sarama.WaitForAll //Ensures that the producer waits for all in-sync replicas to acknowledge the message before considering it successfully sent.
	config.Producer.Retry.Max = 5                    //Sets the maximum number of retries for sending a message.

	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ======================================== CREATES COMMENT ========================================

func createComment(c *fiber.Ctx) error {

	// Instantiate new Message struct
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
	return err
}

// ======================================== SENDS A MESSAGE TO A PARTICULAR KAFKA TOPIC ========================================

func PushCommentToQueue(topic string, message []byte) error {
	brokersUrl := []string{"localhost:29092"}
	producer, err := ConnectProducer(brokersUrl) //this line calls the ConnectProducer function, passing the broker URL slice.
	if err != nil {
		return err
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{ //Creates a new Kafka message with the specified topic and value.
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}
	fmt.Printf("Message is stored in topic(%s)/partition(&d)/offset(%d)\n", topic, partition, offset)
	return nil
}

func main() {
	app := fiber.New()
	api := app.Group("/api/vi")          //This creates a new route group with the prefix /api/vi. This is useful for versioning the API.
	api.Post("/comments", createComment) //This line registers a new POST route "/comments" within the /api/vi group. It specifies that the createComment function should handle requests to this route.
	app.Listen(":3000")
}
