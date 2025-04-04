package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mxk/go-sqlite/sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID       int    `json:UserID`
	UserName     string `json:UserName`
	UserEmail    string `json:UserEmail`
	UserPassword string `json:UserPassword`
}

type userLogin struct {
	UserEmail    string `json:UserEmail`
	UserPassword string `json:UserPassword`
}

type getUser struct {
	UserID   int    `json:UserID`
	UserName string `json:UserName`
}

type newUser struct {
	UserName     string `json:UserName`
	UserEmail    string `json:UserEmail`
	UserPassword string `json:Password`
}

type deleteUser struct {
	UserID int `json:UserID`
}

type updateUser struct {
	UserID   int    `json:UserID`
	UserName string `json:UserName`
}

func CreateHash(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 14)
	return string(bytes), err
}

func CheckPass(pass string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

func GetUsers(w http.ResponseWriter, r *http.Request) {

	var user getUser

	c1, _ := sqlite3.Open("users.db")

	sql := "SELECT UserID, UserName FROM users"

	for s, err := c1.Query(sql); err == nil; err = s.Next() {
		s.Scan(&user.UserID, &user.UserName)

		json.NewEncoder(w).Encode(user)

	}

	c1.Close()
}

func NewUser(w http.ResponseWriter, r *http.Request) {

	c2, _ := sqlite3.Open("users.db")

	var newUser newUser

	var user getUser

	json.NewDecoder(r.Body).Decode(&newUser)

	if newUser.UserName == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Introduce un nombre de usuario, por favor")
		return
	}

	if newUser.UserEmail == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Es necesario introducir un email para registrarse")
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
	s.Close()
	if LastUserID == 0 {
		LastUserID = 1
	}

	pass, err := CreateHash(newUser.UserPassword)
	if err != nil {
		fmt.Fprintf(w, "Error generating hash", err)
	}
	args := sqlite3.NamedArgs{"$UserEmail": newUser.UserEmail}

	sql = "SELECT UserEmail FROM users WHERE UserEmail='$UserEmail'"

	s, err = c2.Query(sql, args)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "El email ya existe")
		return
	}
	s.Close()

	args = sqlite3.NamedArgs{"$LastUserID": LastUserID, "$UserName": newUser.UserName, "$UserEmail": newUser.UserEmail, "$UserPassword": pass}
	err = c2.Exec("INSERT INTO users (UserID, UserName, UserEmail, UserPassword) VALUES ($LastUserID, $UserName, $UserEmail, $UserPassword)", args)
	if err != nil {
		fmt.Fprintf(w, "Error añadiendo el usuario %s", err)
		return
	}

	c2.Close()

	user.UserID = LastUserID
	user.UserName = newUser.UserName
	json.NewEncoder(w).Encode(user)

}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	c3, _ := sqlite3.Open("users.db")

	var user deleteUser

	json.NewDecoder(r.Body).Decode(&user)

	args := sqlite3.NamedArgs{"$UserID": user.UserID}

	s, err := c3.Query("SELECT UserID FROM users WHERE UserID=$UserID", args)
	if err != nil {
		fmt.Fprintf(w, "Error usuario no existe!")
		return
	}
	s.Close()

	sql := "DELETE FROM users WHERE UserID=$UserID"
	err = c3.Exec(sql, args)
	c3.Commit()
	if err != nil {
		fmt.Fprintf(w, "Error borrando el usuario: %d", err)
		return
	}

	fmt.Fprintf(w, "Usuario borrado con exito")
	c3.Close()
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user updateUser

	c4, _ := sqlite3.Open("users.db")

	json.NewDecoder(r.Body).Decode(&user)

	args := sqlite3.NamedArgs{"$UserID": user.UserID}

	s, err := c4.Query("SELECT UserID FROM users WHERE UserID=$UserID", args)
	if err != nil {
		fmt.Fprintf(w, "El usuario no existe!")
		return
	}
	s.Close()

	args = sqlite3.NamedArgs{"$UserID": user.UserID, "$UserName": user.UserName}
	err = c4.Exec("UPDATE users SET UserName=$UserName WHERE UserID=$UserID", args)
	if err != nil {
		fmt.Fprintf(w, "Error actualizando el usuario!")
		return
	}

	fmt.Fprintf(w, "Usuario actualizado correctamente!")

	c4.Close()

}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var userLogin userLogin
	var userPassword string

	c5, _ := sqlite3.Open("users.db")

	json.NewDecoder(r.Body).Decode(&userLogin)

	args := sqlite3.NamedArgs{"$UserEmail": userLogin.UserEmail}

	sql := "SELECT UserPassword FROM users WHERE UserEmail=$UserEmail"

	s, err := c5.Query(sql, args)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error obteniendo la contraseña", err)
		return
	}
	s.Scan(&userPassword)

	if CheckPass(userLogin.UserPassword, userPassword) {
		fmt.Fprintf(w, "Contraseña correcta!")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Contraseña incorrecta!")
	}
}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/user", GetUsers).Methods("GET")
	r.HandleFunc("/user", NewUser).Methods("POST")
	r.HandleFunc("/user", DeleteUser).Methods("DELETE")
	r.HandleFunc("/user", UpdateUser).Methods("PUT")
	r.HandleFunc("/login", LoginUser).Methods("GET")

	http.ListenAndServe(":80", r)

}
