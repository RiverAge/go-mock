package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/xid"
)

type rawFilter struct {
	Name     string `db:"name" json:"name"`
	Value    string `db:"value" json:"value"`
	MoudleID string `db:"module_id" json:"moduleId"`
	Fixed    string `db:"fixed" json:"fixed"`
	ID       string `db:"id" json:"id"`
	Status   string `json:"status"`
}

type userFilter struct {
	Name   string `db:"name" json:"name"`
	Value  string `db:"value" json:"value"`
	Fixed  string `db:"fixed" json:"fixed"`
	Hidden string `db:"hidden" json:"hidden"`
	ID     string `db:"id" json:"id"`
	Status string `json:"status" json:"-"`
	Seq    string `db:"seq" json:"-"`
}

// GetMaintenanceFilter 运维表数据
func GetMaintenanceFilter(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		getMaintenanceFilter(w, r)
	case http.MethodPost:
		updateMaintenanceFilter(w, r)
	}
}

func getMaintenanceFilter(w http.ResponseWriter, r *http.Request) {

	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		log.Fatalln(err)
	}

	ids := r.URL.Query()["id"]
	id := ""
	if len(ids) >= 1 {
		id = ids[0]
	}

	filters := []rawFilter{}
	err = db.Select(&filters, "select id, name, value, fixed from `raw_filter` where module_id = $1 order by seq", id)
	if err != nil {
		panic(err)
	}

	db.Close()

	res := resResultT{
		Code:   "0",
		Des:    "",
		Result: filters,
	}
	json.NewEncoder(w).Encode(res)
}

func updateMaintenanceFilter(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	filters := []rawFilter{}
	err := decoder.Decode(&filters)
	if err != nil {
		panic(err)
	}

	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		panic(err)
	}

	ids := r.URL.Query()["id"]
	id := ""
	if len(ids) >= 1 {
		id = ids[0]
	}

	tx := db.MustBegin()

	for i, datum := range filters {
		if datum.Status == "0" {
			tx.MustExec("insert into raw_filter(name, value, module_id, fixed, seq, id) values($1,$2,$3,$4,$5, $6)", datum.Name, datum.Value, id, datum.Fixed, i, xid.New().String())
		} else if datum.Status == "1" {
			tx.MustExec("update `raw_filter` set name=$1, value=$2, fixed=$3, seq=$4 where module_id=$5 and id=$6", datum.Name, datum.Value, datum.Fixed, i, id, datum.ID)
		} else if datum.Status == "2" {
			tx.MustExec("delete from `raw_filter` where value=$1", datum.Value)
		}
	}
	tx.Commit()
	db.Close()

	res := resResultT{
		Code: "0",
		Des:  "",
	}
	json.NewEncoder(w).Encode(res)
}

// GetUserMaintenanceFilter 用户过滤数据
func GetUserMaintenanceFilter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUserMaintenanceFilter(w, r)
	case http.MethodPost:
		updateUserMaintenanceFilter(w, r)
	}
}

func updateUserMaintenanceFilter(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	filter := []userFilter{}
	err := decoder.Decode(&filter)

	if err != nil {
		panic(err)
	}

	ids := r.URL.Query()["id"]
	id := ""
	if len(ids) >= 1 {
		id = ids[0]
	}

	token := r.Header.Get("token")

	if id == "" || token == "" {
		panic("id or token can not be null")
	}

	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		panic(err)
	}

	tx := db.MustBegin()
	allUserValues := sliceString{}
	err = db.Select(&allUserValues, "select value from user_filter where user_id=$1", token)
	if err != nil {
		panic(err)
	}

	for i, datum := range filter {

		if datum.Status == "1" {
			if allUserValues.search(datum.Value) {
				tx.MustExec("update user_filter set hidden=$1, seq=$2 where module_id=$3 and user_id=$4 and value=$5", datum.Hidden, i, id, token, datum.Value)
			} else {
				tx.MustExec("insert into user_filter(module_id, user_id, value, hidden, seq, id) values($1,$2,$3,$4,$5, $6)", id, token, datum.Value, datum.Hidden, i, xid.New().String())
			}
		}
	}

	tx.Commit()
	db.Close()

	res := resResultT{
		Code: "0",
		Des:  "",
	}
	json.NewEncoder(w).Encode(res)
}

func getUserMaintenanceFilter(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		log.Fatalln(err)
	}

	ids := r.URL.Query()["id"]
	var id string
	if len(ids) >= 1 {
		id = ids[0]
	}

	token := r.Header.Get("token")

	if id == "" || token == "" {
		panic("id or token is nil")
	}

	filter := []userFilter{}
	err = db.Select(&filter, `
						select name, fixed, raw_filter.value,
						case when user_filter.seq is null
                        then raw_filter.seq else user_filter.seq
                        end as seq,
                        case when hidden is null
                        then '0' else hidden
                        end as hidden,
                        case when user_filter.id is null
                        then raw_filter.id else user_filter.id
                        end as id
                        from raw_filter left join (
							select hidden, id,
							user_id, value, module_id, seq from user_filter where user_id=$1
                        ) as user_filter
                        on raw_filter.module_id=user_filter.module_id 
                        and raw_filter.value=user_filter.value 
						where raw_filter.module_id=$2
						order by seq asc
						`, token, id)

	if err != nil {
		panic(err)
	}

	db.Close()

	res := resResultT{
		Code:   "0",
		Des:    "",
		Result: filter,
	}
	json.NewEncoder(w).Encode(res)
}

// ResetUserMaintenanceFilter 重置用户过滤数据
func ResetUserMaintenanceFilter(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		log.Fatalln(err)
	}

	ids := r.URL.Query()["id"]
	var id string
	if len(ids) >= 1 {
		id = ids[0]
	}

	token := r.Header.Get("token")

	if id == "" || token == "" {
		panic("id or token is nil")
	}

	tx := db.MustBegin()
	tx.MustExec("delete from user_filter where user_id=$1 and module_id=$2", token, id)
	tx.Commit()
	db.Close()
	res := resResultT{
		Code: "0",
		Des:  "",
	}
	json.NewEncoder(w).Encode(res)
}

// OverrideUserMaintenanceFilter 重置用户过滤数据
func OverrideUserMaintenanceFilter(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		log.Fatalln(err)
	}

	ids := r.URL.Query()["id"]
	var id string
	if len(ids) >= 1 {
		id = ids[0]
	}

	if id == "" {
		panic("id can not be null")
	}

	decoder := json.NewDecoder(r.Body)
	filter := make(map[string][]string)
	err = decoder.Decode(&filter)
	if err != nil {
		panic(err)
	}

	allUserFilterFields := sliceString{
		"hidden",
	}

	tx := db.MustBegin()
	for k, v := range filter {
		sqlStr := "update user_filter set "
		for _, filed := range v {
			if allUserFilterFields.search(filed) {
				sqlStr = sqlStr + filed + "=null, "
			}
		}
		if strings.HasSuffix(sqlStr, ", ") {
			newSQLStr := strings.TrimSuffix(sqlStr, ", ")
			newSQLStr += " where value=$1 and module_id=$2"
			tx.MustExec(newSQLStr, k, id)
		}
	}
	tx.Commit()

	db.Close()
	res := resResultT{
		Code: "0",
		Des:  "",
	}
	json.NewEncoder(w).Encode(res)
}
