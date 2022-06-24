package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
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
	ID            string `json:"id"`
	Firstname     string `json:"firstname"`
	Lastname      string `json:"lastname"`
	Age           int    `json:"age"`
	Payedvacation int    `json:"payedvacation"`
}

// albums slice to seed record album data.
var users = []User{
	{ID: "1", Firstname: "雄人", Lastname: "寺内", Age: 20, Payedvacation: 10},
	{ID: "2", Firstname: "鷹哉", Lastname: "清水", Age: 20, Payedvacation: 20},
	{ID: "3", Firstname: "", Lastname: "本田", Age: 20, Payedvacation: 30},
	{ID: "4", Firstname: "寛太", Lastname: "梅澤", Age: 20, Payedvacation: 40},
}

type Post struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func main() {
	// Initialize template parameters.
	service := os.Getenv("K_SERVICE")
	if service == "" {
		service = "???"
	}

	revision := os.Getenv("K_REVISION")
	if revision == "" {
		revision = "???"
	}

	// router := gin.Default()
	// router.GET("/albums", getAlbumsgin)
	// router.GET("/users", getUsersgin)

	// router.Run("localhost:8080")

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
	r := mux.NewRouter()

	r.HandleFunc("/", helloRunHandler)

	r.HandleFunc("/test", testRunHandler)
	r.HandleFunc("/albums", getAlbums)
	r.HandleFunc("/users", getUsers)

	// Restrict the request handler to http/https.
	r.HandleFunc("/secure", SecureHandler).Schemes("https")
	r.HandleFunc("/insecure", InsecureHandler).Schemes("http")

	r.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
		var user User
		json.NewDecoder(r.Body).Decode(&user)

		fmt.Fprintf(w, "%s %s is %d years old!", user.Firstname, user.Lastname, user.Age)
	})

	r.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		// peter := User{
		// 	Firstname: "John",
		// 	Lastname:  "Doe",
		// 	Age:       25,
		// }

		json.NewEncoder(w).Encode(posts)
	})

	// fs := http.FileServer(http.Dir("./assets"))
	// r.Handle("/assets/", http.StripPrefix("/assets/", fs))
	// r.PathPrefix("/").Handler(http.FileServer(http.Dir("./assets/")))
	s := http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/")))
	r.PathPrefix("/assets/").Handler(s)

	// PORT environment variable is provided by Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		title := vars["title"]
		page := vars["page"]

		fmt.Fprintf(w, "You've requested the book: %s on page %s\n", title, page)
	})

	http.Handle("/", r)

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

func SecureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "SecureHandler, you've requested: %s\n", r.URL.Path)
}

func InsecureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "InsecureHandler, you've requested: %s\n", r.URL.Path)
}

// getAlbums responds with the list of all albums as JSON.
func getAlbumsgin(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// getAlbums responds with the list of all albums as JSON.
func getUsersgin(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, users)
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(w http.ResponseWriter, r *http.Request) {
	e := json.NewEncoder(w)
	e.SetIndent("", strings.Repeat(" ", 4))
	e.Encode(albums)
}

// getAlbums responds with the list of all albums as JSON.
func getUsers(w http.ResponseWriter, r *http.Request) {
	e := json.NewEncoder(w)
	e.SetIndent("", strings.Repeat(" ", 4))
	e.Encode(users)
}
