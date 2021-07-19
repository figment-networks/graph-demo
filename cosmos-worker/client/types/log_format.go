package types

import (
	"bytes"
	"encoding/json"
)

// LogFormat format of logs from cosmos
type LogFormat struct {
	MsgIndex float64     `json:"msg_index"`
	Success  bool        `json:"success"`
	Log      string      `json:"log"`
	Events   []LogEvents `json:"events"`
}

// LogEvents format of events from logs cosmos
type LogEvents struct {
	Type string `json:"type"`
	//Attributes []string `json:"attributes"`
	Attributes []*LogEventsAttributes `json:"attributes"`
}

// LogEventsAttributes enhanced format of event attributes
type LogEventsAttributes struct {
	Module         string
	Action         string
	Amount         []string
	Sender         []string
	Validator      map[string][]string
	Withdraw       map[string][]string
	Recipient      []string
	CompletionTime string
	Commission     []string
	Others         map[string][]string
}

type kvHolder struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// UnmarshalJSON LogEvents into a different format,
// to be able to parse it later more easily
// thats fulfillment of json.Unmarshaler inferface
func (lea *LogEventsAttributes) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	kc := &kvHolder{}
	for dec.More() {
		err := dec.Decode(kc)
		if err != nil {
			return err
		}
		switch kc.Key {
		case "sender":
			lea.Sender = append(lea.Sender, kc.Value)
		case "recipient":
			lea.Recipient = append(lea.Recipient, kc.Value)
		case "module":
			lea.Module = kc.Value
		case "action":
			lea.Action = kc.Value
		case "amount":
			lea.Amount = append(lea.Amount, kc.Value)
		default:
			if lea.Others == nil {
				lea.Others = map[string][]string{}
			}

			k, ok := lea.Others[kc.Key]
			if !ok {
				k = []string{}
			}
			lea.Others[kc.Key] = append(k, kc.Value)
		}
	}
	return nil
}
