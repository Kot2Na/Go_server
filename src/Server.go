package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// Transaction asd
type Transaction struct {
	Info  string  `json:"info"`
	Value float64 `json:"value"`
	Date  string  `json:"date"`
}

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
	IDUser    int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// RequestData asd
type RequestData struct {
	ReplenishmentData ReplenishmentStruct `json:"replenishment"`
	WithdrawalData    WithdravalStruct    `json:"withdrawal"`
	TransferData      TransferStruct      `json:"transfer"`
	User              UserStruct          `json:"user"`
	Currency          string              `json:"currency"`
	Sort              string              `json:"sort"`
}

//RequestStruct is data struct, that will contain json request fields
type RequestStruct struct {
	Method string      `json:"method"`
	Data   RequestData `json:"data"`
}

//ResponseData asd
type ResponseData struct {
	Message           string
	IDUser            int64
	IDTrunsuction     int64
	UserBalance       float64
	UsersTransactions []Transaction `json:"userstransactions"`
}

//ResponseStruct asd
type ResponseStruct struct {
	Method string
	Data   ResponseData
}

//SendResponse asd
func SendResponse(response *ResponseStruct, w *http.ResponseWriter) {
	responseJSON, _ := json.Marshal(*response)
	(*w).Write(responseJSON)
}

//ReadRequest asd
func ReadRequest(r *http.Request, request *RequestStruct) error {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, request)
	if err != nil {
		return err
	}
	return nil
}

// CheckUser asd
func (env *Env) CheckUser(userid int64) error {
	var flag int
	result := env.db.QueryRow("select exists(select user_id from users_info where user_id = ?) as user_id", userid)
	err := result.Scan(&flag)
	if err != nil {
		return fmt.Errorf("Database: error")
	}
	if flag == 0 {
		return fmt.Errorf("Request: user is not exist")
	}
	return nil
}

// CheckBalance asd
func (env *Env) CheckBalance(userid int64, sum float64) error {
	var balance float64
	result := env.db.QueryRow("select balance from users_balance where user_id = ?", userid)
	err := result.Scan(&balance)
	if err != nil {
		return err
	}
	if sum > balance {
		return fmt.Errorf("Not enough funds on balance")
	}
	return nil
}

// ChangeBalance asd
func (env *Env) ChangeBalance(userid int64, sum float64) error {
	if _, err := env.db.Exec("update users_balance set balance = balance + ? where user_id = ?",
		sum, userid); err != nil {
		return err
	}
	return nil
}

// GetTransactionList asd
func (env *Env) GetTransactionList(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	var response ResponseStruct
	var sort string
	if err := ReadRequest(r, &request); err != nil {
		http.Error(w, err.Error(), 300)
		return
	}
	if err := env.CheckUser(request.Data.User.IDUser); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if request.Data.Sort == "time" {
		sort = "tr_date"
	} else {
		sort = "ABS(tr_value)"
	}
	qtext := fmt.Sprintf("select info, tr_value, tr_date from transactions where user_id = %d order by %s desc", request.Data.User.IDUser, sort)
	rows, err := env.db.Query(qtext)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for rows.Next() {
		var Tr Transaction
		if err2 := rows.Scan(&Tr.Info, &Tr.Value, &Tr.Date); err2 != nil {
			http.Error(w, err2.Error(), 500)
			return
		}
		response.Data.UsersTransactions = append(response.Data.UsersTransactions, Tr)
	}
	response.Data.Message = "List have been created"
	response.Method = request.Method
	SendResponse(&response, &w)
}

//AddUser method is created to add users
func (env *Env) AddUser(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	var response ResponseStruct
	if err := ReadRequest(r, &request); err != nil {
		http.Error(w, err.Error(), 300)
		return
	}
	result, err := env.db.Exec("insert into users_info (first_name, last_name, reg_date) values(?,?,now())",
		request.Data.User.FirstName, request.Data.User.LastName)
	if err != nil {
		http.Error(w, "Error add user to database", 500)
		return
	}
	response.Data.IDUser, _ = result.LastInsertId()
	if _, err := env.db.Exec("insert into users_balance values(?,0)",
		response.Data.IDUser); err != nil {
		http.Error(w, "Error set user's balance to database", 500)
		return
	}
	response.Data.Message = "User have been created successfully"
	response.Method = request.Method
	SendResponse(&response, &w)
}

//Balance method returns the value of user's balance
func (env *Env) Balance(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	var response ResponseStruct
	var balance float64
	if err := ReadRequest(r, &request); err != nil {
		http.Error(w, err.Error(), 300)
		return
	}
	if err := env.CheckUser(request.Data.User.IDUser); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	result := env.db.QueryRow("select round(balance, 3) from users_balance where user_id = ?",
		request.Data.User.IDUser)
	if err := result.Scan(&balance); err != nil {
		http.Error(w, err.Error(), 300)
		return
	}
	if request.Data.Currency != "RUB" {
		var result map[string]map[string]float64
		r, err := http.Get("https://api.exchangeratesapi.io/latest?base=RUB")
		if err != nil {
			http.Error(w, err.Error(), 300)
			return
		}
		byteValue, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal([]byte(byteValue), &result)
		balance = math.Round(balance*result["rates"][request.Data.Currency]*1000) / 1000
	}
	response.Data.IDUser = request.Data.User.IDUser
	response.Data.UserBalance = balance
	response.Data.Message = "User's balance"
	response.Method = request.Method
	SendResponse(&response, &w)
}

//Replenishment asd
func (env *Env) Replenishment(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	var response ResponseStruct
	if err := ReadRequest(r, &request); err != nil {
		http.Error(w, err.Error(), 300)
		return
	}
	if err := env.CheckUser(request.Data.ReplenishmentData.ID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	result, err := env.db.Exec("insert into transactions (user_id, tr_value, info, tr_date) values(?,?,?,now())",
		request.Data.ReplenishmentData.ID, request.Data.ReplenishmentData.Value, "replenishment")
	if err != nil {
		http.Error(w, "Error add transaction to database", 500)
		return
	}
	if err = env.ChangeBalance(request.Data.ReplenishmentData.ID, request.Data.ReplenishmentData.Value); err != nil {
		http.Error(w, "Error update user balance", 500)
		return
	}
	response.Method = request.Method
	response.Data.Message = "transaction have been created successfully"
	response.Data.IDTrunsuction, _ = result.LastInsertId()
	SendResponse(&response, &w)
}

//Withdrawal asd
func (env *Env) Withdrawal(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	var response ResponseStruct
	if err := ReadRequest(r, &request); err != nil {
		http.Error(w, err.Error(), 300)
		return
	}
	if err := env.CheckUser(request.Data.WithdrawalData.ID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := env.CheckBalance(request.Data.WithdrawalData.ID, request.Data.WithdrawalData.Value); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	result, err := env.db.Exec("insert into transactions (user_id, tr_value, info, tr_date) values(?,?,?,now())",
		request.Data.WithdrawalData.ID, -request.Data.WithdrawalData.Value, "withdrawal")
	if err != nil {
		http.Error(w, "Error add transaction to database", 500)
		return
	}
	if err = env.ChangeBalance(request.Data.WithdrawalData.ID, -request.Data.WithdrawalData.Value); err != nil {
		http.Error(w, "Error update user balance", 500)
		return
	}
	response.Method = request.Method
	response.Data.Message = "transaction have been created successfully"
	response.Data.IDTrunsuction, _ = result.LastInsertId()
	SendResponse(&response, &w)
}

//Transfer asd
func (env *Env) Transfer(w http.ResponseWriter, r *http.Request) {
	var request RequestStruct
	var response ResponseStruct
	if err := ReadRequest(r, &request); err != nil {
		http.Error(w, err.Error(), 300)
		return
	}
	if err := env.CheckUser(request.Data.TransferData.IDFrom); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := env.CheckUser(request.Data.TransferData.IDTo); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := env.CheckBalance(request.Data.TransferData.IDFrom, request.Data.TransferData.Value); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	result, err := env.db.Exec("insert into transactions (user_id, tr_value, info, tr_date) values(?,?,?,now())",
		request.Data.TransferData.IDFrom, -request.Data.TransferData.Value, "transfer")
	if err != nil {
		http.Error(w, "Error add transaction to database", 500)
		return
	}
	response.Data.IDTrunsuction, _ = result.LastInsertId()
	_, err = env.db.Exec("insert into transactions values(?,?,?,?,now())",
		response.Data.IDTrunsuction, request.Data.TransferData.IDTo, request.Data.TransferData.Value, "transfer")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err = env.ChangeBalance(request.Data.TransferData.IDFrom, -request.Data.TransferData.Value); err != nil {
		http.Error(w, "Error update user balance", 500)
		return
	}
	if err = env.ChangeBalance(request.Data.TransferData.IDTo, request.Data.TransferData.Value); err != nil {
		http.Error(w, "Error update user balance", 500)
		return
	}
	response.Method = request.Method
	response.Data.Message = "Transaction have been created successfully"
	SendResponse(&response, &w)
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
	http.HandleFunc("/Withdrawal", env.Withdrawal)
	http.HandleFunc("/Transfer", env.Transfer)
	http.HandleFunc("/Balance", env.Balance)
	http.HandleFunc("/Transactions", env.GetTransactionList)
	http.ListenAndServe(":8080", nil)
}
