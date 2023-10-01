package main

import (
	. "CinemaTicketService/handlers"
	"CinemaTicketService/helpers"
	"log"
	"net/http"
)

func main() {
	helpers.ConnectToMongoDB()
	helpers.ConnectToRabbitMQ()
	helpers.ConnectToRedis()
	http.HandleFunc("/ticket/buy", HandleCORS(BuyTicketByName))
	http.HandleFunc("/ticket/free", HandleCORS(BuyFreeTicketByName))
	http.HandleFunc("/create/film", HandleCORS(CreateEvent))
	log.Fatal(http.ListenAndServe("0.0.0.0:8082", nil))
}
