package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/tjgurwara99/mdgen/include"
)

func main() {
	dir := flag.String("d", "./", "Directory to serve over the server")
	flag.Parse()

	root := os.DirFS(*dir)
	handler := include.FileServer(root)
	if err := http.ListenAndServe(":8000", handler); err != nil {
		log.Fatal(err)
	}
}
