package webserver

import (
	"log"
	"net/http"
)

// простой веб-сервер для отдачи статических данных от нашего API-сервера
func Start(addr string) {
	// Отдаём содержимое папки web/ по адресу /
	fs := http.FileServer(http.Dir("./frontend/"))
	http.Handle("/", fs)

	log.Printf("Web server запущен на порту %s", addr)

	log.Fatal(http.ListenAndServe(":"+addr, nil))
}
