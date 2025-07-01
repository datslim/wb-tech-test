package webserver

import (
	"log"
	"net/http"
)

// простой веб-сервер для отдачи статических данных от нашего API-сервера
func Start() {
	// Отдаём содержимое папки web/ по адресу /
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	log.Println("Web server started at :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
