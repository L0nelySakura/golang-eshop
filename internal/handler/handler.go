package handler

import (
	"encoding/json"
	"go-postgres-docker/internal/cache"
	"go-postgres-docker/internal/model"
	"go-postgres-docker/internal/kafka"
	"go-postgres-docker/internal/repository"
	"net/http"
	"strings"
)

type OrderHandler struct {
	repo  *repository.OrderRepository
	cache *cache.OrderCache
	kafkaProducer *kafka.Producer
}

func NewProductHandler(repo *repository.OrderRepository, cache *cache.OrderCache, producer *kafka.Producer) *OrderHandler {
	return &OrderHandler{repo: repo, cache: cache, kafkaProducer: producer}
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.repo.GetOrders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	uid := strings.TrimPrefix(r.URL.Path, "/order/")
	if uid == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	if order, found := h.cache.Get(uid); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
		return
	}

	order, err := h.repo.GetOrderByID(uid)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	h.cache.Set(order)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Сериализуем заказ в JSON
	data, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "failed to marshal order", http.StatusInternalServerError)
		return
	}

	// Отправляем в Kafka
	err = h.kafkaProducer.Publish(r.Context(), []byte(order.OrderUID), data)
	if err != nil {
		http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"order sent to kafka"}`))
}

