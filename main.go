package main

import (
	"fmt"
    "net/http"
    "encoding/json"

    "github.com/gorilla/mux"
	"github.com/mxk/go-sqlite/sqlite3"
)


type User struct {
	UserID    int     `json:UserID`
	UserName  string  `json:UserName`
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	
	c1, _ := sqlite3.Open("users.db")
	
	sql := "SELECT UserID, UserName FROM users"
	
	for s, err := c1.Query(sql); err == nil; err = s.Next() {
		var user User
		s.Scan(&user.UserID, &user.UserName)
		
		json.NewEncoder(w).Encode(user)
		
	}
	

	c1.Close()
}

func NewUser(w http.ResponseWriter, r *http.Request) {
	
	c2, _ := sqlite3.Open("users.db")

	var user User
	
	json.NewDecoder(r.Body).Decode(&user)

	if user.UserName == "" {
		fmt.Fprintf(w, "Introduce un nombre de usuario, por favor")
		return
	}
	

	sql := "SELECT MAX(UserID)+1 AS LastUserID FROM users"
	
	var LastUserID int
	s, err := c2.Query(sql)
	if err != nil {
		fmt.Fprintf(w, "Error obteniendo el último ID")
		return
	}
	s.Scan(&LastUserID)

	args := sqlite3.NamedArgs{"$LastUserID": LastUserID, "$UserName": user.UserName}
	err = c2.Exec("INSERT INTO users (UserID, UserName) VALUES ($LastUserID, $UserName)", args)
	if err != nil {
		fmt.Fprintf(w, "Error añadiendo el usuario %s", err)
		return
	}

	c2.Close()

	user.UserID = LastUserID
	json.NewEncoder(w).Encode(user)

}

func main () {

	r := mux.NewRouter()

	r.HandleFunc("/user", GetUsers).Methods("GET")
	r.HandleFunc("/user", NewUser).Methods("POST")


	http.ListenAndServe(":80", r)

}