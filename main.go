package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mozillazg/go-pinyin"
	"github.com/rs/xid"

	_ "github.com/mattn/go-sqlite3"
)

var schema = `
CREATE TABLE IF NOT EXISTS raw_column (
    value     VARCHAR (255),
    module_id VARCHAR (255),
    fixed     BOOLEAN,
    name      VARCHAR (255),
    location  CHAR (2),
    PRIMARY KEY (
        value,
        module_id
    )
);

CREATE TABLE IF NOT EXISTS column (
	name VARCHAR(80),
	value VARCHAR(80),
	width INTEGER,
	location VARCHAR(80),
	fixed VARCHAR(10),
	hidden VARCHAR(10),
	frozen VARCHAR(10),
	` + "`" + `order` + "`" + `INTEGER
);

CREATE TABLE  IF NOT EXISTS person (
	id INTEGER,
	first_name VARCHAR(80),
	last_name VARCHAR(80),
	email VARCHAR(80),
	gender  VARCHAR(80), 
	ip_address VARCHAR(80),
	
	city VARCHAR(80),
	country VARCHAR(80),
	latitude VARCHAR(80),
	longitude VARCHAR(80),
	guid VARCHAR(80)
);
`

type Column struct {
	Name     string  `db:"name" json:"name"`
	Value    string  `db:"value" json:"value"`
	Width    float32 `db:"width" json:"width"`
	Location string  `db:"location" json:"location"`
	Order    string  `db:"order" json:"order"`
	Fixed    string  `db:"fixed" json:"fixed"`
	Hidden   string  `db:"hidden" json:"hidden"`
	Frozen   string  `db:"frozen" json:"frozen"`
}

type Person struct {
	ID        string `db:"id" json:"id"`
	FirstName string `db:"first_name" json:"firstName"`
	LastName  string `db:"last_name" json:"lastName"`
	Email     string `db:"email" json:"email"`
	Gender    string `db:"gender" json:"gender"`
	IPAddress string `db:"ip_address" json:"IPaddr"`

	City      string `db:"city" json:"city"`
	Country   string `db:"country" json:"country"`
	Latitude  string `db:"latitude" json:"latitude"`
	Longitude string `db:"longitude" json:"longitude"`
	GUID      string `db:"guid" json:"guid"`
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", index)
	router.HandleFunc("/market/list", market)
	router.HandleFunc("/market/freight", freightIndex)
	router.HandleFunc("/market/private", privateIndex)
	router.HandleFunc("/market/container", containerIndex)
	router.HandleFunc("/market/goods", goodsIndex)
	router.HandleFunc("/market/product", productIndex)
	router.HandleFunc("/market/country", countryIndex)
	router.HandleFunc("/market/detail", detailIndex)
	router.HandleFunc("/market/period", periodIndex)
	router.HandleFunc("/market/del", delIndex)
	router.HandleFunc("/market/save", saveIndex)
	router.HandleFunc("/market/private/container/site", siteIndex)
	router.HandleFunc("/market/customer/search", search)
	router.HandleFunc("/market/customer/summary", summary)
	router.HandleFunc("/market_hgx/hgxForklift!queryForkliftSelect.dhtml", fleetLst)
	router.HandleFunc("/market_hgx/hgxForklift!queryForkliftList.dhtml", containerIndex)
	router.HandleFunc("/market_hgx/hgxForklift!insOrUpdForklift.do", setFleet)

	router.HandleFunc("/market/out-application/list", outApplicationList)
	router.HandleFunc("/market/out-application/apply", outApplicationApply)
	router.HandleFunc("/market/out-application/cancel", outApplicationCancel)

	router.HandleFunc("/market/plugin-application/list", pluginApplicationList)
	router.HandleFunc("/market/plugin-application/plugin/apply", outApplicationApply)
	router.HandleFunc("/market/plugin-application/plugout/apply", outApplicationApply)
	router.HandleFunc("/market/plugin-application/cancel", outApplicationCancel)

	router.HandleFunc("/market/settlement/list", marketSettlementList)
	router.HandleFunc("/market/settlement/detail", marketSettlementDetail)
	router.HandleFunc("/market/settlement/confirm", outApplicationApply)

	router.HandleFunc("/market/statistics/enter", enterStatistics)
	router.HandleFunc("/market/statistics/enter/detail", enterDetailStatistics)
	router.HandleFunc("/market/statistics/enter/product", enterProductStatistics)
	router.HandleFunc("/market/statistics/enter/customer", enterCustomerStatistics)

	router.HandleFunc("/market/lau/list", marketLauList)

	router.HandleFunc("/flutter/task/insert", flutterTaskInsert)

	router.HandleFunc("/flutter/new/version", newVersion)

	router.HandleFunc("/new/platform", newPlatform)

	router.HandleFunc("/login", login)
	router.HandleFunc("/permission", permission)

	router.HandleFunc("/data/person", dataPerson)
	router.HandleFunc("/data/column", dataColumn)
	router.HandleFunc("/data/column/update", dataColumnUpdate)
	router.HandleFunc("/data/column/width/update", dataColumnWidthUpdate)
	router.HandleFunc("/data/update/from/csv", updateFromCSV)
	router.HandleFunc("/data/upload_file", uploadFile)

	router.HandleFunc("/data/test_query_string", testQueryString)

	router.HandleFunc("/api/ff-admin/v1/employee/getEno", getEno)
	router.HandleFunc("/api/ff-flatcar/v1/boardInfo/queryBoardInfoList/search", plateSearch)

	router.HandleFunc("/drop-down/ds", dropDownDS)
	router.HandleFunc("/cascade/ds", cascadeDS)

	router.HandleFunc("/custom-table/maintenance/table", GetMaintenanceTable)
	router.HandleFunc("/custom-table/user/maintenance/table", GetUserMaintenanceTable)
	router.HandleFunc("/custom-table/user/maintenance/table/width", UpdateUserMaintenanceTableWidth)
	router.HandleFunc("/custom-table/user/maintenance/reset", ResetUserMaintenanceTable)
	router.HandleFunc("/custom-table/maintenance/table/overrie-columns", OverrideUserMaintenanceTable)
	router.HandleFunc("/custom-table/maintenance/filter", GetMaintenanceFilter)
	router.HandleFunc("/custom-table/user/maintenance/filter", GetUserMaintenanceFilter)
	router.HandleFunc("/custom-table/user/maintenance/filter/reset", ResetUserMaintenanceFilter)
	router.HandleFunc("/custom-table/maintenance/filter/overrie-columns", OverrideUserMaintenanceFilter)

	log.Fatal(http.ListenAndServe(":8088", router))
}

type codeRetT struct {
	Code   string      `json:"code"`
	Result interface{} `json:"result"`
	Des    string      `json:"des"`
}

func cascadeDS(w http.ResponseWriter, r *http.Request) {
	plan, _ := ioutil.ReadFile("account.json")
	var data interface{}
	err := json.Unmarshal(plan, &data)
	if err != nil {
		// fmt.Println(err)
		panic(err)
	}

	// fmt.Println(data)

	json.NewEncoder(w).Encode(data)
}

func dropDownDS(w http.ResponseWriter, r *http.Request) {
	plan, _ := ioutil.ReadFile("bank.json")
	var data interface{}
	err := json.Unmarshal(plan, &data)
	if err != nil {
		// fmt.Println(err)
		panic(err)
	}
	// fmt.Println(data)

	json.NewEncoder(w).Encode(data)
}

func plateSearch(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	
	plan, _ := ioutil.ReadFile("plate-search.json")
	var data interface{}
	err := json.Unmarshal(plan, &data)
	if err != nil {
		// fmt.Println(err)
		panic(err)
	}
	// fmt.Println(data)

	json.NewEncoder(w).Encode(data)

}

func getEno(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	json.NewEncoder(w).Encode(codeRetT{
		Code:   "0",
		Result: "0009",
		Des:    "fullfill the request!",
	})
}

func testQueryString(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	queryStr := r.URL.Query()["queryStr"][0]
	fmt.Println(queryStr)

	json.NewEncoder(w).Encode("0")
}

func updateFromCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		panic(err)
	}

	csvFile, err := os.Open("MOCK_DATA.csv")
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	tx := db.MustBegin()

	csvReader := csv.NewReader(csvFile)
	csvReader.Read()
	i := 1
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err) // or handle it another way
		} else {
			tx.MustExec("update `person` set city=$1, country=$2, latitude=$3 ,longitude=$4, guid=$5 where id=$6", row[0], row[1], row[2], row[3], row[4], i)
		}
		i++
		// use the `row` here
	}
	tx.Commit()

	db.Close()
}

func dataColumnWidthUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		panic(err)
	}

	column := r.URL.Query()["column"][0]
	width := r.URL.Query()["width"][0]

	tx := db.MustBegin()
	tx.MustExec("update `column` set width =$1 where value=$2", width, column)
	tx.Commit()
	db.Close()
	json.NewEncoder(w).Encode(codeRetT{
		Code: "0",
	})
}

func dataColumnUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		panic(err)
	}

	column := r.URL.Query()["column"][0]

	var data []Column
	err = json.Unmarshal([]byte(column), &data)
	if err != nil {
		panic(err)
	}

	tx := db.MustBegin()
	for i, datum := range data {
		tx.MustExec("update `column` set location = $1, `order`=$2, hidden = $3, frozen = $4 where value=$5", datum.Location, i, datum.Hidden, datum.Frozen, datum.Value)
	}
	tx.Commit()

	// fmt.Println(data)

	ret := codeRetT{
		Code: "0",
		Des:  "Request has been fullfilled!",
	}

	db.Close()

	json.NewEncoder(w).Encode(ret)
}

func dataColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(schema)

	columns := []Column{}
	db.Select(&columns, "select * from `column` order by `order`")

	db.Close()

	ret := codeRetT{
		Code:   "0",
		Des:    "Request has been fullfilled!",
		Result: columns,
	}

	json.NewEncoder(w).Encode(ret)

}

func dataPerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	db, err := sqlx.Connect("sqlite3", "_db.db")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(schema)

	people := []Person{}
	currentPages := r.URL.Query()["currentPage"]
	pageSizes := r.URL.Query()["pageSize"]
	currentPage := 1
	pageSize := 10
	if len(currentPages) != 0 {
		currentPage, err = strconv.Atoi(currentPages[0])
	}
	if len(pageSizes) != 0 {
		pageSize, err = strconv.Atoi(pageSizes[0])
	}

	startOffset := (currentPage - 1) * pageSize
	db.Select(&people, "select * from person limit $1,$2", startOffset, pageSize)
	count := make([]int, 2)
	err = db.Select(&count, "select count(*) from person")

	db.Close()

	type PaginationT struct {
		PageSize    int      `json:"pageSize"`
		CurrentPage int      `json:"currentPage"`
		Total       int      `json:"total"`
		Content     []Person `json:"content"`
	}

	pa := PaginationT{
		PageSize:    pageSize,
		CurrentPage: currentPage,
		Total:       count[2],
		Content:     people,
	}

	ret := codeRetT{
		Code:   "0",
		Des:    "Request has been fullfilled!",
		Result: pa,
	}

	json.NewEncoder(w).Encode(ret)

}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	username := r.URL.Query()["username"][0]
	password := r.URL.Query()["password"][0]

	var code, des string

	if username == "username" && password == "password" {
		code = "0"
		des = "登录成功"
	} else {
		code = "1"
		des = "用户名或密码错误"
	}

	ret := codeRetT{
		Code:   code,
		Result: "json web token",
		Des:    des,
	}

	time.Sleep(1 * time.Second)
	json.NewEncoder(w).Encode(ret)
}

type permissionT struct {
	Name    string        `json:"name"`
	Path    string        `json:"path"`
	Icon    string        `json:"icon"`
	SubMenu []permissionT `json:"subMenu"`
}

func permission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	plan, _ := ioutil.ReadFile("ds.json")
	var data []permissionT

	err := json.Unmarshal(plan, &data)

	ret := codeRetT{
		Code:   "0",
		Des:    "response success",
		Result: data,
	}
	if err == nil {
		time.Sleep(2 * time.Second)
		json.NewEncoder(w).Encode(ret)
	} else {
		fmt.Println(err)
	}

}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

type cell struct {
	GID                     string `json:"gId"`
	BlNo                    string `json:"blNo"`
	AgentCompany            string `json:"companyname"`
	PlateNo                 string `json:"plateNo"`
	ForecastUserCompany     string `json:"forecastUserCompany"`
	ForecastUserCompanyRole string `json:"forecastUserCompanyRole"`
	GoodsSource             string `json:"goodsSourceName"`
	GoodsSourceCode         string `json:"goodsSourceCode"`
	CreaterRole             string `json:"createrRole"`
	ForecastDate            string `json:"forecastDate"`
	Product                 string `json:"product"`
	ForecaseEnterDate       string `json:"forecastEnterDate"`
	ContainerNo             string `json:"containerNo"`
	FrameNo                 string `json:"frameNo"`
	DischargeStatus         string `json:"dischargeStatus"`
	IsPublicSite            string `json:"isPublicSite"`
	DropCabinetPosition     string `json:"privateSiteName"`
	GrossWeight             string `json:"sumGrossWeight"`
	PalletNumber            string `json:"sumPallet"`
	ContainerSizeName       string `json:"containerSizeName"`
	PrivateSiteID           string `json:"privateSiteId"`
	ContainerTypeID         string `json:"containerTypeId"`
	ForecastUserName        string `json:"forecastUserName"`
	ForecastConfirmDate     string `json:"forecastConfirmDate"`
	ActualEnterDate         string `json:"actualEnterDate"`
	AllowStatus             bool   `json:"allowStatus"`
	DisallowStatus          bool   `json:"disallowStatus"`
	CancelAllowedStatus     bool   `json:"cancelAllowedStatus"`
	CancelDisallowedStatus  bool   `json:"cancelDisallowedStatus"`
}

type freight struct {
	ID          string `json:"id"`
	Companyname string `json:"companyname"`
	Pinyin      string `json:"pinyin"`
}

type resData struct {
	PageSize    string `json:"pageSize"`
	CurrentPage string `json:"currentPage"`
	Content     []cell `json:"content"`
}

type resFreightRet struct {
	Result bool      `json:"result"`
	Msg    string    `json:"msg"`
	Data   []freight `json:"data"`
}

type resRet struct {
	Result bool        `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

func market(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	var s []cell
	for i := 0; i < 10; i++ {
		s = append(s, cell{
			GID:                     "#!989XXDDF" + strconv.Itoa(i),
			BlNo:                    "OPX9089",
			AgentCompany:            "上海欧恒进出口贸易有限公司",
			ForecastUserCompany:     "上海欧恒进出口贸易预报有限公司",
			GoodsSourceCode:         "0",
			CreaterRole:             "1",
			ForecastDate:            "2018-1-1 12:00",
			PlateNo:                 "沪A89834E",
			GoodsSource:             "海运柜",
			PrivateSiteID:           "XXXXX:::::LLLLLL::::::::XXX",
			Product:                 "鲜蓝莓",
			DischargeStatus:         "0",
			ForecaseEnterDate:       "2018-6-30 晚上",
			ContainerNo:             "OOL9093XXF",
			FrameNo:                 "XFEFG33422",
			IsPublicSite:            "0",
			DropCabinetPosition:     "40#",
			GrossWeight:             "100KG",
			ContainerSizeName:       "40",
			PalletNumber:            "249",
			ForecastUserCompanyRole: "市场",
			ForecastConfirmDate:     "2019-1-1",
			ForecastUserName:        "limeng",
			AllowStatus:             true,
			DisallowStatus:          true,
			CancelAllowedStatus:     true,
			CancelDisallowedStatus:  true,
			ContainerTypeID:         strconv.Itoa(i % 2),
		})
	}

	resRet0 := resRet{
		Result: true,
		Msg:    "!!resRet!!",
		Data: resData{
			PageSize:    "10",
			CurrentPage: "1",
			Content:     s,
		},
	}

	fmt.Println(r.Form)
	json.NewEncoder(w).Encode(resRet0)

}

func freightIndex(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("c.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var freights []freight
	scanner := bufio.NewScanner(file)
	a := pinyin.NewArgs()

	for scanner.Scan() {
		var py string
		for _, p := range pinyin.Pinyin(scanner.Text(), a) {
			py = py + p[0] + " "
		}

		freights = append(freights, freight{
			ID:          xid.New().String(),
			Companyname: scanner.Text(),
			Pinyin:      py,
		})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	resRet0 := resFreightRet{
		Result: true,
		Msg:    "",
		Data:   freights,
	}

	json.NewEncoder(w).Encode(resRet0)

}

func privateIndex(w http.ResponseWriter, r *http.Request) {
	type dataT struct {
		ID   string `json:"id"`
		Site string `json:"site"`
	}

	rand.Seed(time.Now().UnixNano())

	var data []dataT

	for i := 0; i < 20; i++ {
		data = append(data, dataT{
			ID:   xid.New().String(),
			Site: string(rand.Intn(26)+65) + strconv.Itoa(i),
		})
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data:   data,
	}

	json.NewEncoder(w).Encode(response)
}

func containerIndex(w http.ResponseWriter, r *http.Request) {
	type dataT struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	rand.Seed(time.Now().UnixNano())

	var data []dataT

	for i := 0; i < 20; i++ {
		data = append(data, dataT{
			ID:   xid.New().String(),
			Name: strconv.Itoa(rand.Intn(90) + 10),
		})
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data:   data,
	}

	json.NewEncoder(w).Encode(response)
}

func goodsIndex(w http.ResponseWriter, r *http.Request) {
	type dataT struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	}

	rand.Seed(time.Now().UnixNano())

	var data []dataT

	for i := 0; i < 20; i++ {
		data = append(data, dataT{
			ID:   xid.New().String(),
			Name: strconv.Itoa(rand.Intn(90) + 10),
			Code: strconv.Itoa(i % 2),
		})
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data:   data,
	}

	json.NewEncoder(w).Encode(response)
}

type pageData struct {
	CurrentPage string      `json:"currentPage"`
	PageSize    string      `json:"pageSize"`
	Content     interface{} `json:"content"`
}

func productIndex(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("e.txt")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	type productT struct {
		ID     string `json:"id"`
		Cname  string `json:"cname"`
		Pinyin string `json:"pinyin"`
	}

	var product []productT

	scanner := bufio.NewScanner(file)

	a := pinyin.NewArgs()
	for scanner.Scan() {
		var py string
		for _, p := range pinyin.Pinyin(scanner.Text(), a) {
			py = py + p[0] + " "
		}
		product = append(product, productT{
			ID:     xid.New().String(),
			Cname:  scanner.Text(),
			Pinyin: py,
		})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData{
			CurrentPage: "1",
			PageSize:    "10",
			Content:     product,
		},
	}

	json.NewEncoder(w).Encode(response)

}

func countryIndex(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("f.txt")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	type productT struct {
		ID     string `json:"id"`
		Cname  string `json:"cname"`
		Pinyin string `json:"pinyin"`
	}

	var product []productT

	scanner := bufio.NewScanner(file)

	a := pinyin.NewArgs()
	for scanner.Scan() {
		var py string
		for _, p := range pinyin.Pinyin(scanner.Text(), a) {
			py = py + p[0] + " "
		}
		product = append(product, productT{
			ID:     xid.New().String(),
			Cname:  scanner.Text(),
			Pinyin: py,
		})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData{
			CurrentPage: "1",
			PageSize:    "10",
			Content:     product,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func detailIndex(w http.ResponseWriter, r *http.Request) {
	type productT struct {
		BlNo         string `json:"blNo"`
		ProductID    string `json:"productId"`
		ProductName  string `json:"productName"`
		CountryID    string `json:"countryID"`
		CountryName  string `json:"countryName"`
		GrossWeight  string `json:"grossWeight"`
		PalletNumber string `json:"palletNumber"`
	}
	type detailT struct {
		GID                     string     `json:"gId"`
		ForecastUserCompany     string     `json:"forecastUserCompany"`
		CreaterRole             string     `json:"createrRole"`
		ForecastDate            string     `json:"forecastDate"`
		IsEditable              bool       `json:"isEditable"`
		ForecastTime            string     `json:"forecastTime"`
		ConfirmAreaName         string     `json:"confirmAreaName"`
		DropCabinetPositionName string     `json:"dropCabinetPositionName"`
		ForecastConfirmUser     string     `json:"forecastConfirmUser"`
		ForecastTimeName        string     `json:"forecastTimeName"`
		ForecastConfirmDate     string     `json:"forecastConfirmDate"`
		ForecastEnterDate       string     `json:"forecastEnterDate"`
		AgentCompanyID          string     `json:"agentCompanyId"`
		AgentCompanyName        string     `json:"agentCompanyName"`
		GoodsSourceID           string     `json:"goodsSourceId"`
		GoodsSourceName         string     `json:"goodsSourceName"`
		GoodsSourceCode         string     `json:"goodsSourceCode"`
		ContainerSizeID         string     `json:"containerSizeId"`
		ContainerSizeName       string     `json:"containerSizeName"`
		PlateNo                 string     `json:"plateNo"`
		ContainerNo             string     `json:"containerNo"`
		FrameNo                 string     `json:"frameNo"`
		DischargeStatus         string     `json:"dischargeStatus"`
		ElectricStatus          string     `json:"electricStatus"`
		DropCabinetPosition     string     `json:"dropCabinetPosition"`
		PrivateSiteID           string     `json:"privateSiteId"`
		PrivateSiteName         string     `json:"privateSiteName"`
		DriverTel               string     `json:"driverTel"`
		Remark                  string     `json:"remark"`
		Product                 []productT `json:"product"`
	}

	product := productT{
		BlNo:         xid.New().String(),
		ProductID:    xid.New().String(),
		ProductName:  "苹果",
		CountryID:    xid.New().String(),
		CountryName:  "中国",
		GrossWeight:  "1200",
		PalletNumber: "1500",
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data: detailT{
			GID:                     xid.New().String(),
			IsEditable:              true,
			ForecastTime:            xid.New().String(),
			ForecastTimeName:        "下午",
			ForecastEnterDate:       "2018-07-01",
			AgentCompanyID:          xid.New().String(),
			CreaterRole:             "2",
			ForecastDate:            "2018-1-1 9:12",
			ForecastUserCompany:     "预报公司",
			GoodsSourceCode:         "0",
			AgentCompanyName:        "上海欧恒进出口贸易有限公司",
			ContainerNo:             xid.New().String(),
			GoodsSourceID:           xid.New().String(),
			GoodsSourceName:         "散货",
			ConfirmAreaName:         "A-2",
			DropCabinetPositionName: "AC",
			ForecastConfirmUser:     "陈科宇",
			ForecastConfirmDate:     "2019-1-1",
			ContainerSizeID:         xid.New().String(),
			ContainerSizeName:       "40`",
			PlateNo:                 "沪A8888",
			FrameNo:                 "车架号Acdfe",
			DischargeStatus:         "0",
			ElectricStatus:          "1",
			DropCabinetPosition:     "0",
			PrivateSiteID:           xid.New().String(),
			PrivateSiteName:         "A3",
			DriverTel:               "19094546452",
			Remark:                  "备注",
			Product:                 []productT{product},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func periodIndex(w http.ResponseWriter, r *http.Request) {
	type periodT struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Time string `json:"time"`
	}

	var period []periodT
	period = append(period, periodT{
		ID:   xid.New().String(),
		Time: "1:00-2:00",
		Name: "上午",
	})
	period = append(period, periodT{
		ID:   xid.New().String(),
		Time: "1:00-2:00",
		Name: "下午",
	})
	period = append(period, periodT{
		ID:   xid.New().String(),
		Time: "1:00-2:00",
		Name: "晚上",
	})

	response := resRet{
		Result: true,
		Msg:    "",
		Data:   period,
	}

	json.NewEncoder(w).Encode(response)
}

func delIndex(w http.ResponseWriter, r *http.Request) {
	response := resRet{
		Result: true,
		Msg:    "",
	}

	json.NewEncoder(w).Encode(response)
}

func saveIndex(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1024 * 1024)

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}

	// var a Month = Januara
	// time.Time

	response := resRet{
		Result: true,
		Msg:    "该柜位有插拔电记录",
	}

	json.NewEncoder(w).Encode(response)
}

func siteIndex(w http.ResponseWriter, r *http.Request) {
	response := resRet{
		Result: true,
		Msg:    "",
		Data:   "122345",
	}
	json.NewEncoder(w).Encode(response)
}

func search(w http.ResponseWriter, r *http.Request) {
	type searchT struct {
		ForecastEnterDate       string `json:"forecastEnterDate"`
		ContainerNo             string `json:"containerNo"`
		FrameNo                 string `json:"frameNo"`
		PlateNo                 string `json:"plateNo"`
		GoodsSourceName         string `json:"goodsSourceName"`
		GoodsSourceID           string `json:"goodsSourceCode"`
		ContainerSizeName       string `json:"containerSizeName"`
		DischargeStatus         string `json:"dischargeStatus"`
		ForecastTimeName        string `json:"forecastTimeName"`
		IsPublicSite            string `json:"isPublicSite"`
		PrivateSiteName         string `json:"privateSiteName"`
		ConfirmAreaName         string `json:"confirmAreaName"`
		DropCabinetPositionName string `json:"dropCabinetPositionName"`
		Product                 string `json:"product"`
		ForecastUserCompany     string `json:"forecastUserCompany"`
		ForecastUserCompanyRole string `json:"forecastUserCompanyRole"`
		ForecastConfirmTime     string `json:"forecastConfirmTime"`
		Operator                string `json:"operator"`
		AcutalEnterTime         string `json:"acutalEnterTime"`
		AcutalEnterTimer        string `json:"acutalEnterTimer"`
		AcutalOutTime           string `json:"acutalOutTime"`
		AcutalOutTimer          string `json:"acutalOutTimer"`
		PluginTime              string `json:"pluginTime"`
		PluginTimer             string `json:"pluginTimer"`
		PlugoutTime             string `json:"plugoutTime"`
		PlugoutTimer            string `json:"plugoutTimer"`
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	cell := make([]searchT, 0)
	for i := 0; i < 100; i++ {
		cell = append(cell, searchT{
			ForecastEnterDate:       "2018-08-29",
			ContainerNo:             xid.New().String()[0:5],
			FrameNo:                 xid.New().String()[0:5],
			PlateNo:                 xid.New().String()[0:5],
			GoodsSourceName:         "散货",
			ContainerSizeName:       "#13",
			DischargeStatus:         strconv.Itoa(rd.Int() % 2),
			GoodsSourceID:           strconv.Itoa(rd.Int() % 2),
			IsPublicSite:            strconv.Itoa(rd.Int() % 2),
			PrivateSiteName:         "A34",
			ConfirmAreaName:         "A 区",
			DropCabinetPositionName: "A345",
			Product:                 "殷桃 樱桃 车厘子",
			ForecastUserCompany:     "上海欧恒进出口贸易有限公司",
			ForecastUserCompanyRole: "货代",
			ForecastConfirmTime:     "2018-10-10 23:00",
			ForecastTimeName:        "下午",
			Operator:                "陈科宇",
			AcutalEnterTime:         "2018-10-10 23:00",
			AcutalEnterTimer:        "10天 20小时 3 分",
			AcutalOutTime:           "2018-10-10 23:00",
			AcutalOutTimer:          "10天 20小时 3 分",
			PluginTime:              "2018-10-10 23:00",
			PluginTimer:             "10天 20小时 3 分",
			PlugoutTime:             "2018-10-10 23:00",
			PlugoutTimer:            "10天 20小时 3 分",
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData{
			CurrentPage: "1",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	json.NewEncoder(w).Encode(response)
}

func summary(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]string)

	r0 := rand.New(rand.NewSource(time.Now().UnixNano()))

	m["0"] = strconv.Itoa(r0.Intn(200))
	m["1"] = strconv.Itoa(r0.Intn(200))
	m["2"] = strconv.Itoa(r0.Intn(200))
	m["3"] = strconv.Itoa(r0.Intn(200))
	m["4"] = strconv.Itoa(r0.Intn(200))
	m["5"] = strconv.Itoa(r0.Intn(200))
	m["6"] = strconv.Itoa(r0.Intn(200))
	m["7"] = strconv.Itoa(r0.Intn(200))
	m["8"] = strconv.Itoa(r0.Intn(200))
	m["9"] = strconv.Itoa(r0.Intn(200))
	m["10"] = strconv.Itoa(r0.Intn(200))

	response := resRet{
		Result: true,
		Msg:    "",
		Data:   m,
	}

	json.NewEncoder(w).Encode(response)
}

func fleetLst(w http.ResponseWriter, r *http.Request) {
	type fleetT struct {
		GID               string `json:"gId"`
		ForecastEnterDate string `json:"forecastEnterDate"`
		ActualEnterDate   string `json:"actualEnterDate"`
		ContainerNo       string `json:"containerNo"`
		FrameNo           string `json:"frameNo"`
		PlateNo           string `json:"plateNo"`
		GoodsSourceName   string `json:"goodsSourceName"`
		GoodsSourceID     string `json:"goodsSourceCode"`
		ContainerSizeName string `json:"containerSizeName"`
		CustomID          string `json:"customId"`
		Product           string `json:"product"`

		ForecastUserCompany     string `json:"forecastUserCompany"`
		ForecastUserCompanyRole string `json:"forecastUserCompanyRole"`

		Fleet             string `json:"fleet"`
		FleetSelectedDate string `json:"fleetSelectedDate"`
		Operator          string `json:"operator"`
		HasDeparted       bool   `json:"hasDeparted"`
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	cell := make([]fleetT, 0)
	for i := 0; i < 100; i++ {
		cell = append(cell, fleetT{
			GID:                     xid.New().String(),
			ForecastEnterDate:       "2018-08-29",
			ActualEnterDate:         "2018-08-29",
			ContainerNo:             xid.New().String()[0:5],
			FrameNo:                 xid.New().String()[0:5],
			PlateNo:                 xid.New().String()[0:5],
			GoodsSourceName:         "散货",
			CustomID:                xid.New().String(),
			ContainerSizeName:       "#13",
			Fleet:                   "车队 A",
			FleetSelectedDate:       "2018-08-29",
			HasDeparted:             rd.Int()%2 == 0,
			GoodsSourceID:           strconv.Itoa(rd.Int() % 2),
			Product:                 "殷桃 樱桃 车厘子",
			ForecastUserCompany:     "上海欧恒进出口贸易有限公司",
			ForecastUserCompanyRole: "货代",
			Operator:                "陈科宇",
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}
	fmt.Println()

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData{
			CurrentPage: "1",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	json.NewEncoder(w).Encode(response)

}

func setFleet(w http.ResponseWriter, r *http.Request) {
	response := resRet{
		Result: true,
		Msg:    "",
	}
	json.NewEncoder(w).Encode(response)
}

func outApplicationList(w http.ResponseWriter, r *http.Request) {
	type outAppListT struct {
		GID                     string `json:"gId"`
		ContainerNo             string `json:"containerNo"`
		FrameNo                 string `json:"frameNo"`
		PlateNo                 string `json:"plateNo"`
		GoodsSourceName         string `json:"goodsSourceName"`
		GoodsSourceID           string `json:"goodsSourceCode"`
		ContainerSizeName       string `json:"containerSizeName"`
		ForecastUserCompany     string `json:"forecastUserCompany"`
		ForecastUserCompanyRole string `json:"forecastCompanyRoleName"`
		IsPublicSite            string `json:"isPublicSite"`
		PrivateSiteName         string `json:"privateSiteName"`
		ConfirmAreaName         string `json:"confirmAreaName"`
		DropCabinetPositionName string `json:"dropCabinetPositionName"`
		CancelStatus            bool   `json:"cancelStatus"`
		ApplyOutTime            string `json:"applyOutTime"`
		ApplyOutOperateTime     string `json:"applyOutOperateTime"`
		ApplyOutUser            string `json:"applyOutUser"`
		LastOutTime             string `json:"lastOutTime"`
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	cell := make([]outAppListT, 0)

	for i := 0; i < 100; i++ {
		cell = append(cell, outAppListT{
			GID:                     xid.New().String(),
			ContainerNo:             xid.New().String()[0:5],
			FrameNo:                 xid.New().String()[0:5],
			PlateNo:                 xid.New().String()[0:5],
			GoodsSourceName:         "散货",
			GoodsSourceID:           strconv.Itoa(rd.Int() % 2),
			ContainerSizeName:       "#13",
			ForecastUserCompany:     "上海欧恒进出口贸易有限公司",
			ForecastUserCompanyRole: "货代",
			IsPublicSite:            strconv.Itoa(rd.Int() % 2),
			PrivateSiteName:         "A3",
			ConfirmAreaName:         "A 区",
			DropCabinetPositionName: "A345",
			CancelStatus:            rd.Int()%2 == 0,
			ApplyOutOperateTime:     "2018-09-11",
			ApplyOutTime:            "2018-09-11",
			ApplyOutUser:            "陈科宇",
			LastOutTime:             "2018-09-11",
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}
	fmt.Println()

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData{
			CurrentPage: "1",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	json.NewEncoder(w).Encode(response)
}

func outApplicationApply(w http.ResponseWriter, r *http.Request) {
	response := resRet{
		Result: true,
		Msg:    "错误",
	}
	json.NewEncoder(w).Encode(response)
}

func outApplicationCancel(w http.ResponseWriter, r *http.Request) {
	response := resRet{
		Result: true,
		Msg:    "错误",
	}
	json.NewEncoder(w).Encode(response)
}

func pluginApplicationList(w http.ResponseWriter, r *http.Request) {
	type appT struct {
		GID                     string `json:"gId"`
		EID                     string `json:"eId"`
		ConfirmAreaName         string `json:"confirmAreaName"`
		DropCabinetPositionName string `json:"dropCabinetPositionName"`
		ContainerNo             string `json:"containerNo"`
		FrameNo                 string `json:"frameNo"`
		PlateNo                 string `json:"plateNo"`
		GoodsSourceID           string `json:"goodsSourceCode"`
		PluginDuration          string `json:"pluginDuration"`
		ApplyPlugInDate         string `json:"applyPlugInDate"`
		ApplyPlugOutDate        string `json:"applyPlugOutDate"`
		ApplyDate               string `json:"applyDate"`
		Operator                string `json:"operator"`
		CancelStatus            bool   `json:"cancelStatus"`
		PluginStatus            bool   `json:"pluginStatus"`
		PlugoutStatus           bool   `json:"plugoutStatus"`
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	cell := make([]appT, 0)

	for i := 0; i < 100; i++ {
		cell = append(cell, appT{
			GID:                     xid.New().String(),
			EID:                     xid.New().String(),
			ConfirmAreaName:         "A 区",
			DropCabinetPositionName: "A345",
			ContainerNo:             xid.New().String()[0:5],
			FrameNo:                 xid.New().String()[0:5],
			PlateNo:                 xid.New().String()[0:5],
			GoodsSourceID:           strconv.Itoa(rd.Int() % 2),
			PluginDuration:          "11 D",
			ApplyPlugInDate:         "2018-08-11",
			ApplyPlugOutDate:        "2018-08-11",
			ApplyDate:               "2018-08-11",
			Operator:                "陈科宇",
			CancelStatus:            rd.Int()%2 == 0,
			PluginStatus:            rd.Int()%2 == 0,
			PlugoutStatus:           rd.Int()%2 == 0,
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}
	fmt.Println()

	type pageData1 struct {
		CurrentPage string      `json:"currentPage"`
		PageSize    string      `json:"pageSize"`
		Content     interface{} `json:"content"`
		Total       string      `json:"total"`
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData1{
			CurrentPage: "1",
			Total:       "23123",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	json.NewEncoder(w).Encode(response)
}

func marketSettlementList(w http.ResponseWriter, r *http.Request) {
	type settlementT struct {
		GID                       string `json:"gId"`
		ActualEnterDate           string `json:"actualEnterDate"`
		ActualOutDate             string `json:"actualOutDate"`
		ContainerNo               string `json:"containerNo"`
		FrameNo                   string `json:"frameNo"`
		PlateNo                   string `json:"plateNo"`
		GoodsSourceName           string `json:"goodsSourceName"`
		ContainerSizeName         string `json:"containerSizeName"`
		Product                   string `json:"product"`
		ForecastUserCompany       string `json:"forecastUserCompany"`
		ForecastUserCompanyRole   string `json:"forecastUserCompanyRole"`
		Fee                       string `json:"fee"`
		SettlementConfirmCompany  string `json:"settlementConfirmCompany"`
		SettlementConfirmDate     string `json:"settlementConfirmDate"`
		SettlementConfirmOperator string `json:"settlementConfirmOperator"`
	}

	cell := make([]settlementT, 0)

	for i := 0; i < 100; i++ {
		cell = append(cell, settlementT{
			GID:                     xid.New().String(),
			ContainerNo:             xid.New().String()[0:5],
			FrameNo:                 xid.New().String()[0:5],
			PlateNo:                 xid.New().String()[0:5],
			GoodsSourceName:         "散货",
			ContainerSizeName:       "#13",
			Product:                 "殷桃 樱桃 车厘子",
			ForecastUserCompany:     "上海欧恒进出口贸易有限公司",
			ForecastUserCompanyRole: "货代",
			ActualEnterDate:         "2018-09-19",
			ActualOutDate:           "2018-10-10",
			Fee:                     "1200",
			SettlementConfirmCompany:  "上海欧恒进出口贸易有限公司",
			SettlementConfirmDate:     "2018-10-10",
			SettlementConfirmOperator: "陈科宇",
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}
	fmt.Println()

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData{
			CurrentPage: "1",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	json.NewEncoder(w).Encode(response)
}

func marketSettlementDetail(w http.ResponseWriter, r *http.Request) {
	type productT struct {
		Product      string `json:"product"`
		GrossWeight  string `json:"grossWeight"`
		PalletNumber string `json:"palletNumber"`
	}

	type productTA []productT

	var prod productTA
	prod = append(prod, productT{
		Product:      "苹果",
		GrossWeight:  "1200",
		PalletNumber: "345",
	})

	prod = append(prod, productT{
		Product:      "栗子",
		GrossWeight:  "2200",
		PalletNumber: "345",
	})

	prod = append(prod, productT{
		Product:      "牛油果",
		GrossWeight:  "3200",
		PalletNumber: "345",
	})

	prod = append(prod, productT{
		Product:      "樱桃",
		GrossWeight:  "110",
		PalletNumber: "345",
	})

	type pluginRecordT struct {
		PluginDate  string `json:"pluginDate"`
		PlugoutDate string `json:"plugoutDate"`
		PluginDays  string `json:"pluginDays"`
		PluginFee   string `json:"pluginFee"`
	}

	type plugRecordA []pluginRecordT

	var plugR plugRecordA
	plugR = append(plugR, pluginRecordT{
		PluginDate:  "2018-1-1",
		PlugoutDate: "2018-1-1",
		PluginDays:  "16 d",
		PluginFee:   "1200",
	})

	plugR = append(plugR, pluginRecordT{
		PluginDate:  "2018-1-1",
		PlugoutDate: "2018-1-1",
		PluginDays:  "16 d",
		PluginFee:   "1200",
	})

	plugR = append(plugR, pluginRecordT{
		PluginDate:  "2018-1-1",
		PlugoutDate: "2018-1-1",
		PluginDays:  "16 d",
		PluginFee:   "1200",
	})

	type regulationRecordT struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	type regulationA []regulationRecordT

	var regulation regulationA
	regulation = append(regulation, regulationRecordT{
		Name:  "装卸费",
		Value: "1000",
	})
	regulation = append(regulation, regulationRecordT{
		Name:  "装卸费",
		Value: "1000",
	})
	regulation = append(regulation, regulationRecordT{
		Name:  "装卸费",
		Value: "1000",
	})
	regulation = append(regulation, regulationRecordT{
		Name:  "装卸费",
		Value: "1000",
	})

	type sdetailT struct {
		GID                 string    `json:"gId"`
		ForecastCompany     string    `json:"forecastCompany"`
		GoodsSourceCode     string    `json:"goodsSourceCode"`
		GoodsSourceName     string    `json:"goodsSourceName"`
		ContainerSizeName   string    `json:"containerSizeName"`
		FrameNo             string    `json:"frameNo"`
		ContainerNo         string    `json:"containerNo"`
		PlateNo             string    `json:"plateNo"`
		IsPublicSite        string    `json:"isPublicSite"`
		ProductName         string    `json:"productName"`
		Product             productTA `json:"product"`
		EntryFee            string    `json:"entryFee"`
		LoadingFee          string    `json:"loadingFee"`
		ActualEnterDate     string    `json:"actualEnterDate"`
		ActualOutDate       string    `json:"actualOutDate"`
		DaysInVenue         string    `json:"daysInVenue"`
		VenueFee            string    `json:"venueFee"`
		PluginDays          string    `json:"pluginDays"`
		PluginFee           string    `json:"pluginFee"`
		SettleConfirmStatus string    `json:"settleConfirmStatus"`

		PluginRecord     []pluginRecordT     `json:"pluginRecord"`
		RegulationRecord []regulationRecordT `json:"regulationRecord"`

		Fee                       string `json:"totalFee"`
		SettlementConfirmCompany  string `json:"settlementConfirmCompany"`
		SettlementConfirmDate     string `json:"settlementConfirmDate"`
		SettlementConfirmOperator string `json:"settlementConfirmOperator"`
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	detail := sdetailT{
		GID:             xid.New().String(),
		ForecastCompany: "上海欧恒进出口贸易有限公司",
		// GoodsSourceCode:   strconv.Itoa(rd.Int() % 2),
		GoodsSourceCode:   "0",
		GoodsSourceName:   "散货",
		ContainerSizeName: "#13",
		ContainerNo:       "abd345",
		FrameNo:           xid.New().String()[0:5],
		PlateNo:           xid.New().String()[0:5],
		// IsPublicSite:      strconv.Itoa(rd.Intn(3)),
		IsPublicSite: "0",
		ProductName:  "殷桃 樱桃 车厘子",
		Product:      prod,

		EntryFee:            strconv.Itoa(rd.Intn(1000)),
		SettleConfirmStatus: strconv.Itoa(rd.Intn(2)),
		LoadingFee:          strconv.Itoa(rd.Intn(1000)),
		ActualEnterDate:     strconv.Itoa(rd.Intn(1000)),
		ActualOutDate:       strconv.Itoa(rd.Intn(1000)),
		DaysInVenue:         strconv.Itoa(rd.Intn(1000)),
		VenueFee:            strconv.Itoa(rd.Intn(1000)),
		PluginDays:          "11天",
		PluginFee:           "2456元",

		PluginRecord:     plugR,
		RegulationRecord: regulation,

		Fee: "12345",
		SettlementConfirmCompany:  "上海欧恒进出口贸易有限公司",
		SettlementConfirmDate:     "2018-10-10",
		SettlementConfirmOperator: "陈科宇",
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data:   detail,
	}

	json.NewEncoder(w).Encode(response)
}

func marketLauList(w http.ResponseWriter, r *http.Request) {
	type settlementT struct {
		GID               string `json:"gId"`
		ContainerNo       string `json:"containerNo"`
		FrameNo           string `json:"frameNo"`
		PlateNo           string `json:"plateNo"`
		GoodsSourceName   string `json:"goodsSourceName"`
		ContainerSizeName string `json:"containerSizeName"`

		CustomerCompany       string `json:"customerCompany"`
		FleetSelectedTime     string `json:"fleetSelectedTime"`
		FleetSelectedOperator string `json:"fleetSelectedOperator"`
		AcutalEnterTime       string `json:"acutalEnterTime"`
	}

	cell := make([]settlementT, 0)

	for i := 0; i < 100; i++ {
		cell = append(cell, settlementT{
			GID:                   xid.New().String(),
			ContainerNo:           xid.New().String()[0:5],
			FrameNo:               xid.New().String()[0:5],
			PlateNo:               xid.New().String()[0:5],
			GoodsSourceName:       "散货",
			ContainerSizeName:     "#13",
			CustomerCompany:       "上海欧恒进出口贸易有限公司",
			FleetSelectedTime:     "2018-09-19",
			AcutalEnterTime:       "2018-10-10",
			FleetSelectedOperator: "陈科宇",
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}
	fmt.Println()

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData{
			CurrentPage: "1",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	// var w0 io.Writer
	// w0 = os.Stdout
	json.NewEncoder(w).Encode(response)
}

func flutterTaskInsert(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())

	duration := time.Duration(rand.Int63n(10))
	fmt.Println(duration)
	time.Sleep(duration * time.Second)
	result := rand.Intn(10)%2 == 0
	msg := ""
	if !result {
		msg = "数据取得失败!"
	}
	response := resRet{
		Result: result,
		Msg:    msg,
	}
	json.NewEncoder(w).Encode(response)
}

func newVersion(w http.ResponseWriter, r *http.Request) {
}

func newPlatform(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	type codeRetT struct {
		Code   string `json:"code"`
		Result string `json:"reslt"`
		Des    string `json:"des"`
	}

	type upT struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	codeRet := codeRetT{
		Code:   "0",
		Des:    "登录成功",
		Result: "验证不通过",
	}

	decoder := json.NewDecoder(r.Body)
	var data upT
	err := decoder.Decode(&data)
	if err != nil {
		panic(err)
	}

	log.Println(data.Username)
	log.Println(data.Password)

	time.Sleep(1 * time.Second)

	json.NewEncoder(w).Encode(codeRet)
	// w.WriteHeader(http.StatusInternalServerError)

}

func enterStatistics(w http.ResponseWriter, r *http.Request) {
	type dataT struct {
		Total      string `json:"total"`
		EContainer string `json:"eContainer"`
		EWeight    string `json:"eWeight"`
		ETotal     string `json:"eTotal"`
		WContainer string `json:"wContainer"`

		WWeight    string `json:"wWeight"`
		WTotal     string `json:"wTotal"`
		DContainer string `json:"dContainer"`
		DWeight    string `json:"dWeight"`
		DTotal     string `json:"dTotal"`
	}

	rand.Seed(time.Now().UnixNano())

	response := resRet{
		Result: true,
		Msg:    "",
		Data: dataT{
			Total:      strconv.Itoa(rand.Intn(100)),
			EContainer: strconv.Itoa(rand.Intn(100)),
			EWeight:    strconv.Itoa(rand.Intn(100)),
			ETotal:     strconv.Itoa(rand.Intn(100)),
			WContainer: strconv.Itoa(rand.Intn(100)),
			WWeight:    strconv.Itoa(rand.Intn(100)),
			WTotal:     strconv.Itoa(rand.Intn(100)),
			DContainer: strconv.Itoa(rand.Intn(100)),
			DWeight:    strconv.Itoa(rand.Intn(100)),
			DTotal:     strconv.Itoa(rand.Intn(100)),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func enterDetailStatistics(w http.ResponseWriter, r *http.Request) {
	type dataT struct {
		Date       string `json:"date"`
		Total      string `json:"total"`
		EContainer string `json:"eContainer"`
		EWeight    string `json:"eWeight"`
		ETotal     string `json:"eTotal"`
		WContainer string `json:"wContainer"`

		WWeight    string `json:"wWeight"`
		WTotal     string `json:"wTotal"`
		DContainer string `json:"dContainer"`
		DWeight    string `json:"dWeight"`
		DTotal     string `json:"dTotal"`
	}

	cell := make([]dataT, 0)

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		cell = append(cell, dataT{
			Date:       "2018-09-10",
			Total:      strconv.Itoa(i),
			EContainer: strconv.Itoa(rand.Intn(100)),
			EWeight:    strconv.Itoa(rand.Intn(100)),
			ETotal:     strconv.Itoa(rand.Intn(100)),
			WContainer: strconv.Itoa(rand.Intn(100)),
			WWeight:    strconv.Itoa(rand.Intn(100)),
			WTotal:     strconv.Itoa(rand.Intn(100)),
			DContainer: strconv.Itoa(rand.Intn(100)),
			DWeight:    strconv.Itoa(rand.Intn(100)),
			DTotal:     strconv.Itoa(i),
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData{
			CurrentPage: "1",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	// var w0 io.Writer
	// w0 = os.Stdout
	json.NewEncoder(w).Encode(response)
}

type productDataT struct {
	Product   string `json:"product"`
	Container string `json:"container"`
	Weight    string `json:"weight"`
	Total     string `json:"total"`
}

type productDataTSlice []productDataT

type customSort struct {
	t    productDataTSlice
	less func(x, y productDataT) bool
}

func (x customSort) Len() int {
	return len(x.t)
}

func (x customSort) Less(i, j int) bool {
	return x.less(x.t[i], x.t[j])
}

func (x customSort) Swap(i, j int) {
	x.t[i], x.t[j] = x.t[j], x.t[i]
}

func enterProductStatistics(w http.ResponseWriter, r *http.Request) {

	cell := make(productDataTSlice, 0)

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		cell = append(cell, productDataT{
			Product:   "品名" + strconv.Itoa(i),
			Container: strconv.Itoa(rand.Intn(100)),
			Weight:    strconv.Itoa(rand.Intn(100)),
			Total:     strconv.Itoa(rand.Intn(100)),
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	s := r.Form["sortBy"][0]
	o := r.Form["orderBy"][0]

	cus := customSort{
		cell,
		func(x, y productDataT) bool {
			var p1, p2 int
			var err error
			switch s {
			case "0":
				p1, err = strconv.Atoi(x.Container)
				p2, err = strconv.Atoi(y.Container)
			case "1":
				p1, err = strconv.Atoi(x.Weight)
				p2, err = strconv.Atoi(y.Weight)
			case "2":
				p1, err = strconv.Atoi(x.Total)
				p2, err = strconv.Atoi(y.Total)
			}

			if err != nil {
				return false
			}

			if o == "DESC" {
				return p1 > p2
			}
			return p1 < p2
		},
	}
	sort.Sort(cus)

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}

	type pageData1 struct {
		CurrentPage string      `json:"currentPage"`
		PageSize    string      `json:"pageSize"`
		Content     interface{} `json:"content"`
		Total       string      `json:"total"`
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData1{
			CurrentPage: "1",
			Total:       "23123",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	json.NewEncoder(w).Encode(response)
}

func enterCustomerStatistics(w http.ResponseWriter, r *http.Request) {
	type dataT struct {
		Customer   string `json:"customer"`
		CustomerID string `json:"customerId"`
		Date       string `json:"date"`
		Total      string `json:"total"`
		EContainer string `json:"eContainer"`
		EWeight    string `json:"eWeight"`
		ETotal     string `json:"eTotal"`
		WContainer string `json:"wContainer"`

		WWeight    string `json:"wWeight"`
		WTotal     string `json:"wTotal"`
		DContainer string `json:"dContainer"`
		DWeight    string `json:"dWeight"`
		DTotal     string `json:"dTotal"`
	}

	cell := make([]dataT, 0)

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		cell = append(cell, dataT{
			Customer:   "客户" + strconv.Itoa(i),
			CustomerID: xid.New().String()[0:5],
			Date:       "2018-09-10",
			Total:      strconv.Itoa(rand.Intn(100)),
			EContainer: strconv.Itoa(rand.Intn(100)),
			EWeight:    strconv.Itoa(rand.Intn(100)),
			ETotal:     strconv.Itoa(rand.Intn(100)),
			WContainer: strconv.Itoa(rand.Intn(100)),
			WWeight:    strconv.Itoa(rand.Intn(100)),
			WTotal:     strconv.Itoa(rand.Intn(100)),
			DContainer: strconv.Itoa(rand.Intn(100)),
			DWeight:    strconv.Itoa(rand.Intn(100)),
			DTotal:     strconv.Itoa(rand.Intn(100)),
		})
	}

	r.ParseMultipartForm(1024 * 1024)

	c, err := strconv.Atoi(r.Form["currentPage"][0])
	if err != nil {
		c = 1
	}
	p, err := strconv.Atoi(r.Form["pageSize"][0])
	if err != nil {
		p = 10
	}

	start := (c - 1) * p
	end := c * p
	if end > len(cell) {
		end = len(cell)
	}

	for k, v := range r.Form {
		if k != "token" {
			fmt.Printf("%v = %v\n", k, v[0])
		}
	}

	type pageData1 struct {
		CurrentPage string      `json:"currentPage"`
		PageSize    string      `json:"pageSize"`
		Content     interface{} `json:"content"`
		Total       string      `json:"total"`
	}

	response := resRet{
		Result: true,
		Msg:    "",
		Data: pageData1{
			CurrentPage: "1",
			Total:       "2XXXX3123",
			PageSize:    strconv.Itoa(p),
			Content:     cell[start:end],
		},
	}

	json.NewEncoder(w).Encode(response)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// var Buf bytes.Buffer

	// fmt.Println(r.Method)
	// fmt.Println(r.Header)

	type fileST struct {
		OriginalName string `json:"originalName"`
		FileID       string `json:"fileId"`
	}

	var fileS []fileST

	if r.Method == "POST" {

		r.ParseMultipartForm(32 << 20)
		fhs := r.MultipartForm.File["file"]

		fileS = make([]fileST, len(fhs))

		for index, fh := range fhs {
			f, err := fh.Open()
			defer f.Close()
			if err != nil {
				panic(err)
			}

			fileS[index].FileID = xid.New().String()
			fileS[index].OriginalName = fh.Filename

			output, err := os.OpenFile("./file_data/"+fh.Filename, os.O_WRONLY|os.O_CREATE, 0666)
			defer output.Close()

			if err != nil {
				panic(err)
			}

			io.Copy(output, f)

			// fmt.Println(f.Filename)
			// f is one of the files
		}

		// file, handler, err := r.FormFile("file")

		// if err != nil {
		// 	fmt.Println(err)
		// }

		// fmt.Println(handler.Filename)

		// f, err := os.OpenFile("./file_data/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)

		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// io.Copy(f, file)

		// defer file.Close()

	}

	// file, header, err := r.FormFile("file")

	// if err != nil {
	// 	panic(err)
	// }

	// defer file.Close()

	// fmt.Println(header)

	// response := resRet{
	// 	Result: true,
	// 	Msg:    "",
	// 	Data:   fileS,
	// }
	if len(fileS) > 0 {
		json.NewEncoder(w).Encode(fileS[0])
	} else {
		json.NewEncoder(w).Encode("")
	}
}
