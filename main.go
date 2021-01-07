package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
)

const (
	MqttHost = "MQTT_HOST"
	MqttPort = "MQTT_PORT"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/health", health)
	e.POST("/publish", sendMessage)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))

}

func health(c echo.Context) error {
	return c.JSON(http.StatusOK, BaseResponse{
		Message: http.StatusText(http.StatusOK),
	})
}

func sendMessage(c echo.Context) error {
	message, err := prepareMessage(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, BaseResponse{
			Message: err.Error(),
		})
	}

	err = publish(message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, BaseResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, nil)
}

func prepareMessage(c echo.Context) (msg Message, err error) {
	err = c.Bind(&msg)
	if err != nil {
		return
	}
	if msg.Topic == "" {
		err = fmt.Errorf("topic must not be empty")
		return
	}
	if msg.Payload == nil {
		err = fmt.Errorf("payload must not be empty")
		return
	}
	return
}

func publish(message Message) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", getBrokerHost(), getBrokerPort()))
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	token := client.Publish(message.Topic, message.Qos, message.Retained, message.Payload)
	token.Wait()

	return token.Error()
}

func getBrokerHost() string {
	return os.Getenv(MqttHost)
}

func getBrokerPort() string {
	return os.Getenv(MqttPort)
}

type BaseResponse struct {
	Message string `json:"message"`
}

type Message struct {
	Topic    string      `json:"topic"`
	Qos      byte        `json:"qos"`
	Retained bool        `json:"retained"`
	Payload  interface{} `json:"payload"`
}
