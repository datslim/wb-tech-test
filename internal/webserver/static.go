package webserver

import (
	"log"
	"net/http"
)

func Start() {
	// Отдаём содержимое папки web/ по адресу /
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	log.Println("Web server started at :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
