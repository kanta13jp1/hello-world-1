package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// templateData provides template parameters.
type templateData struct {
	Service  string
	Revision string
}

// Variables used to generate the HTML page.
var (
	data templateData
	tmpl *template.Template
)

type User struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
}

type Post struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		title := vars["title"]
		page := vars["page"]

		fmt.Fprintf(w, "You've requested the book: %s on page %s\n", title, page)
	})

	http.ListenAndServe(":80", r)

	// Initialize template parameters.
	service := os.Getenv("K_SERVICE")
	if service == "" {
		service = "???"
	}

	revision := os.Getenv("K_REVISION")
	if revision == "" {
		revision = "???"
	}

	// Prepare template for execution.
	tmpl = template.Must(template.ParseFiles("index.html"))
	data = templateData{
		Service:  service,
		Revision: revision,
	}

	// posts.json を読み込む
	postsJsonFile, error := os.Open("./assets/posts.json")

	// posts.json の読み込みに失敗した場合
	if error != nil {
		log.Fatal(error)
	}

	// defer で postsJsonFile を閉じる
	defer postsJsonFile.Close()

	// postsJsonFile を読み込みパースする
	postsByteValue, _ := ioutil.ReadAll(postsJsonFile)
	var posts []Post
	json.Unmarshal(postsByteValue, &posts)

	fmt.Println(posts) // [{132 Ditto} {133 Eevee} {143 Snorlax}]

	// Define HTTP server.
	http.HandleFunc("/", helloRunHandler)

	http.HandleFunc("/test", testRunHandler)

	http.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
		var user User
		json.NewDecoder(r.Body).Decode(&user)

		fmt.Fprintf(w, "%s %s is %d years old!", user.Firstname, user.Lastname, user.Age)
	})

	http.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		// peter := User{
		// 	Firstname: "John",
		// 	Lastname:  "Doe",
		// 	Age:       25,
		// }

		json.NewEncoder(w).Encode(posts)
	})

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// PORT environment variable is provided by Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Print("Hello from Cloud Run! The container started successfully and is listening for HTTP requests on $PORT")
	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// helloRunHandler responds to requests by rendering an HTML page.
func helloRunHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		log.Printf("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

func testRunHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
}
