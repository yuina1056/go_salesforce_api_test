package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

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
	values.Add("password", os.Getenv("password"))
	res, err := http.PostForm(os.Getenv("loginURL"), values)
	if err != nil {
		log.Fatal(err)
	}
	var session map[string]string
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&session); err != nil {
		log.Fatal(err)
	}
	log.Printf("Login successful. Instance: %s", session["instance_url"])

	/* API叩くアレ */

	req, err := http.NewRequest(http.MethodGet, session["instance_url"]+"/services/data/v41.0/query/", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", session["token_type"]+" "+session["access_token"])
	req.Header.Add("Accept", "application/json")
	values = url.Values{}
	values.Add("q", "SELECT Amount FROM Opportunity")
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
	fmt.Printf("%v\n", out)
}
