package wsclientsdk


const (
	REQUEST_TYPE = iota

	RESPONSE_TYPE
)

type TempCommand struct {
	RequestId string
	Type      int // 0 request 1 reponse
	Command   string
}

type WsCommand struct {
	Type int // 0 request 1 reponse

	Command   string
	RequestId string
	Data      interface{}
}
type WsResult struct {
	Type int // 0 request 1 reponse

	RequestId string
	Command   string

	Status int
	Msg    string
}

func NewCommand(name string, data interface{}) *WsCommand {

	return &WsCommand{

		Type:    REQUEST_TYPE,
		Command: name,
		Data:    data,
	}
}

