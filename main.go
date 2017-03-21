package main

import (
	"log"
	"net/http"
	"os"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type options struct {
	AccountID     string `env:"TWILIO_ACCOUNT_SID"`
	Token         string `env:"TWILIO_TOKEN"`
	Sender        string `env:"TWILIO_SENDER"`
	API           string `env:"TWILIO_API" envDefault:"https://api.twilio.com/2010-04-01/"`
	VoiceScript   string `env:"TWILIO_VOICE_SCRIPT"`
	ListenAddres  string `env:"LISTEN_ADDRESS" envDefault:"8087"`
	SheetID       string `env:"ONCALL_SHEET_ID"`
	DefaultOnDuty string `env:"ONCALL_DEFAULT_RECEIVER"`
	OffShiftStart string `env:"ONCALL_OFFSHIFT_START" envDefault:"0"`
	OffShiftStop  string `env:"ONCALL_OFFSHIFT_STOP" envDefault:"7"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Failed to parse")
	}

	var opts = options{}
	err = env.Parse(&opts)
	if err != nil {
		log.Fatalln("%+v", err)
	}

	if opts.AccountID == "" || opts.Token == "" || opts.Sender == "" {
		log.Fatalln("'SID', 'TOKEN', and 'SENDER' environment variables need to be set")
		os.Exit(1)
	}

	twilioGW := NewTwilioGW(&opts)
	http.Handle("/", twilioGW)

	log.Printf("Starting server at %s\n", opts.ListenAddres)
	if err := http.ListenAndServe(opts.ListenAddres, nil); err != nil {
		log.Fatalln("ListendAndServer: ", err)
	}

}
