package main

import(
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"log"
	"time"
)

const delay = 5 * time.Second

type Client struct{
	clientID string
	addr	 string
	pubKey	 []byte	
	privKey	 []byte
	message	 *RequestMsg
	replyLog map[string]*ReplyMsg
	primaryTable map[string]string
}

func newClient(clientID string, addr string, primaryTable map[string]string) *Client{
	c := new(Client)
	c.clientID = clientID
	c.addr = addr
	c.pubKey = getPubKey(clientID)
	c.privKey = getPrivKey(clientID)
	c.message = nil 
	c.replyLog = make(map[string]*ReplyMsg)
	c.primaryTable = primaryTable

	return c
}

func (c *Client) Initiate(){
	//start := time.Now()
	//fmt.Println(start)
	//fmt.Println("breakpoint")
	ping := func(){
		c.sendRequest()}
	c.transactionSchedule(ping, delay)
	ln, err := net.Listen("tcp", c.addr)
	if err != nil{
		log.Panic(err)
	}
	defer ln.Close()
	for{
		conn, err := ln.Accept()
		if err != nil{
			log.Panic(err)
		}

		go c.handleConnection(conn)
		//duration := time.Since(start)
		//fmt.Printf("Execution time: %v\n", duration)
	}
}

func (c *Client) handleConnection(conn net.Conn){
	req, err := ioutil.ReadAll(conn)
	header, payload, _ := splitMsg(req)
	if err != nil{
		log.Panic(err)
	}
	switch header{
	case Reply:
		c.handleReply(payload)
		return
	}
}

func (c *Client) sendRequest(){
	req := fmt.Sprintf("%d Transaction need to be approved", rand.Int())

	r := new(RequestMsg)
	r.Operation = "immediate consensus required, please do it now"
	r.Timestamp = int(time.Now().Unix()) //volatile, hash will be different
	r.ClientID = c.clientID
	r.CMessage.Request = req
	r.CMessage.Digest = generateDigest(req)
	r.CAddr = c.addr
	
	sig, err := signMessage(generateDigest(req), c.privKey)
	if err != nil{
		log.Panic(err)
		fmt.Printf("error happened: %d", err)
		return
	}
	
	rp, err := json.Marshal(r)
	if err != nil{
		log.Panic(err)
	}
	fmt.Println("breakpoint2")
	//primaryNode := findPrimaryN()
	//add mergemsg
	for _, v := range c.primaryTable{
		fmt.Println(v)
		send(mergeMsg(Request, rp, sig), v)
	}
	c.message = r
}

func (c* Client) handleReply(payload []byte){
	rep := new(ReplyMsg)
	err := json.Unmarshal(payload, &rep)
	fmt.Println(rep)
	c.replyLog[rep.NodeID] = rep
	rlen := len(c.replyLog)
	fmt.Println(rlen)
	if err != nil{
		fmt.Printf("error happened: %d", err)
		return
	}
	if rlen >= countTotalMsgAmount(){
		fmt.Println("request approved!")
	}
}

// func findPrimaryN() JsonNode{ //need to improve here; probably grab data from the json file (primary nodes) or at main.go.
// 	var primaryNode JsonNode = JsonNode{
// 		"N0",
// 		"127.0.0.1:8080",
// 	} 
// 	return primaryNode
// }

func (c* Client) transactionSchedule(ping func(), delay time.Duration)chan bool{
	stop := make(chan bool)

	go func(){
		for{
			start := time.Now()
			fmt.Println(start)
			ping()	
			select{
			case <- time.After(delay):
				c.replyLog = make(map[string]*ReplyMsg) // clear message log
				duration := time.Since(start) - delay
				fmt.Printf("Execution time: %v\n", duration)
				write := fmt.Sprintf("Duaration: %v\n", duration)
				if !isExist("layer_test_results.txt"){
					err := ioutil.WriteFile("layer_test_results.txt", []byte(write), 0644)
					if err != nil{
						log.Panicf("datalog: error creating file: %s", err)
					}
				}
				f, err :=os.OpenFile("layer_test_results.txt", os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil{
					log.Panicf("datalog: error opening file: %s", err)
				}
				defer f.Close() 

				if _, err = f.WriteString(write); err != nil{
					log.Panicf("error recording results: %s", err)

				}
			case <- stop:
				return
			}
		}
	}()
	return stop
}