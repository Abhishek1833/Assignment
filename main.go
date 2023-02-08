package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Response struct {
	Status     string      `json:"status"`
	StatucCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}
type Transaction struct {
	Amount float64   `json:"amount"`
	Time   time.Time `json:"timestamp"`
	UserID int       `json:"userId"`
}
type Statistics struct {
	Sum   float64 `json:"sum"`
	Avg   float64 `json:"avg"`
	Max   float64 `json:"max"`
	Min   float64 `json:"min"`
	Count int     `json:"count"`
}
type UserLocation struct {
	City string `json:"city"`
}

var Userloc = UserLocation{}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/transactions", addTransaction).Methods("POST")
	r.HandleFunc("/statistics", getstatistics).Methods("GET")
	r.HandleFunc("/transactions", deleteTransactions).Methods("DELETE")
	r.HandleFunc("/location", setLocation).Methods("POST")
	r.HandleFunc("/location", resetlocation).Methods("PUT")
	http.ListenAndServe(":4000", r)
}

var transactions []Transaction

func addTransaction(w http.ResponseWriter, r *http.Request) {
	var t Transaction
	w.Header().Set("Content-Type", "application/json")
	if r.Body == nil {
		//json.NewEncoder(w).Encode(&Response{"failed", 400, "Please send a request body"})
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		//json.NewEncoder(w).Encode(&Response{"failed", 400, err.Error()})
		http.Error(w, err.Error(), 400)
		return
	}
	userID := r.Header.Get("user_id")
	a, _ := strconv.ParseInt(userID, 10, 32)
	t.UserID = int(a)
	timestamp := time.Now().UTC()
	if t.Time.After(timestamp) {
		http.Error(w, "Transaction timestamp cannot be in the future", 422)
		//json.NewEncoder(w).Encode(&Response{"failed", 422, "Transaction timestamp cannot be in the future"})
		return
	}
	if timestamp.Sub(t.Time).Seconds() > 60 {
		w.WriteHeader(204)
	} else {
		w.WriteHeader(201)
	}
	tmp := Transaction{}
	tmp.Amount = t.Amount
	tmp.Time = t.Time
	tmp.UserID = t.UserID
	transactions = append(transactions, tmp)
	return
	//jsonResp, err := json.Marshal(tmp)
	//w.Write(jsonResp)
	//w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(tmp)
}

func getstatistics(w http.ResponseWriter, r *http.Request) {
	var sum, avg, max, min float64
	if Userloc.City != "" {
		loc := r.URL.Query().Get("location")
		if loc != Userloc.City {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			resp := make(map[string]string)
			resp["message"] = "Unauthorized"
			json.NewEncoder(w).Encode(resp)
			return
		}

	}
	var count int
	now := time.Now().UTC()
	for _, t := range transactions {
		if now.Sub(t.Time).Seconds() <= 60 {
			sum += t.Amount
			count++
			if t.Amount > max {
				max = t.Amount
			}
			if t.Amount < min || min == 0 {
				min = t.Amount
			}
		}
	}
	if count > 0 {
		avg = sum / float64(count)
	}
	statistics := Statistics{
		Sum:   sum,
		Avg:   avg,
		Max:   max,
		Min:   min,
		Count: count,
	}
	json.NewEncoder(w).Encode(statistics)
}

func deleteTransactions(w http.ResponseWriter, r *http.Request) {
	transactions = make([]Transaction, 0)
	w.WriteHeader(204)
	return
}

func setLocation(w http.ResponseWriter, r *http.Request) {
	json.NewDecoder(r.Body).Decode(&Userloc)
	json.NewEncoder(w).Encode(Userloc)
	return
}
func resetlocation(w http.ResponseWriter, r *http.Request) {
	Userloc.City = ""
	json.NewEncoder(w).Encode(Userloc)
	return
}
