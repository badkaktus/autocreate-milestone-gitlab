package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/badkaktus/gorocket"
)

type Milestone struct {
	ID          int    `json:"id"`
	Iid         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	DueDate     string `json:"due_date"`
	StartDate   string `json:"start_date"`
	WebURL      string `json:"web_url"`
}

var glURL, glTemplate, glToken, rocketURL, rocketUser, rocketPass, rocketChannel *string
var glGroupId, mlLength *int
var client *http.Client
var rocketClient *gorocket.Client
var lastIssetDay time.Time

func main() {
	glURL = flag.String("gitlaburl", "", "GitLab URL")
	glToken = flag.String("token", "", "GitLab Private Token")
	glTemplate = flag.String("mlname", "", "Milestone template (example: Week %W/%Y)")
	glGroupId = flag.Int("group", 0, "GitLab Group ID")
	mlLength = flag.Int("mllength", 0, "Milestone length (days)")
	rocketURL = flag.String("rocketurl", "", "RocketChat URL")
	rocketUser = flag.String("user", "", "RocketChat User")
	rocketPass = flag.String("pass", "", "RocketChat Password")
	rocketChannel = flag.String("channel", "", "RocketChat channel to post")
	flag.Parse()

	client = &http.Client{}
	rocketClient = gorocket.NewClient(*rocketURL)

	payload := gorocket.LoginPayload{
		User:     *rocketUser,
		Password: *rocketPass,
	}

	loginResp, _ := rocketClient.Login(&payload)
	log.Printf("Rocket login status: %s", loginResp.Status)
	if loginResp.Message != "" {
		log.Printf("Rocket login response message: %s", loginResp.Message)
	}

	lastMilestone()
}

func lastMilestone() {

	allML := []Milestone{}

	url := fmt.Sprintf("%s/api/v4/groups/%v/milestones?state=active", *glURL, *glGroupId)

	json.Unmarshal(httpHelper("GET", url, map[string]interface{}{}), &allML)

	var lastDayOfMilestones []int64
	for _, k := range allML {
		t, _ := time.Parse("2006-01-02", k.DueDate)
		lastDayOfMilestones = append(lastDayOfMilestones, t.Unix())
	}

	sort.Slice(lastDayOfMilestones, func(i, j int) bool { return lastDayOfMilestones[i] < lastDayOfMilestones[j] })

	lastIssetDay = time.Unix(lastDayOfMilestones[len(lastDayOfMilestones)-1], 0)

	fmt.Println("lastIssetDay", lastIssetDay)

	for (lastIssetDay.Sub(time.Now()).Hours() / 24) < 14 {
		createMileStone(lastIssetDay)
	}

}

func createMileStone(lastDay time.Time) {
	milestoneStartDate := lastDay.AddDate(0, 0, 1)
	milestoneDueDate := lastDay.AddDate(0, 0, *mlLength)

	_, week := milestoneStartDate.ISOWeek()
	weekStr := strconv.Itoa(week)
	year := milestoneStartDate.Format("06")

	result := strings.Replace(*glTemplate, "%W", weekStr, -1)
	result = strings.Replace(result, "%Y", year, -1)

	data := map[string]interface{}{
		"title":      result,
		"due_date":   milestoneDueDate.Format("2006-01-02"),
		"start_date": milestoneStartDate.Format("2006-01-02"),
	}

	url := fmt.Sprintf("%s/api/v4/groups/%v/milestones", *glURL, *glGroupId)

	log.Println("Send request to " + url)
	log.Println("Send data:")
	log.Printf("%+v", data)

	httpHelper("POST", url, data)

	opt := gorocket.Message{
		Text:    fmt.Sprintf(":star_struck: Ура! У нас новый Milestone \"%v\"", result),
		Channel: *rocketChannel,
	}

	hresp, err := rocketClient.PostMessage(&opt)

	log.Printf("PostMessage response status: %v", hresp.Success)

	if err != nil || hresp.Success == false {
		log.Printf("Sending message to Rocket.Chat error")
	}
	lastIssetDay = milestoneDueDate
}

func httpHelper(method, url string, sendData map[string]interface{}) []byte {
	bytesRepresentation, err := json.Marshal(sendData)
	if err != nil {
		log.Fatalln(err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bytesRepresentation))
	req.Header.Add("Private-Token", *glToken)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		log.Printf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)

	return body
}
