package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tjgurwara99/mdgen/include"
)

func main() {
	root := os.DirFS("./")
	handler := include.FileServer(root)
	if err := http.ListenAndServe(":8000", handler); err != nil {
		log.Fatal(err)
	}
}
