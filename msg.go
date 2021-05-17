package main

//global variables
type Header string
const headerLength = 12

//header title
const(
	Request		Header = "Request"
	PrePrepare	Header = "PrePrepare"
	Prepare		Header = "Prepare"
	Commit		Header = "Commit"
	Reply		Header = "Reply"
)

//message format
type Message struct{
	Request		string	`json:"message"`
	Digest		string	`json:"digest"`
}

//<REQUEST, o, t, c> with digital signature, σ
type RequestMsg struct{
	Operation	string `json:"operation"`
	Timestamp	int    `json:"timestamp"`
	ClientID	string `json:"clientID"`
	//Signature 	[]byte `json: "signature"`
	CMessage	Message `json:"clientmessage"`
	CAddr	string	`json:"clientaddress"`
}

//<PRE-PREPREPARE, v, n, d> with digital signature, σ, m>
//According to the original paper, client requests are not included in pre-prepare packets to keep them small
type PrePrepareMsg struct{
	Request		RequestMsg	`json:"Request"`
	View 		int 		`json:"view"`
	SequenceID	int 		`json:"sequenceID"`
	//digest for message, m
	Digest		string		`json:"digest"`
	Signature	[]byte		`json:"signature"`
	//NodeID		string		`json: "nodeID"`
}

//<PREPARE, v, n, d, i> with digital signature, σ
type PrepareMsg struct{
	View		int		`json:"view"`
	SequenceID	int		`json:"sequenceID"`
	Digest		string 	`json:"digest"`
	//nodeID = the current/sender ID of this prepare msg
	NodeID		string	`json:"nodeID"`
	Signature	[]byte	`json:"signature"`	
}

//<COMMIT, v, n, d, i> with digital signature, σ
type CommitMsg struct{
	View		int		`json:"view"`
	Digest		string 	`json:"digest"`
	SequenceID	int		`json:"sequenceID"`
	NodeID		string 	`json:"nodeID"`
	Signature	[]byte	`json:"signature"`
}

//<RESULT, v, t, c, i, r>
type ReplyMsg struct{
	View		int	   `json:"view"`
	Timestamp	int	   `json:"timestamp"`
	//ClientID	string `json:"clientID"`
	NodeID		string `json:"nodeID"`
	Result		string `json:"result"`
	Signature	[]byte	`json:"signature"`
}

func mergeMsg(header Header, payload []byte) []byte{
	first := make([]byte, headerLength)
	for i, v := range []byte(header){
		first[i] = v
	} 
	
	res := make([]byte, headerLength + len(payload))
	copy(res[:headerLength], first)
	copy(res[headerLength:], payload)
	
	return res
}

func splitMsg(message []byte) (Header, []byte, []byte){
	var header		Header
	var payload		[]byte
	var signature	[]byte 

	headerBytes := message[:headerLength]
	newHeaderBytes := make([]byte, 0)
	for _, h := range headerBytes{
		if h != byte(0){
			newHeaderBytes = append(newHeaderBytes, h)
		}
	}

	header = Header(newHeaderBytes)
	switch header{
	case Request, PrePrepare, Prepare, Commit:
		payload = message[headerLength:len(message)-256]
		signature = message[len(message)-256:]

	case Reply:
		payload = message[headerLength:]
		signature = []byte{}
	}

	return header, payload, signature
}

/*
func printMsgLog(msg Msg){
	fmt.Println(msg.String())
}

func logHandleMsg(header Header, msg Msg, nodeID string){
	fmt.Printf("Receive %s msg packet from localhost: %d\n", header, nodeID)
	printMsgLog(msg)
}

func logBroadcastMsg(header Header, msg Msg){
	fmt.Printf("send/broadcast %s msg \n", header)
	printMsgLog(msg)
}
*/