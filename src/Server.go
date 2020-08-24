package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// Env asd
type Env struct {
	db *sql.DB
}

// ReplenishmentStruct asd
type ReplenishmentStruct struct {
	ID    int     `json:"id"`
	Value float64 `json:"value"`
}

// WithdravalStruct asd
type WithdravalStruct struct {
	ID    int     `json:"id"`
	Value float64 `json:"value"`
}

// TransferStruct asd
type TransferStruct struct {
	IDFrom int     `json:"id_from"`
	IDTo   int     `json:"id_to"`
	Value  float64 `json:"value"`
}

// UserStruct asd
type UserStruct struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// RequestData asd
type RequestData struct {
	ReplenishmentData ReplenishmentStruct `json:"replenishment"`
	WithdravalData    WithdravalStruct    `json:"withdraval"`
	TransferData      TransferStruct      `json:"transfer"`
	User              UserStruct          `json:"add_user"`
}

//RequestStruct is data struct, that will contain json request fields
type RequestStruct struct {
	Method string      `json:"method"`
	Data   RequestData `json:"data"`
}

//AddUser method is created to add users
func (env *Env) AddUser(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error read body")
		return
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		fmt.Println("error decode json")
		return
	}
	result, err := env.db.Exec("insert into users_info (first_name, last_name, reg_date) values(?,?,now())",
		request.Data.User.FirstName, request.Data.User.LastName)
	if err != nil {
		fmt.Println("error add user to database")
		fmt.Println(err)
		return
	}
	fmt.Println(result.LastInsertId())
}

func main() {
	db, err := sql.Open("mysql", "go_server:pass@/balance_service")
	if err != nil {
		fmt.Println("error open database")
		fmt.Println(err)
	}
	env := &Env{db: db}
	http.HandleFunc("/AddUser", env.AddUser)
	http.ListenAndServe(":8080", nil)
}
