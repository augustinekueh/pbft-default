package main

import(
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
)

//global variables
var localMessagePool = []Message{}
var viewID = 0

type Node struct{
	nodeID 				string
	addr 				string
	pubKey 				[]byte
	privKey 			[]byte
	sequenceID 			int
	view 				int
	msgQueue			chan []byte
	mutex 				sync.Mutex
	requestPool 		map[string]*RequestMsg
	prepareConfirmCount map[string]map[string]bool
	commitConfirmCount	map[string]map[string]bool
	isCommitBroadcast	map[string]bool
	isReply				map[string]bool
	msgLog		 		*MsgLog
}

type MsgLog struct{
	preprepareLog map[string]map[string]bool
	prepareLog 	  map[string]map[string]bool
	commitLog 	  map[string]map[string]bool
	replyLog	  map[string]bool
}


func newNode(nodeID string, addr string)*Node{
	n := new(Node)
	n.nodeID = nodeID
	n.addr = addr
	n.pubKey = n.getPubKey(nodeID)
	n.privKey = n.getPrivKey(nodeID)
	n.sequenceID = 0
	n.view = viewID
	n.msgQueue = make(chan []byte)
	n.mutex = sync.Mutex{}
	n.requestPool = make(map[string]*RequestMsg)
	n.msgLog = &MsgLog{
		make(map[string]map[string]bool),
		make(map[string]map[string]bool),
		make(map[string]map[string]bool),
		make(map[string]bool),
	}
	return n
}

func (n *Node) Initiate(){
	n.handleMsg()
	ln, err := net.Listen("tcp", n.addr)
	if err != nil{
		panic(err)
	}
	defer ln.Close()
	fmt.Printf("node server starts at %s\n", n.addr)
	for{
		conn, err := ln.Accept()
		if err != nil{
			panic(err)
		}
		n.handleConnection(conn)
	}
}

func (n* Node) handleConnection(conn net.Conn){
	req, err := ioutil.ReadAll(conn)
	if err != nil{
		panic(err)
	}
	n.msgQueue <- req
}

func (n *Node) addSID() int{
	seq := n.sequenceID
	n.sequenceID++
	return seq
}


func (n *Node) handleMsg(){
	for{
		data := <- n.msgQueue 
		header, payload, sig := splitMsg(data)

		switch Header(header){
		case Request:
			n.handleRequest(payload, sig)
		case PrePrepare:
			n.handlePrePrepare(payload, sig)
		case Prepare:
			n.handlePrepare(payload, sig)
		case Commit:
			n.handleCommit(payload, sig)
		}
	}
}

func (n *Node) handleRequest(payload []byte, sig []byte){
	//create instance of requestmsg
	r := new(RequestMsg)
	//convert json to struct fromat
	err := json.Unmarshal(payload, r)
	if err != nil{
		log.Panic(err)
	}
	n.addSID()
	//obtain digest of requestmsg
	digest := createDigest(*r)
	strDigest := hex.EncodeToString(digest[:])
	//store map; digest = key, requestmsg = value
	n.requestPool[strDigest] = r
	//decode string to byte format for signing
	digestByte, _ := hex.DecodeString(strDigest)
	signature, err := n.signMessage(digestByte, n.privKey)
	if err != nil{
		log.Panic(err)
	}
	prePreparePacket := PrePrepareMsg{
		*r,
		viewID,
		n.sequenceID,
		strDigest,
		signature,
	} 
	//convert struct to json format
	done, err := json.Marshal(prePreparePacket)
	if err != nil{
		log.Panic(err)
	}
	message := mergeMsg(PrePrepare, done)
	n.mutex.Lock()
	//put preprepare msg into preprepare log
	if n.msgLog.preprepareLog[prePreparePacket.Digest] == nil{
		n.msgLog.preprepareLog[prePreparePacket.Digest] = make(map[string]bool)
	}
	n.msgLog.preprepareLog[prePreparePacket.Digest][n.nodeID] = true
	n.mutex.Unlock()

	n.broadcast(message)
}

func (n *Node) handlePrePrepare(payload []byte, sig []byte){
	//create instance of preprepare
	pp := new(PrePrepareMsg)
	err := json.Unmarshal(payload, pp)
	if err != nil{
		log.Panic(err)
	}
	//get primary node's public key for verification
	primaryNodePubKey := n.getPubKey("N0")
	
	//decode string to byte format for signing
	digestByte, _ := hex.DecodeString(pp.Digest)
	//set approval conditions
	if digest := createDigest(pp.Request); hex.EncodeToString(digest[:]) != pp.Digest{
		fmt.Println("digest not match, further application rejected!")
	} else if n.sequenceID+1 != pp.SequenceID{
		fmt.Println("incorrect sequence, further application rejected!")
	} else if !n.verifySignature(digestByte, pp.Signature, primaryNodePubKey){
		fmt.Println("key verification failed, further application rejected!")
	} else {
		//success
		n.sequenceID = pp.SequenceID
		fmt.Println("stored into message pool")
		n.requestPool[pp.Digest] = &pp.Request
		signature, err := n.signMessage(digestByte, n.privKey)
		if err != nil{
			log.Panic(err)
		}
		preparePacket := PrepareMsg{
			viewID,
			pp.SequenceID,
			pp.Digest, 
			n.nodeID, 
			signature,
		}
		done, err := json.Marshal(preparePacket)
		if err != nil{
			log.Panic(err)
		}
		message := mergeMsg(Prepare, done)
		//put prepare msg into prepare log
		n.mutex.Lock()
		if n.msgLog.prepareLog[preparePacket.Digest] == nil{
			n.msgLog.prepareLog[preparePacket.Digest] = make(map[string]bool)
		} 
		n.msgLog.prepareLog[preparePacket.Digest][n.nodeID] = true
		n.mutex.Unlock()

		n.broadcast(message)
	}
}

func (n *Node) handlePrepare(payload []byte, sig []byte){
	pre := new(PrepareMsg)
	err := json.Unmarshal(payload, pre)
	if err != nil{
		log.Panic(err)
	}

	msgNodePubKey := n.getPubKey(pre.NodeID)
	//decode string to byte format 
	digestByte, _ := hex.DecodeString(pre.Digest)
	if _, ok := n.requestPool[pre.Digest]; !ok{
		fmt.Println("unable to retrieve digest, further application rejected!")
	} else if n.sequenceID != pre.SequenceID{
		fmt.Println("incorrect sequence, further application rejected!")
	} else if !n.verifySignature(digestByte, pre.Signature, msgNodePubKey){
		fmt.Println("key verification failed, further application rejected!")
	} else{
		//success
		n.setPrepareConfirmMap(pre.Digest, pre.NodeID, true)
		count := 0
		for range n.prepareConfirmCount[pre.Digest]{
			count++
		}
		specifiedCount := 0
		if n.nodeID == "N0"{
			specifiedCount = nodeCount / 3 * 2
		} else{
			specifiedCount = (nodeCount / 3 * 2) -1
		}
		n.mutex.Lock()

		if count >= specifiedCount && !n.isCommitBroadcast[pre.Digest]{
			fmt.Println("minimum (prepare) consensus achieved!")
			signature, err := n.signMessage(digestByte, n.privKey)
			if err != nil{
				log.Panic(err)
			}
			c := CommitMsg{
				viewID,
				pre.Digest,
				pre.SequenceID,
				n.nodeID,
				signature,
			}
			done, err := json.Marshal(c)
			if err != nil{
				log.Panic(err)
			} 
			fmt.Println("broadcasting commit message..")

			message := mergeMsg(Commit, done)
			//put commit msg into commit log
			n.mutex.Lock()
			if n.msgLog.commitLog[c.Digest] == nil{
				n.msgLog.commitLog[c.Digest] = make(map[string]bool)
			}
			n.msgLog.commitLog[c.Digest][n.nodeID] = true
			n.mutex.Unlock()

			n.broadcast(message)
			n.isCommitBroadcast[pre.Digest] = true
			fmt.Println("committed successfully")
			}
		n.mutex.Unlock()
	}
}

func (n *Node) handleCommit(payload []byte, sig []byte){
	cmt := new(CommitMsg)
	err := json.Unmarshal(payload, cmt)
	if err != nil{
		log.Panic(err)
	}

	msgNodePubKey := n.getPubKey(cmt.NodeID)
	digestByte, _ := hex.DecodeString(cmt.Digest)

	if _, ok := n.prepareConfirmCount[cmt.Digest]; !ok{
		fmt.Println("unable to retrive digest, further application rejected")
	} else if n.sequenceID != cmt.SequenceID{
		fmt.Println("incorrect sequence, further application rejected!")
	} else if !n.verifySignature(digestByte, cmt.Signature, msgNodePubKey){
		fmt.Println("key verification failed, further application rejected!")
	} else{
		n.setCommitConfirmMap(cmt.Digest, cmt.NodeID, true)
		count := 0
		for range n.commitConfirmCount[cmt.Digest]{
			count++
		}
		n.mutex.Lock()
		if count >= nodeCount / 3 * 2 && !n.isReply[cmt.Digest] && n.isCommitBroadcast[cmt.Digest]{
			fmt.Println("minimum (commit) consensus achieved!")

			signature, err := n.signMessage(digestByte, n.privKey)
			if err != nil{
				log.Panic(err)
			}

			d := ReplyMsg{
				viewID,
				int(time.Now().Unix()),
				//clientID,
				n.nodeID,
				"success",
				signature,
			}

			done, err := json.Marshal(d)
			if err != nil{
				log.Panic(err)
			}

			fmt.Println("broadcasting results..")
			message := mergeMsg(Reply, done)
			send(message, n.requestPool[cmt.Digest].CAddr)
			//localMessagePool = append(localMessagePool, n.requestPool[cmt.Digest].CMessage)
			//info := n.nodeID + n.requestPool[cmt.Digest].CMessage.Request
			//fmt.Println(info)
			//fmt.Println("replying client..")
			//send([]byte(info), n.messagePool[cmt.Digest].ClientAddr)
			n.isReply[cmt.Digest] = true
			fmt.Println("successfully replied!")
		}
		n.mutex.Unlock()
	}
}

/*pending methods
func (n *Node) handleCommited(){

}

func (n *Node) handleCheckpoint(){

}

func (n *Node) handleViewChange(){

}

func (n *Node) handleNewView(){

}
*/

func (n *Node) verifyRequestDigest(digest string) error{
	n.mutex.Lock()
	_, ok := n.requestPool[digest]
	if !ok{
		n.mutex.Unlock()
		return fmt.Errorf("verify request digest failed\n")
	}
	n.mutex.Unlock()
	return nil
}
/*
func (n *Node) findVerifiedPrepareMsgCount(digest string) (int, error){
	sum := 0
	n.mutex.Lock()
	for _, exist := range n.msgLog.prepareLog[digest]{
		if exist == true{
			sum++
		}
	}
	n.mutex.Unlock()
	return sum, nil
}

func (n *Node) findVerifiedCommitMsgCount(digest string) (int, error){
	sum := 0
	n.mutex.Lock()
	for _, exist := range n.msgLog.commitLog[digest]{
		if exist == true{
			sum++
		}
	}
	n.mutex.Unlock()
	return sum, nil
}
*/
func (n *Node) broadcast(data []byte){
	for i := range nodeTable{
		if i == n.nodeID{
			continue
		}
		
		go send(data, nodeTable[i])
	}
}

func (n* Node) reply(data []byte, cliaddr string){
	conn, err := net.Dial("tcp", cliaddr)
	if err != nil{
		log.Println("connect error", err)
		return 
	}
	_, err = conn.Write(data)
	if err != nil{
		log.Fatal(err)
	}
	conn.Close()
}

func (n *Node) getPubKey(nodeID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + nodeID + "/" + nodeID + "_RSA_PUB")

	if err != nil{
		log.Panic(err)
	}
	return key
}

func (n *Node) getPrivKey(nodeID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + nodeID + "/" + nodeID + "_RSA_PRIV")

	if err != nil{
		log.Panic(err)
	}
	return key
}

func (n *Node) setPrepareConfirmMap(x, y string, b bool){
	if _, ok := n.prepareConfirmCount[x]; !ok{
		n.prepareConfirmCount[x] = make(map[string]bool)
	}
	n.prepareConfirmCount[x][y] = b
}

func (n *Node) setCommitConfirmMap(x, y string, b bool){
	if _, ok := n.commitConfirmCount[x]; !ok{
		n.commitConfirmCount[x] = make(map[string]bool)
	}
	n.commitConfirmCount[x][y] = b
}