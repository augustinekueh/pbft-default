package main

import(
	"fmt"
)

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
	Digest		[]byte	`json:"digest"`
}

//<REQUEST, o, t, c> with digital signature, σ
type RequestMsg struct{
	Operation	string  `json:"operation"`
	Timestamp	int     `json:"timestamp"`
	ClientID	string  `json:"clientID"`
	CMessage	Message `json:"clientmessage"`
	CAddr		string	`json:"clientaddress"`
	//Signature 	[]byte  `json:"signature"`
}

//<PREPREPARE, v, n, d> with digital signature, σ, m>
//According to the original paper, client requests are not included in pre-prepare packets to keep them small
type PrePrepareMsg struct{
	Request		RequestMsg	`json:"Request"`
	View 		int 		`json:"view"`
	SequenceID	int 		`json:"sequenceID"`
	//digest for message, m
	Digest		string		`json:"digest"`
	//Signature	[]byte		`json:"signature"`
	//NodeID		string		`json: "nodeID"`
}

//<PREPARE, v, n, d, i> with digital signature, σ
type PrepareMsg struct{
	View		int		`json:"view"`
	SequenceID	int		`json:"sequenceID"`
	Digest		string 	`json:"digest"`
	//nodeID = the current/sender ID of this prepare msg
	NodeID		string	`json:"nodeID"`
	//Signature	[]byte	`json:"signature"`	
}

//<COMMIT, v, n, d, i> with digital signature, σ
type CommitMsg struct{
	View		int		`json:"view"`
	Digest		string 	`json:"digest"`
	SequenceID	int		`json:"sequenceID"`
	NodeID		string 	`json:"nodeID"`
	//Signature	[]byte	`json:"signature"`
}

//<RESULT, v, t, c, i, r>
type ReplyMsg struct{
	View		int	   `json:"view"`
	Timestamp	int	   `json:"timestamp"`
	//ClientID	string `json:"clientID"`
	NodeID		string `json:"nodeID"`
	Result		string `json:"result"`
	//Signature	[]byte	`json:"signature"`
}


//need to add signature
func mergeMsg(header Header, payload []byte, sig []byte) []byte{
	first := make([]byte, headerLength)
	for i, v := range []byte(header){
		first[i] = v
	} 
	
	res := make([]byte, headerLength + len(payload) + len(sig))
	copy(res[:headerLength], first)
	copy(res[headerLength:len(res)-256], payload)
	copy(res[len(res)-256:], sig)
	return res
}

func splitMsg(message []byte) (Header, []byte, []byte){
	fmt.Println("breakpoint2")
	var header		Header
	var payload		[]byte
	var signature	[]byte 

	fmt.Println("breakpoint3")
	headerBytes := message[:headerLength]
	newHeaderBytes := make([]byte, 0)

	fmt.Println("breakpoint4")
	for _, h := range headerBytes{
		if h != byte(0){
			newHeaderBytes = append(newHeaderBytes, h)
		}
	}

	fmt.Println("breakpoint5")
	header = Header(newHeaderBytes)
	fmt.Println("breakpoint6")
	switch header{
	case Request, PrePrepare, Prepare, Commit, Reply:
		fmt.Println("breakpoint7")
		//fmt.Println(len(message))
		payload = message[headerLength:len(message)-256]
		fmt.Println("breakpoint8")
		signature = message[len(message)-256:]

	/*case Reply://here problem
		payload = message[headerLength:]
		signature = []byte{}*/
	}
	fmt.Println("breakpoint9")
	return header, payload, signature
}
