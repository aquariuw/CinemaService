package helpers

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func SendMessage(queueName string, username string, quantity int) {
	ch, err := RabbitMQConn.Channel()
	if err != nil {
		log.Fatalf("Kanal oluşturulurken hata oluştu: %s", err)
	}
	defer ch.Close()

	// Kuyruğu oluştur (eğer daha önce oluşturulmadıysa)
	_, err = ch.QueueDeclare(
		queueName, // Kuyruk adı
		false,     // Kalıcı (durable) değil
		false,     // Kuyruk işlemcisi (exclusive) değil
		false,     // Kuyruk otomatik silinmesin
		false,     // Kuyruk yalnızca bu bağlantı ile kullanılmasın
		nil,       // Diğer özel argümanlar
	)
	if err != nil {
		log.Fatalf("Kuyruk oluştururken hata oluştu: %s", err)
	}

	// Mesajı oluştur ve kuyruğa gönder
	message := fmt.Sprintf("%s,%d", username, quantity)
	err = ch.Publish(
		"",        // Exchange (değiştirme) kullanmayacağız
		queueName, // Kuyruk adı
		false,     // Kuyruğa yayınlama (mandatory) değil
		false,     // Yayınlanan mesajı kuyruk işlemcileri almasın
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		log.Fatalf("Mesajı kuyruğa gönderirken hata oluştu: %s", err)
	}
}
