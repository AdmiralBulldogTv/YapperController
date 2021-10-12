package streamelements

import (
	"bytes"
	"context"
	"regexp"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var re = regexp.MustCompile(`^(\d+)(.*)$`)

const (
	EventListenerSubscription = "subscriber-latest"
	EventListenerCheer        = "cheer-latest"
	EventListenerDonation     = "tip-latest"
)

type Client interface {
	Connect(uri string) error
	Auth(method, token string) error
	RawMessage(event string, payload interface{}) error
	Events() <-chan Event
}

type EventUpdatePayload struct {
	Name string              `json:"name"`
	Data jsoniter.RawMessage `json:"data"`
}

type EventTestPayload struct {
	Listener string              `json:"listener"`
	Event    jsoniter.RawMessage `json:"event"`
}

type Subscription struct {
	Name       string      `json:"name"`
	Amount     int         `json:"amount"`
	Tier       interface{} `json:"tier"`
	Count      int         `json:"count"`
	Gifted     bool        `json:"gifted"`
	BulkGifted bool        `json:"bulkGifted"`
	Sender     string      `json:"sender"`
	Message    string      `json:"message"`
}

type Cheer struct {
	DisplayName string `json:"displayName"`
	Amount      int    `json:"amount"`
	Name        string `json:"name"`
	Message     string `json:"message"`
}

type Donation struct {
	Name    string  `json:"name"`
	Amount  float64 `json:"amount"`
	Message string  `json:"message"`
}

type Event struct {
	Name    string
	Payload jsoniter.RawMessage
}

type cl struct {
	conn      *websocket.Conn
	connected bool
	events    chan Event
}

func NewClient() Client {
	return &cl{
		events: make(chan Event, 100),
	}
}

func (c *cl) Events() <-chan Event {
	return c.events
}

func (c *cl) Connect(uri string) error {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		return err
	}
	c.conn = conn

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c.process()
		cancel()
	}()
	go func() {
		c.ping()
		cancel()
	}()
	go func() {
		<-ctx.Done()
		c.events <- Event{
			Name: "disconnect",
		}
	}()

	return nil
}

func (c *cl) Auth(method, token string) error {
	return c.RawMessage("authenticate", map[string]string{
		"method": method,
		"token":  token,
	})
}

func (c *cl) RawMessage(event string, payload interface{}) error {
	buff := bytes.NewBuffer(nil)
	defer buff.Truncate(0)
	buff.WriteString(`420`)
	data, err := json.Marshal([]interface{}{event, payload})
	if err != nil {
		return err
	}
	_, _ = buff.Write(data)
	data = buff.Bytes()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *cl) ping() {
	conn := c.conn
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()
	defer conn.Close()
	for range tick.C {
		if err := conn.WriteMessage(websocket.TextMessage, []byte{'2'}); err != nil {
			break
		}
	}
}

func (c *cl) process() {
	conn := c.conn
	defer conn.Close()
	for {
		t, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if t != websocket.TextMessage {
			continue
		}

		matches := re.FindAllSubmatch(data, -1)

		if len(matches) == 0 {
			continue
		}

		match := matches[0]

		if !c.connected && len(match[1]) == 2 && match[1][0] == '4' && match[1][1] == '0' {
			c.connected = true
			c.events <- Event{
				Name: "connect",
			}
			continue
		}

		if len(match[2]) == 0 {
			continue
		}

		parts := []jsoniter.RawMessage{}
		if err := json.Unmarshal(match[2], parts); err != nil {
			logrus.WithError(err).Error("failed to parse parts")
			continue
		}

		name := ""
		if err := jsoniter.Unmarshal(parts[1], &name); err != nil {
			logrus.WithError(err).Error("failed to parse event name")
			continue
		}

		c.events <- Event{
			Name:    name,
			Payload: parts[1],
		}
	}
}

// (streamelements.Event) {
//  Name: (string) (len=10) "event:test",
//  Payload: (jsoniter.RawMessage) (len=216 cap=216) {
//   00000000  7b 22 6c 69 73 74 65 6e  65 72 22 3a 22 73 75 62  |{"listener":"sub|
//   00000010  73 63 72 69 62 65 72 2d  6c 61 74 65 73 74 22 2c  |scriber-latest",|
//   00000020  22 65 76 65 6e 74 22 3a  7b 22 70 72 6f 76 69 64  |"event":{"provid|
//   00000030  65 72 49 64 22 3a 22 35  35 37 32 36 30 36 39 36  |erId":"557260696|
//   00000040  22 2c 22 64 69 73 70 6c  61 79 4e 61 6d 65 22 3a  |","displayName":|
//   00000050  22 54 69 65 5f 4b 6e 65  65 22 2c 22 61 6d 6f 75  |"Tie_Knee","amou|
//   00000060  6e 74 22 3a 31 2c 22 74  69 65 72 22 3a 22 31 30  |nt":1,"tier":"10|
//   00000070  30 30 22 2c 22 71 75 61  6e 74 69 74 79 22 3a 30  |00","quantity":0|
//   00000080  2c 22 61 76 61 74 61 72  22 3a 22 68 74 74 70 73  |,"avatar":"https|
//   00000090  3a 2f 2f 63 64 6e 2e 73  74 72 65 61 6d 65 6c 65  |://cdn.streamele|
//   000000a0  6d 65 6e 74 73 2e 63 6f  6d 2f 73 74 61 74 69 63  |ments.com/static|
//   000000b0  2f 64 65 66 61 75 6c 74  2d 61 76 61 74 61 72 2e  |/default-avatar.|
//   000000c0  70 6e 67 22 2c 22 6e 61  6d 65 22 3a 22 74 69 65  |png","name":"tie|
//   000000d0  5f 6b 6e 65 65 22 7d 7d                           |_knee"}}|
//  }
// }

// (streamelements.Event) {
//  Name: (string) (len=10) "event:test",
//  Payload: (jsoniter.RawMessage) (len=258 cap=258) {
//   00000000  7b 22 6c 69 73 74 65 6e  65 72 22 3a 22 73 75 62  |{"listener":"sub|
//   00000010  73 63 72 69 62 65 72 2d  6c 61 74 65 73 74 22 2c  |scriber-latest",|
//   00000020  22 65 76 65 6e 74 22 3a  7b 22 70 72 6f 76 69 64  |"event":{"provid|
//   00000030  65 72 49 64 22 3a 22 31  35 31 35 30 32 38 35 31  |erId":"151502851|
//   00000040  22 2c 22 64 69 73 70 6c  61 79 4e 61 6d 65 22 3a  |","displayName":|
//   00000050  22 46 65 6c 69 70 65 4c  4c 73 22 2c 22 61 6d 6f  |"FelipeLLs","amo|
//   00000060  75 6e 74 22 3a 32 2c 22  74 69 65 72 22 3a 22 70  |unt":2,"tier":"p|
//   00000070  72 69 6d 65 22 2c 22 6d  65 73 73 61 67 65 22 3a  |rime","message":|
//   00000080  22 48 69 20 73 69 72 20  43 4f 4d 45 20 54 4f 20  |"Hi sir COME TO |
//   00000090  42 52 41 5a 49 4c 20 45  4c 45 47 22 2c 22 71 75  |BRAZIL ELEG","qu|
//   000000a0  61 6e 74 69 74 79 22 3a  30 2c 22 61 76 61 74 61  |antity":0,"avata|
//   000000b0  72 22 3a 22 68 74 74 70  73 3a 2f 2f 63 64 6e 2e  |r":"https://cdn.|
//   000000c0  73 74 72 65 61 6d 65 6c  65 6d 65 6e 74 73 2e 63  |streamelements.c|
//   000000d0  6f 6d 2f 73 74 61 74 69  63 2f 64 65 66 61 75 6c  |om/static/defaul|
//   000000e0  74 2d 61 76 61 74 61 72  2e 70 6e 67 22 2c 22 6e  |t-avatar.png","n|
//   000000f0  61 6d 65 22 3a 22 66 65  6c 69 70 65 6c 6c 73 22  |ame":"felipells"|
//   00000100  7d 7d                                             |}}|
//  }
// }

// (streamelements.Event) {
//  Name: (string) (len=10) "event:test",
//  Payload: (jsoniter.RawMessage) (len=221 cap=221) {
//   00000000  7b 22 6c 69 73 74 65 6e  65 72 22 3a 22 74 69 70  |{"listener":"tip|
//   00000010  2d 6c 61 74 65 73 74 22  2c 22 65 76 65 6e 74 22  |-latest","event"|
//   00000020  3a 7b 22 74 69 70 49 64  22 3a 22 36 31 34 30 37  |:{"tipId":"61407|
//   00000030  61 36 61 35 61 39 65 30  38 34 36 64 37 61 35 62  |a6a5a9e0846d7a5b|
//   00000040  31 38 63 22 2c 22 61 6d  6f 75 6e 74 22 3a 33 2c  |18c","amount":3,|
//   00000050  22 63 75 72 72 65 6e 63  79 22 3a 22 45 55 52 22  |"currency":"EUR"|
//   00000060  2c 22 6d 65 73 73 61 67  65 22 3a 22 4a 75 73 74  |,"message":"Just|
//   00000070  20 6b 69 6c 6c 20 79 6f  75 72 20 77 69 66 65 20  | kill your wife |
//   00000080  61 6c 72 65 61 64 79 22  2c 22 61 76 61 74 61 72  |already","avatar|
//   00000090  22 3a 22 68 74 74 70 73  3a 2f 2f 63 64 6e 2e 73  |":"https://cdn.s|
//   000000a0  74 72 65 61 6d 65 6c 65  6d 65 6e 74 73 2e 63 6f  |treamelements.co|
//   000000b0  6d 2f 73 74 61 74 69 63  2f 64 65 66 61 75 6c 74  |m/static/default|
//   000000c0  2d 61 76 61 74 61 72 2e  70 6e 67 22 2c 22 6e 61  |-avatar.png","na|
//   000000d0  6d 65 22 3a 22 42 75 6e  6b 34 22 7d 7d           |me":"Bunk4"}}|
//  }
// }

// (streamelements.Event) {
//  Name: (string) (len=10) "event:test",
//  Payload: (jsoniter.RawMessage) (len=338 cap=338) {
//   00000000  7b 22 6c 69 73 74 65 6e  65 72 22 3a 22 63 68 65  |{"listener":"che|
//   00000010  65 72 2d 6c 61 74 65 73  74 22 2c 22 65 76 65 6e  |er-latest","even|
//   00000020  74 22 3a 7b 22 70 72 6f  76 69 64 65 72 49 64 22  |t":{"providerId"|
//   00000030  3a 22 31 39 37 33 37 35  36 35 33 22 2c 22 64 69  |:"197375653","di|
//   00000040  73 70 6c 61 79 4e 61 6d  65 22 3a 22 52 65 64 61  |splayName":"Reda|
//   00000050  72 64 39 32 22 2c 22 61  6d 6f 75 6e 74 22 3a 33  |rd92","amount":3|
//   00000060  30 30 2c 22 6d 65 73 73  61 67 65 22 3a 22 4d 79  |00,"message":"My|
//   00000070  20 6e 61 6d 65 20 69 73  20 42 75 6c 6c 64 6f 67  | name is Bulldog|
//   00000080  2e 20 4d 79 20 67 61 6d  65 73 20 61 72 65 20 66  |. My games are f|
//   00000090  75 6c 6c 20 6f 66 20 50  6f 67 2e 20 41 72 63 68  |ull of Pog. Arch|
//   000000a0  20 69 73 20 61 20 72 61  63 69 73 74 20 66 72 6f  | is a racist fro|
//   000000b0  67 2e 20 49 26 23 33 39  3b 6d 20 61 20 77 65 65  |g. I&#39;m a wee|
//   000000c0  62 20 49 20 77 61 74 63  68 20 68 65 6e 74 61 69  |b I watch hentai|
//   000000d0  20 62 75 74 20 64 65 6c  65 74 65 20 6d 79 20 6c  | but delete my l|
//   000000e0  6f 67 2e 20 63 68 65 65  72 33 30 30 22 2c 22 71  |og. cheer300","q|
//   000000f0  75 61 6e 74 69 74 79 22  3a 30 2c 22 61 76 61 74  |uantity":0,"avat|
//   00000100  61 72 22 3a 22 68 74 74  70 73 3a 2f 2f 63 64 6e  |ar":"https://cdn|
//   00000110  2e 73 74 72 65 61 6d 65  6c 65 6d 65 6e 74 73 2e  |.streamelements.|
//   00000120  63 6f 6d 2f 73 74 61 74  69 63 2f 64 65 66 61 75  |com/static/defau|
//   00000130  6c 74 2d 61 76 61 74 61  72 2e 70 6e 67 22 2c 22  |lt-avatar.png","|
//   00000140  6e 61 6d 65 22 3a 22 72  65 64 61 72 64 39 32 22  |name":"redard92"|
//   00000150  7d 7d                                             |}}|
//  }
// }
