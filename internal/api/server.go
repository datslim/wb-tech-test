package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"wb-tech-test/internal/cache"
	"wb-tech-test/internal/db"
	"wb-tech-test/internal/model"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

// структура HTTP-сервера
type Server struct {
	router      *mux.Router
	database    *db.DB
	orderCache  *cache.OrderCache
	kafkaWriter *kafka.Writer
}

// функция для создания нового экземпляра сервера
func NewServer(database *db.DB, orderCache *cache.OrderCache) *Server {
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"wb-kafka:9092"},
		Topic:   "orders",
	})

	s := &Server{
		router:      mux.NewRouter(),
		database:    database,
		orderCache:  orderCache,
		kafkaWriter: kafkaWriter,
	}
	s.setupRoutes() // настройка маршрутов
	return s
}

// функция настройки маршрутов
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/order/{order_uid}", s.getOrderByUID).Methods("GET") // маршрут для получения заказа по его UID
	s.router.HandleFunc("/orders", s.handleKafkaProduce).Methods("POST")      // маршрут для отправки заказа в Kafka (для тестирования)
}

// функция для отправки заказа в Kafka (для тестирования)
func (s *Server) handleKafkaProduce(w http.ResponseWriter, r *http.Request) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		log.Printf("[API] Ошибка при декодировании JSON для отправки в Kafka: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	msg, err := json.Marshal(order)
	if err != nil {
		log.Printf("[API] Ошибка при маршалинге заказа для отправки в Kafka: %v", err)
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}
	err = s.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: msg,
		},
	)
	if err != nil {
		log.Printf("[API] Ошибка при отправке заказа в Kafka: %v", err)
		http.Error(w, "kafka error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// функция для получения заказа по его UID полученного из запроса
// в случае если заказ не найден в кэше, то запрос идет в БД
func (s *Server) getOrderByUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	order, ok := s.orderCache.Get(orderUID)
	if !ok {
		log.Printf("[API] Заказ %s не найден в кэше", orderUID) // логируем что заказ не найден в кэше
		order, err := s.database.GetOrder(r.Context(), orderUID)
		if err != nil {
			log.Printf("[API] Заказ %s не найден в БД: %s", orderUID, err) // логируем ошибку если заказ не найден в БД
			http.Error(w, "Заказ не найден", http.StatusNotFound)          // отправляем ответ о том что заказ не найден
			return
		}
		log.Printf("[API] Заказ %s найден в БД", orderUID) // логируем что заказ найден в БД

		// сохраняем заказ в кэш
		s.orderCache.Set(order)
		log.Printf("[API] Заказ %s сохранен в кэш", orderUID) // логируем что заказ сохранен в кэш
	}
	log.Printf("[API] Заказ %s найден в кэше", orderUID) // логируем что заказ найден в кэше
	// если заказ найден в кэше, то отправляем его в ответе
	json.NewEncoder(w).Encode(order)

}

// функция для запуска сервера
func (s *Server) Start(addr string) error {

	// используем gorilla/handlers для разрешения заголовков CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),                      // разрешаем все источники
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}), // разрешаем методы GET, POST, OPTIONS
	)(s.router)

	return http.ListenAndServe(addr, corsHandler)
}
