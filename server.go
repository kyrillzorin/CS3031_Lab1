package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

var db *bolt.DB

//Initialize boltDB used for blocklist storage
func initDB() error {
	var err error
	db, err = bolt.Open("proxy.db", 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("blocklist"))
		return nil
	})
	return err
}

// Main Function
func main() {
	initDB()
	router := httprouter.New()
	router.GET("/*path", handleHTTP)
	router.POST("/*path", handleHTTP)
	router.Handle("CONNECT", "/*path", handleHTTPS)
	server := http.Server{
		Addr:    ":8080",
		Handler: context.ClearHandler(router),
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln("Error: %v", err)
	}
}

//Handle HTTP
func handleHTTP(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	client := &http.Client{}
	req.RequestURI = ""
	fmt.Printf("%s\n", req.URL.String()) // Print request to console
	// Handle management console
	if strings.Contains(req.URL.String(), "://management.console") {
		management(w, req, ps)
		return
	}
	finished := false
	// Handle blocking
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocklist"))

		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if strings.Contains(req.URL.String(), string(k)) {
				fmt.Fprintf(w, "%s", "Blocked by proxy")
				finished = true
				return nil
			}
		}

		return nil
	})
	if finished {
		return
	}
	// Fetch page and return data to client
	resp, _ := client.Do(req)
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

//Blocklist Management Console
func management(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	if req.Method == "POST" {
		// Handle adding to blocklist
		url := req.PostFormValue("block")
		db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("blocklist"))
			if err != nil {
				return err
			}
			return b.Put([]byte(url), []byte("true"))
		})
		http.Redirect(w, req, "/", http.StatusFound)
		return
	} else if strings.Contains(req.URL.String(), "/blocklist") {
		// Display blocklist by reading from DB
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("blocklist"))
			c := b.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				fmt.Fprintf(w, "%s\n", k)
			}
			return nil
		})
	} else {
		http.ServeFile(w, req, "index.html") // Serve index
	}
}

//Handle HTTPS
func handleHTTPS(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	fmt.Printf("%s\n", req.URL.String()) // Print request to console
	hijack, _ := w.(http.Hijacker)
	client, _, _ := hijack.Hijack()
	host := req.URL.Host
	destination, _ := net.Dial("tcp", host)
	client.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
	go copyStream(destination, client)
	go copyStream(client, destination)
}

// Copy HTTP Headers
func copyHeaders(dst, src http.Header) {
	for k, _ := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
	dst.Del("Proxy-Connection")
	dst.Del("Proxy-Authenticate")
	dst.Del("Proxy-Authorization")
	dst.Del("Connection")
}

// Copy TCP stream
func copyStream(w, req net.Conn) {
	io.Copy(w, req)
	req.Close()
}
