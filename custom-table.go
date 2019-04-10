package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/xid"
)

type resResultT struct {
	Code   string      `json:"code"`
	Des    string      `json:"des"`
	Result interface{} `json:"result"`
}

type rawColumn struct {
	Name     string `db:"name" json:"name"`
	Value    string `db:"value" json:"value"`
	MoudleID string `db:"module_id" json:"moduleId"`
	Fixed    string `db:"fixed" json:"fixed"`
	ID       string `db:"id" json:"id"`
	Location string `db:"location" json:"location"`
	Rule     string `db:"rule" json:"rule"`
	Status   string `json:"status"`
}

type userColumn struct {
	Name     string `db:"name" json:"name"`
	Value    string `db:"value" json:"value"`
	Fixed    string `db:"fixed" json:"fixed"`
	Hidden   string `db:"hidden" json:"hidden"`
	Frozen   string `db:"frozen" json:"frozen"`
	ID       string `db:"id" json:"id"`
	Location string `db:"location" json:"location"`
	Rule     string `db:"rule" json:"rule"`
	Status   string `json:"status" json:"-"`
	Seq      string `db:"seq" json:"-"`
	Width    string `db:"width" json:"width"`
}

// GetMaintenanceTable 运维表数据
func GetMaintenanceTable(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	switch r.Method {
	case http.MethodGet:
		getMaintenanceTable(w, r)
	case http.MethodPost:
		updateMaintenanceTable(w, r)
	}
}

func getMaintenanceTable(w http.ResponseWriter, r *http.Request) {

	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		log.Fatalln(err)
	}

	ids := r.URL.Query()["id"]
	id := ""
	if len(ids) >= 1 {
		id = ids[0]
	}

	columns := []rawColumn{}
	err = db.Select(&columns, "select id, name, value, fixed, location, rule from `raw_column` where module_id = $1 order by seq", id)
	if err != nil {
		panic(err)
	}

	db.Close()

	res := resResultT{
		Code:   "0",
		Des:    "",
		Result: columns,
	}
	json.NewEncoder(w).Encode(res)
}

func updateMaintenanceTable(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	columns := []rawColumn{}
	err := decoder.Decode(&columns)
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

	for i, datum := range columns {
		if datum.Status == "0" {
			tx.MustExec("insert into raw_column(name, value, module_id, fixed, location, seq, id, rule) values($1,$2,$3,$4,$5, $6, $7, $8)", datum.Name, datum.Value, id, datum.Fixed, datum.Location, i, xid.New().String(), datum.Rule)
		} else if datum.Status == "1" {
			tx.MustExec("update `raw_column` set name=$1, value=$2, location=$3, fixed=$4, seq=$5, rule=$6 where module_id=$7 and id=$8", datum.Name, datum.Value, datum.Location, datum.Fixed, i, datum.Rule, id, datum.ID)
		} else if datum.Status == "2" {
			tx.MustExec("delete from `raw_column` where value=$1", datum.Value)
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

type sliceString []string

func (slice sliceString) search(str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// GetUserMaintenanceTable 用户表数据
func GetUserMaintenanceTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	switch r.Method {
	case http.MethodGet:
		getUserMaintenanceTable(w, r)
	case http.MethodPost:
		updateUserMaintenanceTable(w, r)
	}
}

func updateUserMaintenanceTable(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	columns := []userColumn{}
	err := decoder.Decode(&columns)

	// fmt.Println(columns)

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
	err = db.Select(&allUserValues, "select value from user_column where user_id=$1", token)
	if err != nil {
		panic(err)
	}

	for i, datum := range columns {

		if datum.Status == "1" {
			if allUserValues.search(datum.Value) {
				tx.MustExec("update user_column set hidden=$1, frozen=$2, seq=$3, location=$4, rule=$5 where module_id=$6 and user_id=$7 and value=$8", datum.Hidden, datum.Frozen, i, datum.Location, datum.Rule, id, token, datum.Value)
			} else {
				tx.MustExec("insert into user_column(module_id, user_id, value, hidden, frozen, seq, location, id, rule) values($1,$2,$3,$4,$5, $6, $7, $8, $9)", id, token, datum.Value, datum.Hidden, datum.Frozen, i, datum.Location, xid.New().String(), datum.Rule)
			}
		} else {
			// tx.MustExec("update user_column set seq=$1 where module_id=$2 and user_id=$3 and value=$4", i, id, token, datum.Value)
		}
		// } else {
		// tx.MustExec("insert into user_column(module_id, user_id, value, hidden, frozen, seq, location, id, rule) values($1,$2,$3,$4,$5, $6, $7, $8, $9)", id, token, datum.Value, datum.Hidden, datum.Frozen, i, datum.Location, xid.New().String(), datum.Rule)
		// }
	}

	tx.Commit()
	db.Close()

	res := resResultT{
		Code: "0",
		Des:  "",
	}
	json.NewEncoder(w).Encode(res)
}

func getUserMaintenanceTable(w http.ResponseWriter, r *http.Request) {
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

	columns := []userColumn{}
	err = db.Select(&columns, `
						select name, fixed, raw_column.value, 
						case when width is null
						then '' else width
						end as width,
						case when user_column.seq is null
                        then raw_column.seq else user_column.seq
                        end as seq,
                        case when hidden is null
                        then '0' else hidden
                        end as hidden,
                        case when frozen is null
                        then '' else frozen
                        end as frozen, 
                        case when user_column.id is null
                        then raw_column.id else user_column.id
                        end as id,
						case when user_column.location is null
						then raw_column.location  else user_column.location
						end as location,
						case when user_column.rule is null
						then raw_column.rule else user_column.rule
						end as rule
                        from raw_column left join (
							select hidden, frozen, location, id, rule, width,
							user_id, value, module_id, seq from user_column where user_id=$1
                        ) as user_column
                        on raw_column.module_id=user_column.module_id 
                        and raw_column.value=user_column.value 
						where raw_column.module_id=$2
						order by seq asc
						`, token, id)

	if err != nil {
		panic(err)
	}

	db.Close()

	res := resResultT{
		Code:   "0",
		Des:    "",
		Result: columns,
	}
	json.NewEncoder(w).Encode(res)
}

// ResetUserMaintenanceTable 重置用户表数据
func ResetUserMaintenanceTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

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
	tx.MustExec("delete from user_column where user_id=$1 and module_id=$2", token, id)
	tx.Commit()
	db.Close()
	res := resResultT{
		Code: "0",
		Des:  "",
	}
	json.NewEncoder(w).Encode(res)
}

// OverrideUserMaintenanceTable 重置用户表数据
func OverrideUserMaintenanceTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	
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
	columns := make(map[string][]string)
	err = decoder.Decode(&columns)
	if err != nil {
		panic(err)
	}

	allUserTableFields := sliceString{
		"hidden", "frozen", "location", "width",
	}

	tx := db.MustBegin()
	for k, v := range columns {
		sqlStr := "update user_column set "
		for _, filed := range v {
			if allUserTableFields.search(filed) {
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

//UpdateUserMaintenanceTableWidth 设置表格宽度
func UpdateUserMaintenanceTableWidth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		log.Fatalln(err)
	}

	token := r.Header.Get("token")

	if token == "" {
		panic("token is nil")
	}

	decoder := json.NewDecoder(r.Body)
	param := make(map[string]string)
	err = decoder.Decode(&param)
	if err != nil {
		panic(err)
	}

	width := param["width"]
	id := param["id"]
	value := param["value"]

	res := resResultT{
		Code: "0",
		Des:  "",
	}

	if width == "" || value == "" || id == "" {
		json.NewEncoder(w).Encode(res)
		return
	}

	tx := db.MustBegin()
	var dataExist int
	err = db.Get(&dataExist, "select count(*) from user_column where module_id=$1 and user_id=$2 and value=$3", id, token, value)
	if err != nil {
		panic(err)
	}
	if dataExist == 0 {
		tx.MustExec("insert into user_column(module_id, user_id, value, width, id) values($1,$2,$3,$4,$5)", id, token, value, width, xid.New().String())
	} else {
		tx.MustExec("update user_column set width=$1 where module_id=$2 and user_id=$3 and value=$4", width, id, token, value)
	}

	tx.Commit()
	db.Close()

	json.NewEncoder(w).Encode(res)
}
