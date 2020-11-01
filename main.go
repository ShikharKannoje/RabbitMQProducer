package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/streadway/amqp"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("sad .env file found")
	}
}

func initialize(body []byte) {
	//fmt.Println("Demo RabbitMQ")
	conn, err := amqp.Dial(os.Getenv("RABITTMQ_CRED"))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to RabitMQ")

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(os.Getenv("QUEUE_NAME"), false, false, false, false, nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println(q)

	err = ch.Publish("", os.Getenv("QUEUE_NAME"), false, false, amqp.Publishing{
		ContentType: "json/application",
		Body:        body,
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println("Successfully published the message on the queue.")
}

func main() {

	r := mux.NewRouter()

	//home page
	r.HandleFunc("/", home).Methods("GET")

	//app
	r.HandleFunc("/app", app).Methods("POST")

	log.Fatal(http.ListenAndServe("localhost:8080", r))

}

func home(w http.ResponseWriter, r *http.Request) {

	JSON(w, http.StatusOK, "Sender UP and running")

}

func app(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	fmt.Println(body)
	initialize(body)
}

//JSON formatter
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

//ERROR function
func ERROR(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		JSON(w, statusCode, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		return
	}
	JSON(w, http.StatusBadRequest, nil)
}
