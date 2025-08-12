package app

import (
	"context"
	"fmt"
	"go-postgres-docker/internal/cache"
	"go-postgres-docker/internal/config"
	"go-postgres-docker/internal/db"
	"go-postgres-docker/internal/handler"
	"go-postgres-docker/internal/repository"
	"go-postgres-docker/internal/kafka"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	// Загрузка конфига
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ициализируем БД
	database, err := db.Init(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Миграция БД
	if err := db.Migrate(database, cfg); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Создаём новый кэш
	repo := repository.NewProductRepository(database)
	orderCache := cache.NewOrderCache()

	// Загрузка кэша из БД
	orders, err := repo.GetOrders()
	if err != nil {
		log.Fatalf("Failed to preload cache: %v", err)
	}
	for _, o := range orders {
		full, err := repo.GetOrderByID(o.OrderUID)
		if err == nil {
			orderCache.Set(full)
		}
	}

	// Запуск producer для отправки сообщений в сonsumer
	// Producer подключен для "удобной" отправки через сайт
	kafkaProducer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)	
	fmt.Print("Kafka_producer", kafkaProducer)

	defer kafkaProducer.Close()

	h := handler.NewProductHandler(repo, orderCache, kafkaProducer)

	// Запускаем consumer (graceful cancellable)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// StartConsumer будет блокирующим поэтому запускаем в горутине
		kafka.StartConsumer(ctx, cfg, repo, orderCache)
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Маршруты
	http.HandleFunc("/orders", h.GetOrders)
	http.HandleFunc("/order/", h.GetOrderByID)
	http.Handle("/", http.FileServer(http.Dir("web")))

	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateOrder(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("Server running at %s", addr)
		serverErrCh <- http.ListenAndServe(addr, nil)
	}()

	select {
	case sig := <-sigCh:
		log.Printf("Received signal %s, shutting down...", sig)
		cancel()
	case err := <-serverErrCh:
		log.Printf("HTTP server stopped: %v", err)
		cancel()
	}
}
