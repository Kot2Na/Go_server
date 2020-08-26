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
	ID    int64   `json:"id"`
	Value float64 `json:"value"`
}

// WithdravalStruct asd
type WithdravalStruct struct {
	ID    int64   `json:"id"`
	Value float64 `json:"value"`
}

// TransferStruct asd
type TransferStruct struct {
	IDFrom int64   `json:"id_from"`
	IDTo   int64   `json:"id_to"`
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

//ResponseData asd
type ResponseData struct {
	Message       string
	IDUser        int64
	IDTrunsuction int64
}

//ResponseStruct asd
type ResponseStruct struct {
	Status int
	Method string
	Data   ResponseData
}

//ReadRequest asd
func ReadRequest(r *http.Request, request *RequestStruct) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error read body")
		return err
	}
	err = json.Unmarshal(body, request)
	return err
}

//AddUser method is created to add users
func (env *Env) AddUser(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	var response ResponseStruct
	err := ReadRequest(r, &request)
	if err != nil {
		fmt.Println("error read request")
		return
	}
	result, err := env.db.Exec("insert into users_info (first_name, last_name, reg_date) values(?,?,now())",
		request.Data.User.FirstName, request.Data.User.LastName)
	if err != nil {
		fmt.Println("error add user to database")
		return
	}
	response.Method = request.Method
	response.Status = 0
	response.Data.Message = "User have benn created successfully"
	response.Data.IDUser, _ = result.LastInsertId()
	responseJSON, _ := json.Marshal(response)
	result, err = env.db.Exec("insert into users_balance values(?,0)",
		response.Data.IDUser)
	if err != nil {
		fmt.Println("error set user's balance to database")
		return
	}
	w.Write(responseJSON)
}

//Replenishment asd
func (env *Env) Replenishment(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	var response ResponseStruct
	err := ReadRequest(r, &request)
	if err != nil {
		fmt.Println("error read request")
		return
	}
	result, err := env.db.Exec("insert into transuctions (user_id, tr_value, info, tr_date) values(?,?,?,now())",
		request.Data.ReplenishmentData.ID, request.Data.ReplenishmentData.Value, "replenishment")
	if err != nil {
		fmt.Println("error add trunsuction to database")
		return
	}
	response.Method = request.Method
	response.Status = 0
	response.Data.Message = "Transuction have been created successfully"
	response.Data.IDTrunsuction, _ = result.LastInsertId()
	responseJSON, _ := json.Marshal(response)
	result, err = env.db.Exec("update users_balance set balance = balance + ? where user_id = ?",
		request.Data.ReplenishmentData.Value, request.Data.ReplenishmentData.ID)
	if err != nil {
		fmt.Println("error update user balance")
		return
	}
	w.Write(responseJSON)
}

func main() {
	db, err := sql.Open("mysql", "go_server:pass@/balance_service")
	if err != nil {
		fmt.Println("error open database")
		fmt.Println(err)
	}
	env := &Env{db: db}
	http.HandleFunc("/AddUser", env.AddUser)
	http.HandleFunc("/Replenishment", env.Replenishment)
	http.ListenAndServe(":8080", nil)
}
