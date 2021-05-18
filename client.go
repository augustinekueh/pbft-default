package main

import(
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"log"
	"strconv"
	"time"
)

type Client struct{
	clientID string
	addr	 string
	pubKey	 []byte	
	privKey	 []byte
	message	 *RequestMsg
	replyLog map[int]*ReplyMsg
	//mutex	 sync.mutex
}

func newClient(clientID string, addr string) *Client{
	c := new(Client)
	c.clientID = clientID
	c.addr = addr
	c.pubKey = c.getPubKey(clientID)
	c.privKey = c.getPrivKey(clientID)
	c.message = nil 
	c.replyLog = make(map[int]*ReplyMsg)

	return c
}

func (c *Client) Initiate(){
	c.sendRequest()
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
	}
}

//rewrite
func (c *Client) sendRequest(){
	req := fmt.Sprintf("%d Transaction need to be approved", rand.Int())

	r := new(RequestMsg)
	r.Operation = "immediate consensus required"
	r.Timestamp = int(time.Now().Unix())
	r.ClientID = c.clientID
	r.CMessage.Request = req
	r.CAddr = c.addr

	rp, err := json.Marshal(r)
	if err != nil{
		log.Panic(err)
	}

	fmt.Println(string(rp))

	if err != nil{
		log.Panic(err)
		fmt.Printf("error happened: %d", err)
		return
	}

	packet := mergeMsg(Request, rp)
	primaryNode := findPrimaryN()
	send(packet, primaryNode.URL)
	c.message = r
}

func (c* Client) handleReply(payload []byte){
	var replyMsg ReplyMsg
	err := json.Unmarshal(payload, &replyMsg)
	rlen := len(c.replyLog)
	if err != nil{
		log.Panic(err)
		fmt.Printf("error happened: %d", err)
		return
	}

	if rlen >= countTotalMsgAmount(){
		fmt.Println("requst approved!")
	}
}

func (c *Client) getPubKey(clientID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + clientID + "/" + clientID + "_RSA_PUB")

	if err != nil{
		log.Panic(err)
	}
	return key
}

func (c *Client) getPrivKey(clientID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + clientID + "/" + clientID + "_RSA_PRIV")

	if err != nil{
		log.Panic(err)
	}
	return key
}

func findPrimaryN() JsonNode{
	//not sure correct or not..
	var primaryNode JsonNode 
	primaryID := viewID % len(jnodes.JsonNodes)
	for _, v := range jnodes.JsonNodes{
		if v.ID == strconv.Itoa(primaryID){
			continue
		}
		primaryNode = v
	}
	return primaryNode
}