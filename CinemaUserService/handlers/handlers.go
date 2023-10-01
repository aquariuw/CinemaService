package handlers

import (
	. "CinemaUserService/helpers"
	. "CinemaUserService/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"unicode"
)

var (
	signingKey = []byte("AbDuRRaHmAn")
)

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the Cinema Ticket User Service...")
}
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodOptions {
		log.Printf("Invalid request method: %s from %s", r.Method, r.RemoteAddr)
		http.Error(w, "Only POST and OPTIONS request are allowed", http.StatusMethodNotAllowed)
		return
	}
	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	errorcode := check(newUser.Username)
	if errorcode != "" {
		http.Error(w, errorcode, http.StatusBadRequest)
		return
	}
	// Kullanıcı adının veritabanında var olup olmadığını kontrol et
	existingUser, err := GetUserByUserName(newUser.Username)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("Error checking username in DB: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kullanıcı adı zaten varsa hata döndür
	if existingUser != nil {
		http.Error(w, "This username already exist", http.StatusConflict)
		return
	}
	newUser.Balance = 500

	_, err = UserCollection.InsertOne(context.Background(), newUser)
	if err != nil {
		log.Printf("Error inserting user into DB: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("User saved Succesfully: %s", newUser.Username)
	w.WriteHeader(http.StatusCreated)

}
func GetAllTicketByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("username")
	if name == "" {
		http.Error(w, "Name parameter is missing", http.StatusBadRequest)
		return
	}

	filter := bson.M{"username": name}
	cursor, err := TicketCollection.Find(context.Background(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var tickets []Ticket
	for cursor.Next(context.Background()) {
		var ticket Ticket
		if err := cursor.Decode(&ticket); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tickets = append(tickets, ticket)
	}

	// _id'yi stringe çevirip JSON verisine ekliyoruz
	var responseTickets []map[string]interface{}
	for _, ticket := range tickets {
		responseTicket := map[string]interface{}{
			"ID":       ticket.ID.Hex(), // ObjectID'yi hex formatında alıyoruz
			"Username": ticket.Username,
			"Film ":    ticket.Film,
			"quantity": ticket.Quantity,
			"date":     ticket.Date,
		}
		responseTickets = append(responseTickets, responseTicket)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseTickets)
}
func GetToken(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	existingUser, err := GetUserByUserName(username)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("Error checking username in DB: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kullanıcı adı yok
	if existingUser == nil {
		http.Error(w, "There is no user with this username", http.StatusConflict)
		return
	}
	// Kullanıcı kimlik doğrulandıktan sonra JWT oluşturulur.
	token := jwt.New(jwt.SigningMethodHS256)

	// JWT içeriği olarak kullanıcı adını ekleyin.
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Tokenin geçerlilik süresi (örneğin, 1 gün)

	// JWT'yi imzala ve string formatına dönüştür.
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"token": tokenString}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
func GetUserByUserName(username string) (*User, error) {
	var user User
	err := UserCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func check(username string) string {
	if username == "" {
		return "Username can not be empty"
	}
	if len(username) > 16 {
		return "Username can not be longer than 16 characters"
	}
	if !unicode.IsUpper([]rune(username)[0]) {
		return "Username must start with uppercase letter"
	}
	if !isAlphabetic(username) {
		return "Username can only contain letters"
	}
	return ""
}
func isAlphabetic(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
func GetbyIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectID}
	var user User
	err = UserCollection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("User not found for ID: %s", id)
			http.Error(w, "User Not Found", http.StatusNotFound)

		} else {
			log.Printf("Error retrieving user for ID %s: %s", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
func GetbyNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("username")
	if name == "" {
		http.Error(w, "Name parameter is missing", http.StatusBadRequest)
		return
	}

	filter := bson.M{"username": name}
	var user User
	err := UserCollection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User Not Found", http.StatusNotFound)

		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
func GetHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	cursor, err := UserCollection.Find(context.Background(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var users []User
	for cursor.Next(context.Background()) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		users = append(users, user)
	}

	// _id'yi stringe çevirip JSON verisine ekliyoruz
	var responseUsers []map[string]interface{}
	for _, user := range users {
		responseUser := map[string]interface{}{
			"ID":           user.ID.Hex(), // ObjectID'yi hex formatında alıyoruz
			"Username":     user.Username,
			"loyalt Point": user.LoyaltyPoint,
			"free Ticket":  user.FreeTicket,
			"balance":      user.Balance,
		}
		responseUsers = append(responseUsers, responseUser)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseUsers)
}

func DeletebyIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, fmt.Sprintf("Only DELETE requests are allowed. Your method is %s", r.Method), http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectID}
	_, err = UserCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func DeletebyNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, fmt.Sprintf("Only DELETE requests are allowed. Your method is %s", r.Method), http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("username")
	if name == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	filter := bson.M{"username": name}
	_, err := UserCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdatebyIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, fmt.Sprintf("Only PUT requests are allowed, your method is %s", r.Method), http.StatusMethodNotAllowed)

		return
	}

	var updatedUser User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"username": updatedUser.Username}}

	_, err = UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func UpdatebyNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, fmt.Sprintf("Only PUT requests are allowed, your method is %s", r.Method), http.StatusMethodNotAllowed)

		return
	}

	var updatedUser User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := r.URL.Query().Get("username")
	if name == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	filter := bson.M{"username": name}
	update := bson.M{"$set": bson.M{"username": updatedUser.Username}}

	_, err := UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
