package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// net/http based router
type route struct {
	pattern *regexp.Regexp
	verb    string
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, verb string, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, verb, handler})
}

func (h *RegexpHandler) HandleFunc(r string, v string, handler func(http.ResponseWriter, *http.Request)) {
	re := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{re, v, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) && route.verb == r.Method {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

// store "context" values and connections in the server struct
type Server struct {
	db    *sql.DB
	table string
}

// InitDB connects to the database
func (server *Server) InitDB() error {
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlPort := os.Getenv("MYSQL_PORT")

	mysqlTable := os.Getenv("MYSQL_TABLE")

	mysqlConnectString := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)
	log.Printf("Trying to connect to mysql %v", mysqlHost)
	db, err := sql.Open("mysql", mysqlConnectString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}
	db.SetMaxIdleConns(100)
	server.db = db
	if mysqlTable == "" {
		mysqlTable = "gotodo"
	}
	server.table = mysqlTable
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v (`key` VARCHAR(50), value TEXT, PRIMARY KEY(`key`)) CHARACTER SET utf8 COLLATE utf8_unicode_ci", server.table)
	_, err = db.Exec(sql)
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		db.Close()
		return err
	}
	return nil
}

// SetValue sets a json string into database
func (server *Server) SetValue(w http.ResponseWriter, r *http.Request) {
	pathItems := strings.Split(r.URL.Path, "/")
	key := pathItems[len(pathItems)-1]
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	dataStr := string(data)
	sql := fmt.Sprintf("INSERT INTO %v (`key`, `value`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `value`=?", server.table)
	_, err = server.db.Exec(sql, key, dataStr, dataStr)
	log.Printf("Inserted %v=%v, err=%v", key, dataStr, err)
	w.WriteHeader(http.StatusOK)
}

// GetValue gets a json string from database
func (server *Server) GetValue(w http.ResponseWriter, r *http.Request) {
	pathItems := strings.Split(r.URL.Path, "/")
	key := pathItems[len(pathItems)-1]

	sql := fmt.Sprintf("SELECT value from %v WHERE `key`=?", server.table)
	rows, err := server.db.Query(sql, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	var value string
	if rows.Next() {
		rows.Scan(&value)
		log.Printf("Retrieved %v=%v", key, value)
	} else {
		log.Printf("Not found %v", key)
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, value)
}

func main() {

	server := &Server{}
	for retry := 0; retry < 10; retry++ {
		err := server.InitDB()
		if err == nil {
			break
		} else {
			log.Printf("Failed to initialize database. Retry %v", retry)
			time.Sleep(1 * time.Second)
		}
	}
	defer server.db.Close()

	reHandler := new(RegexpHandler)

	reHandler.HandleFunc("/set_value/.+$", "POST", server.SetValue)
	reHandler.HandleFunc("/get_value/.+$", "GET", server.GetValue)

	reHandler.HandleFunc("/static/.*", "GET", server.assets)
	reHandler.HandleFunc("/", "GET", server.homepage)

	fmt.Println("Starting server on port 3000")
	http.ListenAndServe(":3000", reHandler)
}

// simple HTML/JS/CSS pages

func (s *Server) homepage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "static/index.html", http.StatusTemporaryRedirect)
}

func (s *Server) assets(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, req.URL.Path[1:])
}
