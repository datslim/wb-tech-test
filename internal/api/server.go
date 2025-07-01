package api

import (
	"encoding/json"
	"log"
	"net/http"
	"wb-tech-test/internal/cache"
	"wb-tech-test/internal/db"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// структура HTTP-сервера
type Server struct {
	router     *mux.Router
	database   *db.DB
	orderCache *cache.OrderCache
}

// функция для создания нового экземпляра сервера
func NewServer(database *db.DB, orderCache *cache.OrderCache) *Server {
	s := &Server{
		router:     mux.NewRouter(),
		database:   database,
		orderCache: orderCache,
	}
	s.setupRoutes() // настройка маршрутов
	return s
}

// функция настройки маршрутов
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/order/{order_uid}", s.getOrderByUID).Methods("GET")
}

// функция для получения заказа по его UID полученного из запроса
// в случае если заказ не найден в кэше, то запрос идет в БД
func (s *Server) getOrderByUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	order, ok := s.orderCache.Get(orderUID)
	if !ok {
		log.Printf("Заказ %s не найден в кэше", orderUID) // логируем что заказ не найден в кэше
		order, err := s.database.GetOrder(r.Context(), orderUID)
		if err != nil {
			log.Printf("Заказ %s не найден в БД: %s", orderUID, err) // логируем ошибку если заказ не найден в БД
			http.Error(w, "Заказ не найден", http.StatusNotFound)    // отправляем ответ о том что заказ не найден
			return
		}
		log.Printf("Заказ %s найден в БД", orderUID) // логируем что заказ найден в БД

		// сохраняем заказ в кэш
		s.orderCache.Set(order)
		log.Printf("Заказ %s сохранен в кэш", orderUID) // логируем что заказ сохранен в кэш
	}
	log.Printf("Заказ %s найден в кэше", orderUID) // логируем что заказ найден в кэше
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
