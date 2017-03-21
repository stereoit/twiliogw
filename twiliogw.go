package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	twilio "github.com/carlosdp/twiliogo"
	"github.com/gorilla/mux"
)

// TwilioHandler configuration options
type TwilioHandler struct {
	Options *options
	onCall  *OnCall
}

// NewTwilioGW returns router
func NewTwilioGW(o *options) *mux.Router {
	r := mux.NewRouter()
	t := TwilioHandler{
		o,
		NewOnCall(o),
	}
	r.HandleFunc("/", t.indexHandler)
	r.Handle("/call", postJSONMiddleware(t.CallHandler))
	r.Handle("/sms", postJSONMiddleware(t.SMSHandler))
	return r
}

func (t *TwilioHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	msg, _ := json.Marshal("Nothing to see here")
	w.Write(msg)
}

// postJSONMiddleware to check for correct http method and header
func postJSONMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Only POST allowed"))
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte("Only 'application/json' Content Type allowed"))
			return
		}
		next(w, r)
	})
}

// CallHandler handles phone notifications
func (t *TwilioHandler) CallHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln("Cannot parse request body")
		return
	}
	status, _ := jsonparser.GetString(body, "status")

	if status == "firing" {
		client := &http.Client{}

		callURL := t.Options.API + "Accounts/" + t.Options.AccountID + "/Calls.json"

		data := url.Values{}
		data.Set("From", t.Options.Sender)
		data.Set("To", t.onCall.WhoIsOnCall())
		data.Set("Url", t.Options.VoiceScript)

		req, err := http.NewRequest("POST", callURL, bytes.NewBufferString(data.Encode()))
		if err != nil {
			log.Fatalln("cannot create POST request %s", err)
			return
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(t.Options.AccountID, t.Options.Token)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
			return
		}

		// relay the response back to initiator
		var buff = new(bytes.Buffer)
		buff.ReadFrom(resp.Body)
		w.Write(buff.Bytes())
	}
}

// SMSHandler handles SMS notifications
func (t *TwilioHandler) SMSHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln("Cannot parse request body")
		return
	}
	status, _ := jsonparser.GetString(body, "status")

	if status == "firing" {
		jsonparser.ArrayEach(body, func(alert []byte, dataType jsonparser.ValueType, offset int, err error) {
			go t.sendMessage(alert)
		}, "alerts")
	}

	msg, _ := json.Marshal(fmt.Sprintf("SMS sent %s", status))
	w.Write(msg)
}

func (t *TwilioHandler) sendMessage(alert []byte) {
	o := t.Options

	c := twilio.NewClient(o.AccountID, o.Token)
	body, _ := jsonparser.GetString(alert, "annotations", "summary")

	if body != "" {
		body = findAndReplaceLables(body, alert)
		startsAt, _ := jsonparser.GetString(alert, "startsAt")
		parsedStartsAt, err := time.Parse(time.RFC3339, startsAt)
		if err == nil {
			body = "\"" + body + "\"" + " alert starts at " + parsedStartsAt.Format(time.RFC1123)
		}

		// defined in oncall.go
		receiver := t.onCall.WhoIsOnCall()

		message, err := twilio.NewMessage(c, o.Sender, receiver, twilio.Body(body))
		if err != nil {
			log.Fatalln(err)
		} else {
			log.Printf("Message %s\n", message.Status)
		}
	} else {
		log.Fatalln("Bad format")
	}
}

func findAndReplaceLables(body string, alert []byte) string {
	labelReg := regexp.MustCompile(`\$labels.[a-z]+`)
	matches := labelReg.FindAllString(body, -1)

	if matches != nil {
		for _, match := range matches {
			labelName := strings.Split(match, ".")
			if len(labelName) == 2 {
				replaceWith, _ := jsonparser.GetString(alert, "labels", labelName[1])
				body = strings.Replace(body, match, replaceWith, -1)
			}
		}
	}

	return body
}
