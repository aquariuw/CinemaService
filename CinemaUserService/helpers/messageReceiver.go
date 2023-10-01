package helpers

import (
	. "CinemaUserService/models"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strconv"
	"strings"
)

const loyaltyThreshold = 100

func StartMessageReceiver() {
	conn := ConnectToRabbitMQ()

	// Kanal (channel) oluştur
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// Kuyruk oluştur veya var olanı kullan
	queue, err := ch.QueueDeclare(
		"usernameMQ",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Mesajları dinle
	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Mesajları dinlemeye başla. Çıkış yapmak için CTRL+C'ye basın.")

	// Mesajları işle
	for msg := range msgs {
		messageParts := strings.Split(string(msg.Body), ",")
		if len(messageParts) != 2 {
			log.Printf("Invalid message format: %s", string(msg.Body))
			continue
		}

		username := messageParts[0]
		quantityStr := messageParts[1]
		fmt.Printf("Received message: %s\n", username)

		// quantity bilgisini integer'a çevir
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			log.Printf("Error while converting quantity to integer: %s", err)
			continue
		}
		// Kullanıcıyı bul ve işle
		user, err := GetUserByUsername(username)
		if err != nil {
			log.Printf("Error while finding user: %s", err)
			continue
		}

		// Loyaltypuanı 10 artır
		user.LoyaltyPoint += quantity * 10

		// Loyaltypuan 100 ise Freeticket sayısını artır
		if user.LoyaltyPoint >= loyaltyThreshold {
			user.LoyaltyPoint = user.LoyaltyPoint - 100
			user.FreeTicket++ // Freeticket sayısını artır
			fmt.Printf("%s won free ticket!\n", user.Username)
		}

		// Kullanıcıyı güncelle
		err = UpdateUser(user)
		if err != nil {
			log.Printf("Updating Error: %s", err)
			continue
		}
	}
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := UserCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(user *User) error {
	filter := bson.M{"username": user.Username}
	update := bson.M{"$set": bson.M{"freeTicket": user.FreeTicket, "loyaltyPoint": user.LoyaltyPoint}}
	_, err := UserCollection.UpdateOne(context.TODO(), filter, update)
	return err
}
