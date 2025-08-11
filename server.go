package main

import (
	"fmt"
	"net/http"
	"os"
)

func Middleware(f http.HandlerFunc) http.HandlerFunc {
	return logRequestHandler(f)
}

func ServeTopics(fs http.Handler) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("token")
		token, ok := tokens.GetToken(tokenString)
		if !ok || !token.HasPermission(AdminFlag) {
			http.Error(w, "Token not found or does not have permission", http.StatusUnauthorized)
			return
		}
		fs.ServeHTTP(w, r)
	}
	return Middleware(handler)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func main() {
	http.HandleFunc("GET /create_topic/{topicName}", Middleware(CreateNewTopic))
	http.HandleFunc("GET /pull_topic/{topicName}", Middleware(PullEvent))
	http.HandleFunc("POST /push_topic/{topicName}", Middleware(PushEvent))
	http.HandleFunc("POST /new_token", Middleware(CreateToken))
	http.HandleFunc("GET /remove_token", Middleware(DeleteTokenEP))
	http.HandleFunc("PATCH /subscribe/{topicName}", Middleware(Subscribe))
	http.HandleFunc("GET /token_desc", Middleware(GetCurrentToken))
	fs := http.StripPrefix("/topics/", http.FileServer(http.Dir("./topics")))
	http.Handle("GET /topics/", ServeTopics(fs))

	fmt.Println("loading topics")
	LoadTopics()

	if fileExists("tokens") {
		fmt.Println("loading tokens")
		LoadTokens()
	} else {
		adminToken := InitTokens()
		fmt.Println("initializing tokens", adminToken.Token)
	}

	fmt.Println("Listening on port 8181")
	panic(http.ListenAndServe(":8181", nil))
}
