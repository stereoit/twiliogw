# Prometheus Twilio Gateway

This gateway to [Twilio](https://twilio.com) allows is to receive alarms either on `/sms` or `/call` endpoint (in alertmanager we can decided what is used based on priority).

When alarm is received we check online google sheet to see who has currently on call duty. In case of out of office hours (0-7AM) we send notification to the default manager.

## Configuration

Application is controled via environment variables:

- `TWILIO_ACCOUNT_SID` - Twilio Account SID
- `TWILIO_TOKEN` - Twilio Auth Token
- `TWILIO_SENDER` - Phone number managed by Twilio
- `TWILIO_API="https://api.twilio.com/2010-04-01/"` - Twilio current API
- `TWILIO_VOICE_SCRIPT` - Twilio voice script (used by calls)
- `LISTEN_ADDRESS=":9090"` - Server IP address and port
- `ONCALL_SHEET_ID` - Google sheet ID for determining current on call duty
- `ONCALL_DEFAULT_RECEIVER` - Default phone number for main manager
- `ONCALL_OFFSHIFT_START="0"` - Out of office hours start
- `ONCALL_OFFSHIFT_STOP="7"` - Out of office hours stop


## Test it

```bash
$ curl -H "Content-Type: application/json" -X POST -d \
'{"version":"2","status":"firing","alerts":[{"annotations":{"summary":"Server down"},"startsAt":"2017-03-19T05:54:01Z"}]}' \
http://localhost:9090/sms
```

## Build it

In order to build the source, one needs to have [GO language](https://golang.org) installed, we will do static build.

```bash
$ CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o twiliogw .
$ docker build -t twiliogw .
```

## Run it



```bash
$ sudo docker run -v /path/to/.env:/.env --name twiliogw twiliogw
```

## Docker

Docker instructions are taken from [Bulding Minimal Docker](https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/)



