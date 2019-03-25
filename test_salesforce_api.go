package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/joho/godotenv"
	"github.com/koron/go-dproxy"
)

type SalesAmount struct {
	ID        bson.ObjectId `bson:"_id"`
	Name      string        `bson:"Name"`
	Amount    int64         `bson:"Amount"`
	CloseDate string        `bson:"CloseDate"`
}

func main() {
	/* .env読み込み */
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	/* API認証 */
	values := url.Values{}
	values.Add("grant_type", "password")
	values.Add("client_id", os.Getenv("client_id"))
	values.Add("client_secret", os.Getenv("client_secret"))
	values.Add("username", os.Getenv("username"))
	values.Add("password", os.Getenv("password")+os.Getenv("token"))
	res, err := http.PostForm(os.Getenv("loginURL"), values)
	if err != nil {
		log.Fatal(err)
	}
	var session map[string]string
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&session); err != nil {
		log.Fatal(err)
	}
	log.Print(res)
	log.Printf("Login successful. Instance: %s", session["instance_url"])

	/* API叩くアレ */

	req, err := http.NewRequest(http.MethodGet, session["instance_url"]+"/services/data/v41.0/query/", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", session["token_type"]+" "+session["access_token"])
	req.Header.Add("Accept", "application/json")
	values = url.Values{}
	values.Add("q", "SELECT Name,Amount,CloseDate FROM Opportunity")
	log.Printf("POST %s\n", req.URL)
	req.URL.RawQuery = values.Encode()
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	decoder = json.NewDecoder(res.Body)
	var out interface{}
	if err = decoder.Decode(&out); err != nil {
		log.Fatal(err)
	}

	outdproxy := dproxy.New(out)
	beforerecordlength, err := outdproxy.M("totalSize").Int64()
	if err != nil {
		log.Fatal(err)
	}
	var recordlength = int(beforerecordlength)

	resultarray, err := outdproxy.M("records").Array()
	if err != nil {
		log.Fatal(err)
	}

	dbsession, _ := mgo.Dial(os.Getenv("MONGODB_URI"))
	defer dbsession.Close()
	db := dbsession.DB("heroku_zb22vxl8")

	count := 0
	for count < recordlength {
		resultdproxy := dproxy.New(resultarray[count])
		name, err := resultdproxy.M("Name").String()
		if err != nil {
			log.Fatal(err)
		}
		amount, err := resultdproxy.M("Amount").Int64()
		if err != nil {
			log.Fatal(err)
		}
		closedate, err := resultdproxy.M("CloseDate").String()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("name:%s,amount:%d,closedate:%s", name, amount, closedate)
		fmt.Print("\n")

		salesamount := &SalesAmount{
			ID:        bson.NewObjectId(),
			Name:      name,
			Amount:    amount,
			CloseDate: closedate,
		}
		col := db.C("Amount")
		if err := col.Insert(salesamount); err != nil {
			log.Fatalln(err)
		}

		count++
	}

}
