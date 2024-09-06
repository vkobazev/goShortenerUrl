package main

import (
	"io"
	"math/rand"
	"net/http"
)

var urls = map[string]string{}

func main() {
	http.HandleFunc("/", urlShortner)

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}

func urlShortner(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case http.MethodPost:
			// Handle POST request
			req.URL.Scheme = "http"
			shortURLPath := "/" + RandomString(6)
			shortURL := req.URL.Scheme + "://" + req.Host + shortURLPath

			body, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(res, "No Url in Body", http.StatusBadRequest)
			}
			urls[shortURLPath] = string(body)

			// Writing Response
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusCreated)
			io.WriteString(res, shortURL)

		case http.MethodGet:
			// Handle GET request
			path := req.URL.Path

			// Writing Response
			_, state := urls[path]
			if !state {
				http.Error(res, "No Url in list of Short Urls", http.StatusNotFound)
			} 
			res.Header().Set("Location", urls[path])
			res.WriteHeader(http.StatusTemporaryRedirect)

		default:
			http.Error(res, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func RandomString(num int) string {
    var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
 
    str := make([]rune, num)
    for i := range str {
        str[i] = letters[rand.Intn(len(letters))]
    }
    return string(str)
}
