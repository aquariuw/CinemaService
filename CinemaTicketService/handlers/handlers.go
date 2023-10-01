package handlers

import (
	. "CinemaTicketService/helpers"
	. "CinemaTicketService/models"
	"context"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	signingKey = []byte("AbDuRRaHmAn")
)

func BuyFreeTicketByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodOptions {
		log.Printf("Invalid request method: %s from %s", r.Method, r.RemoteAddr)
		http.Error(w, "Only POST and OPTIONS request are allowed", http.StatusMethodNotAllowed)
		return
	}
	authorizationHeader := r.Header.Get("Authorization")

	//Bearer kontrolu
	var tokenString string
	if authorizationHeader != "" {
		parts := strings.Split(authorizationHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	}

	// JWT'yi dogrula.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Geçersiz JWT", http.StatusUnauthorized)
		return
	}

	// JWT dogrulandı.
	// Kullanıcı adını JWT'den alma
	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	var newTicket Ticket
	if err := json.NewDecoder(r.Body).Decode(&newTicket); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var event EventType
	newTicket.Username = username
	err = EventCollection.FindOne(context.Background(), bson.M{"name": newTicket.Film}).Decode(&event)
	errorcode := check(newTicket.Film, newTicket.Quantity, newTicket.Username)
	if errorcode != "" {
		http.Error(w, errorcode, http.StatusBadRequest)
		return
	}
	newTicket.Price = 0
	newTicket.Date = time.Now()

	existingEvent, err := GetEventByName(newTicket.Film)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("Error checking event in DB: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingEvent == nil {
		http.Error(w, "No Film.Please check the list", http.StatusConflict)
		return
	}
	existingUser, err := GetUserByUsername(newTicket.Username)
	FreeTicket, err := GetInfoByUsername(newTicket.Username)

	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("Error checking username in DB: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingUser == nil {
		http.Error(w, "No user with this username.Please register at /user/save", http.StatusConflict)
		return
	}
	if FreeTicket.FreeTicket < 1 {
		http.Error(w, "Insufficient Free ticket to buy the ticket", http.StatusForbidden)
		return
	} else if FreeTicket.FreeTicket < newTicket.Quantity {
		http.Error(w, "You dont have enough free ticket.", http.StatusForbidden)
		return

	} else {
		_, err := TicketCollection.InsertOne(context.Background(), newTicket)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	update := bson.M{
		"$set": bson.M{"freeTicket": FreeTicket.FreeTicket - newTicket.Quantity},
	}

	filter := bson.M{"username": newTicket.Username}

	_, err = UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func BuyTicketByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodOptions {
		log.Printf("Invalid request method: %s from %s", r.Method, r.RemoteAddr)
		http.Error(w, "Only POST and OPTIONS request are allowed", http.StatusMethodNotAllowed)
		return
	}
	authorizationHeader := r.Header.Get("Authorization")

	//Bearer kontrolu
	var tokenString string
	if authorizationHeader != "" {
		parts := strings.Split(authorizationHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	}

	// JWT'yi dogrula.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Geçersiz JWT", http.StatusUnauthorized)
		return
	}

	// JWT dogrulandı.
	// Kullanıcı adını JWT'den alma
	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	var newTicket Ticket
	if err := json.NewDecoder(r.Body).Decode(&newTicket); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Redis'ten etkinlik bilgisini çekmeyi dene
	redisKey := "event:" + newTicket.Film
	var event EventType

	// Redis'te bilgi varsa, oradan al
	redisData, err := RedisClient.Get(context.Background(), redisKey).Bytes()
	if err == nil {
		// Redis'te veri bulundu, bu veriyi kullan
		if err := json.Unmarshal(redisData, &event); err != nil {
			http.Error(w, "Failed to parse Redis data", http.StatusInternalServerError)
			return
		}
	} else {
		// Redis'te bilgi yoksa, MongoDB'den al ve Redis'e yaz
		err := EventCollection.FindOne(context.Background(), bson.M{"name": newTicket.Film}).Decode(&event)
		if err != nil {
			http.Error(w, "There is no film with this name.", http.StatusInternalServerError)
			return
		}

		// MongoDB'den alınan veriyi JSON formatına dönüştür
		eventJSON, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "Failed to convert data to JSON", http.StatusInternalServerError)
			return
		}

		// Redis'e veriyi yaz
		err = RedisClient.Set(context.Background(), redisKey, eventJSON, 0).Err()
		if err != nil {
			http.Error(w, "Failed to write data to Redis", http.StatusInternalServerError)
			return
		}
	}

	// Diğer işlemleri yapabilirsiniz
	errorcode := check(newTicket.Film, newTicket.Quantity, username)
	if errorcode != "" {
		http.Error(w, errorcode, http.StatusBadRequest)
		return
	}

	// Bilet bilgisini oluştur ve fiyatı etkinlik fiyatıyla ayarla
	newTicket.Price = event.Price * newTicket.Quantity
	newTicket.Date = time.Now()
	newTicket.Username = username
	existingEvent, err := GetEventByName(newTicket.Film)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("Error checking film in DB: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingEvent == nil {
		http.Error(w, "No Film.Please check the list", http.StatusConflict)
		return
	}
	existingUser, err := GetUserByUsername(newTicket.Username)
	userBalance, err := GetInfoByUsername(newTicket.Username)

	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("Error checking username in DB: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingUser == nil {
		http.Error(w, "No user with this username.Please register at /user/save", http.StatusConflict)
		return
	}
	if userBalance.Balance < newTicket.Price {
		http.Error(w, "Insufficient balance to buy the ticket", http.StatusForbidden)
		return
	} else {
		_, err := TicketCollection.InsertOne(context.Background(), newTicket)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	update := bson.M{
		"$set": bson.M{"balance": userBalance.Balance - newTicket.Price},
	}

	filter := bson.M{"username": newTicket.Username}

	_, err = UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	SendMessage("usernameMQ", newTicket.Username, newTicket.Quantity)

}

func check(event string, quantity int, username string) string {
	if username == "" {
		return "username can not be empty"
	}
	if event == "" {
		return "event can not be empty"
	}
	if quantity == 0 {
		return "quantity can not be empty"
	}

	return ""
}
func GetUserByUsername(username string) (*Ticket, error) {
	var newTicket Ticket
	err := UserCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&newTicket)
	if err != nil {
		return nil, err
	}
	return &newTicket, nil
}
func GetInfoByUsername(username string) (*User, error) {
	var user User // Kullanıcının verilerini temsil eden bir User türünde değişken
	err := UserCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetEventByName(event string) (*Ticket, error) {
	var newTicket Ticket
	err := EventCollection.FindOne(context.Background(), bson.M{"name": event}).Decode(&newTicket)
	if err != nil {
		return nil, err
	}
	return &newTicket, nil
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodOptions {
		log.Printf("Invalid request method: %s from %s", r.Method, r.RemoteAddr)
		http.Error(w, "Only POST and OPTIONS requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// İstek verisini EventType yapısına çözümle
	var newEvent EventType
	if err := json.NewDecoder(r.Body).Decode(&newEvent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Etkinlik verisini MongoDB'ye kaydet
	insertResult, err := EventCollection.InsertOne(context.Background(), newEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Oluşturulan etkinliğin kimliğini yanıtla
	response := map[string]interface{}{"_id": insertResult.InsertedID}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}
