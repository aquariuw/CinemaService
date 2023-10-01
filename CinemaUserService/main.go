package main

import (
	. "CinemaUserService/handlers"
	. "CinemaUserService/helpers"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("Service is starting...", time.Second)
	ConnectToMongoDB()
	ConnectToRedis()
	go func() {
		http.HandleFunc("/", HandleCORS(EmptyHandler))
		http.HandleFunc("/user/save", HandleCORS(SaveHandler))
		http.HandleFunc("/user/showbyid", HandleCORS(GetbyIDHandler))
		http.HandleFunc("/user/showtickets", HandleCORS(GetAllTicketByName))
		http.HandleFunc("/user/showall", HandleCORS(GetHandler))
		http.HandleFunc("/user/showbyname", HandleCORS(GetbyNameHandler))
		http.HandleFunc("/user/deletebyid", HandleCORS(DeletebyIDHandler))
		http.HandleFunc("/user/deletebyname", HandleCORS(DeletebyNameHandler))
		http.HandleFunc("/user/updbyid", HandleCORS(UpdatebyIDHandler))
		http.HandleFunc("/user/updbyname", HandleCORS(UpdatebyNameHandler))
		http.HandleFunc("/get/token", HandleCORS(GetToken))
		log.Println("Server is starting on :8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	go StartMessageReceiver()

	select {}
}
