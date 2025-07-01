package api

import (
	"encoding/json"
	"net/http"
	"wb-tech-test/internal/cache"
	"wb-tech-test/internal/db"

	"github.com/gorilla/mux"
)

// структура HTTP-сервера
type Server struct {
	router     *mux.Router
	database   *db.DB
	orderCache *cache.OrderCache
}

// конструктор для создания нового сервера
func NewServer(database *db.DB, orderCache *cache.OrderCache) *Server {
	s := &Server{
		router:     mux.NewRouter(),
		database:   database,
		orderCache: orderCache,
	}
	s.setupRoutes()
	return s
}

// функция для настройки маршрутизации
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/order/{order_uid}", s.getOrderByUID).Methods("GET")
}

func (s *Server) getOrderByUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	order, ok := s.orderCache.Get(orderUID)
	if !ok {
		order, err := s.database.GetOrder(r.Context(), orderUID)
		if err != nil {
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}

		s.orderCache.Set(order)
	}

	json.NewEncoder(w).Encode(order)

}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
