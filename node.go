package main

import(
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//global variables
var localMessagePool = []Message{}
var viewID = 0
var nodeDelay = 200 * time.Millisecond

type Node struct{
	nodeID 				string
	addr 				string
	pubKey 				[]byte
	privKey 			[]byte
	sequenceID 			int
	view 				int
	msgQueue			chan []byte
	mutex 				sync.Mutex
	nodeTable			map[string]string
	requestPool 		map[string]*RequestMsg
	prepareConfirmCount map[string]map[string]bool
	commitConfirmCount	map[string]map[string]bool
	isCommitBroadcast	map[string]bool
	isReply				map[string]bool
	msgLog		 		*MsgLog
	primary			    string
	totalPrimaryTable	map[string]string
	broadcastAddr		string
	block				int
	vicePrimary			string
}

type MsgLog struct{
	preprepareLog map[string]map[string]bool
	prepareLog 	  map[string]map[string]bool
	commitLog 	  map[string]map[string]bool
	replyLog	  map[string]bool
}


func newNode(nodeID string, addr string, nodeTable map[string]string, totalPrimaryTable map[string]string)*Node{
	n := new(Node)
	n.nodeID = nodeID
	n.addr = addr
	n.pubKey = getPubKey(nodeID)
	n.privKey = getPrivKey(nodeID)
	n.sequenceID = 0
	n.view = viewID
	n.msgQueue = make(chan []byte)
	n.mutex = sync.Mutex{}
	n.nodeTable = nodeTable
	n.requestPool = make(map[string]*RequestMsg)
	n.prepareConfirmCount = make(map[string]map[string]bool)
	n.commitConfirmCount = make(map[string]map[string]bool)
	n.isCommitBroadcast = make(map[string]bool)
	n.isReply = make(map[string]bool)
	n.msgLog = &MsgLog{
		make(map[string]map[string]bool),
		make(map[string]map[string]bool),
		make(map[string]map[string]bool),
		make(map[string]bool),
	}
	n.primary = ""
	n.totalPrimaryTable = totalPrimaryTable
	n.broadcastAddr = ""
	n.block = 0
	n.vicePrimary = ""
	return n
}

func (n *Node) Initiate(){
	hierarchy, primary, broadcastAddr := formLayer(n.nodeTable, n.nodeID, n.totalPrimaryTable)
	fmt.Println("results: ", hierarchy)
	fmt.Println("primary: ", primary)
	fmt.Println("broadcastAddr: ", broadcastAddr)
	n.nodeTable = hierarchy
	n.primary = primary
	n.broadcastAddr = broadcastAddr
	fmt.Println("NodeTable: " , n.nodeTable)
	fmt.Println("Primary Node: " , n.primary)
	fmt.Println("Broadcast Address: " , n.broadcastAddr)

	if n.nodeID == "N1" || n.nodeID == "N2" || n.nodeID == "N3" {//(3/7) add N4-N7
		n.vicePrimary = n.nodeID
	}
	
	go n.handleMsg()
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
		go n.handleConnection(conn)
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
	n.mutex.Lock()
	seq := n.sequenceID
	n.sequenceID++
	n.mutex.Unlock()
	return seq
}


func (n *Node) handleMsg(){
	for{
		data := <- n.msgQueue 
		//!NEW
		if n.block == 0{
			if n.nodeID == primary || n.nodeID == n.vicePrimary{
				//NEW! (20/6)
				if n.broadcastAddr != ""{
					send(data, n.broadcastAddr)
				}
				n.block = 1
			 }
		}
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
			n.block = 0
		}
	}
}

func (n *Node) handleRequest(payload []byte, sig []byte){
	r := new(RequestMsg)
	//convert json to struct format
	err := json.Unmarshal(payload, r)
	fmt.Println("request packet: " , r)
	if err != nil{
		log.Panic(err)
	}
	n.addSID()
	digest := createDigest(*r)
	digestmsg := generateDigest(r.CMessage.Request)
	//verify digest
	vdig := verifyDigest(digestmsg, r.CMessage.Digest)
	fmt.Println("verifydigest: ", vdig)
	if vdig == false{
		fmt.Printf("verify digest failed\n")
		return
	}
	strDigest := hex.EncodeToString(digest[:])
	//store map; digest = key, requestmsg = value
	n.requestPool[strDigest] = r
	//decode string to byte format for signing
	digestByte, _ := hex.DecodeString(strDigest)
	signature, err := signMessage(digestByte, n.privKey)
	if err != nil{
		log.Panic(err)
	}
	prePreparePacket := PrePrepareMsg{
		*r,
		viewID,
		n.sequenceID,
		strDigest,
		n.nodeID,
	}
	fmt.Println("sequenceID:(primary) " , n.sequenceID)
	//seq--
	//convert struct to json format
	done, err := json.Marshal(prePreparePacket)
	if err != nil{
		log.Panic(err)
	}
	message := mergeMsg(PrePrepare, done, signature)
	n.mutex.Lock()
	//put preprepare msg into preprepare log
	
	if n.msgLog.preprepareLog[prePreparePacket.Digest] == nil{
		n.msgLog.preprepareLog[prePreparePacket.Digest] = make(map[string]bool)
	}
	n.msgLog.preprepareLog[prePreparePacket.Digest][n.nodeID] = true
	
	n.mutex.Unlock()
	//update(3/7)
	if n.nodeID != "N1" && n.nodeID != "N2" && n.nodeID != "N3" {
		fmt.Println("broadcasting (pre-prepare) message..")
		n.broadcast(message)
	}
	n.sequenceID--
}

func (n *Node) handlePrePrepare(payload []byte, sig []byte){
	//update(23/6)
	if !isExist("./Penalties"){
		err := os.Mkdir("./Penalties", 0700)
		if err != nil{
			log.Panic(err)
		}
	}
	//create instance of preprepare
	pp := new(PrePrepareMsg)
	err := json.Unmarshal(payload, pp)
	if err != nil{
		log.Panic(err)
	}
	//get primary node's public key for verification
	primaryNodePubKey := getPubKey(n.primary)
	
	//decode string to byte format for signing
	digestByte, _ := hex.DecodeString(pp.Digest)
	//set approval conditions
	
	if digest := createDigest(pp.Request); hex.EncodeToString(digest[:]) != pp.Digest{
		//add penalty
		note := fmt.Sprintf("preprepare phase: digest (%s) not match, further application rejected!\n", pp.NodeID)
		n.recordPenalties(note)
	} else if n.sequenceID+1 != pp.SequenceID{
		//add penalty	
		note := fmt.Sprintf("preprepare phase: incorrect sequence (%d)(%d)(%s), further application rejected!\n", n.sequenceID+1, pp.SequenceID, pp.NodeID)
		n.recordPenalties(note)
    } else if !n.verifySignature(digestByte, sig, primaryNodePubKey){
		//add penalty
		note := fmt.Sprintf("preprepare phase: key verification (%s) failed, further application rejected!\n", pp.NodeID)
		n.recordPenalties(note)
	} else {
		//success
		fmt.Println("sequenceID: ", n.sequenceID)
		n.sequenceID = pp.SequenceID
		fmt.Println("preprepare phase: stored into message pool")
		n.requestPool[pp.Digest] = &pp.Request
		signature, err := signMessage(digestByte, n.privKey)
		if err != nil{
			log.Panic(err)
		}
		preparePacket := PrepareMsg{
			viewID,
			pp.SequenceID,
			pp.Digest, 
			n.nodeID, 
		}
		done, err := json.Marshal(preparePacket)
		if err != nil{
			log.Panic(err)
		}
		message := mergeMsg(Prepare, done, signature)
		//put prepare msg into prepare log
		n.mutex.Lock()
		if n.msgLog.prepareLog[preparePacket.Digest] == nil{
			n.msgLog.prepareLog[preparePacket.Digest] = make(map[string]bool)
		} 
		n.msgLog.prepareLog[preparePacket.Digest][n.nodeID] = true
		n.mutex.Unlock()
		fmt.Println("broadcasting (prepare) message..")
		//update (30/6)
		time.Sleep(nodeDelay)
		n.broadcast(message)
	}
}

func (n *Node) handlePrepare(payload []byte, sig []byte){
	pre := new(PrepareMsg)
	err := json.Unmarshal(payload, pre)
	if err != nil{
		log.Panic(err)
	}
	msgNodePubKey := getPubKey(pre.NodeID)
	//decode string to byte format 
	digestByte, _ := hex.DecodeString(pre.Digest)
	// if n.nodeID == n.primary || n.nodeID == n.vicePrimary{
	// 	n.sequenceID++
	// }
	fmt.Println("prepare request pool(pre-): ", n.requestPool[pre.Digest], "nodeID: ", pre.NodeID)
	if _, ok := n.requestPool[pre.Digest]; !ok{
		//add penalty
		note := fmt.Sprintf("prepare phase: unable to retrieve (%s) digest, further application rejected!\n", pre.NodeID)
		fmt.Println("prepare request pool: ", n.requestPool[pre.Digest])
		n.recordPenalties(note)
	} else if n.sequenceID != pre.SequenceID{
		//add penalty
		//update(24/6)
		if n.nodeID != n.primary && n.nodeID != n.vicePrimary{
			note := fmt.Sprintf("prepare phase: incorrect sequence (%s)(current sequenceID: %d)(received sequenceID: %d), further application rejected!\n", pre.NodeID, n.sequenceID, pre.SequenceID)
			n.recordPenalties(note)
		}
	} else if !n.verifySignature(digestByte, sig, msgNodePubKey){
		//add penalty
		note := fmt.Sprintf("prepare phase: key verification (%s) failed, further application rejected!\n", pre.NodeID)
		n.recordPenalties(note)
	} else{
		//success
		n.setPrepareConfirmMap(pre.Digest, pre.NodeID, true)
		count := 0
		for range n.prepareConfirmCount[pre.Digest]{
			count++
		}
		specifiedCount := 0
		fmt.Println("nodecount per group: " , newNodeCount)
		if n.nodeID == n.primary /* "N0" */{
			specifiedCount = newNodeCount / 3 * 2 //<--problem with nodeCount
		} else{
			specifiedCount = (newNodeCount / 3 * 2) -1
		}
		fmt.Println("specifiedCount: ", specifiedCount)
		if count >= specifiedCount && !n.isCommitBroadcast[pre.Digest]{
			fmt.Println("minimum (prepare) consensus achieved!")
			signature, err := signMessage(digestByte, n.privKey)
			if err != nil{
				log.Panic(err)
			}
			c := CommitMsg{
				viewID,
				pre.Digest,
				pre.SequenceID,
				n.nodeID,
			}
			done, err := json.Marshal(c)
			if err != nil{
				log.Panic(err)
			} 
			fmt.Println("broadcasting (commit) message..")
			
			message := mergeMsg(Commit, done, signature)
			//put commit msg into commit log
			n.mutex.Lock()
			if n.msgLog.commitLog[c.Digest] == nil{
				n.msgLog.commitLog[c.Digest] = make(map[string]bool)
			}
			n.msgLog.commitLog[c.Digest][n.nodeID] = true
			n.mutex.Unlock()

			time.Sleep(nodeDelay)
			n.broadcast(message)
			n.isCommitBroadcast[pre.Digest] = true
			fmt.Println("committed successfully")
			}
	}
}

func (n *Node) handleCommit(payload []byte, sig []byte){
	cmt := new(CommitMsg)
	err := json.Unmarshal(payload, cmt)
	if err != nil{
		log.Panic(err)
	}

	msgNodePubKey := getPubKey(cmt.NodeID)
	digestByte, _ := hex.DecodeString(cmt.Digest)

	if _, ok := n.prepareConfirmCount[cmt.Digest]; !ok{
		note := fmt.Sprintf("commit phase: unable to retrieve (%s) digest, further application rejected\n", cmt.NodeID)
		n.recordPenalties(note)
		//add penalty
	} else if n.sequenceID != cmt.SequenceID{
		note := fmt.Sprintf("commit phase: incorrect sequence (%s), further application rejected!\n", cmt.NodeID)
		n.recordPenalties(note)
		//add penalty
	} else if !n.verifySignature(digestByte, sig, msgNodePubKey){
		note := fmt.Sprintf("commit phase: key verification (%s) failed, further application rejected!\n", cmt.NodeID)
		n.recordPenalties(note)
		//add penalty
	} else{
		n.setCommitConfirmMap(cmt.Digest, cmt.NodeID, true)
		count := 0
		for range n.commitConfirmCount[cmt.Digest]{
			count++
		}
		n.mutex.Lock()
		if count >= newNodeCount / 3 * 2 && !n.isReply[cmt.Digest] && n.isCommitBroadcast[cmt.Digest]{
			fmt.Println("minimum (commit) consensus achieved!")

			signature, err := signMessage(digestByte, n.privKey)
			if err != nil{
				log.Panic(err)
			}

			d := ReplyMsg{
				viewID,
				int(time.Now().Unix()),
				n.nodeID,
				"success",
			}

			fmt.Println("reply message: ", d)
			done, err := json.Marshal(d)
			if err != nil{
				log.Panic(err)
			}

			fmt.Println("broadcasting results..")
			message := mergeMsg(Reply, done, signature)
			send(message, n.requestPool[cmt.Digest].CAddr)
			n.isReply[cmt.Digest] = true
			fmt.Println("successfully replied!")
		}
		n.mutex.Unlock()
	}
}

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

func (n *Node) broadcast(data []byte){
	for _, i := range n.nodeTable{
		if i != n.nodeID{
			fmt.Println("node address: " , i)
			send(data, i)
		}
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

func (n *Node) setPrepareConfirmMap(x, y string, b bool){
	if _, ok := n.prepareConfirmCount[x]; !ok{
		n.prepareConfirmCount[x] = make(map[string]bool)
		fmt.Println("prepareConfirmCount(x): ", x)
		fmt.Println("prepareConfirmCount(y): ", y)
	}
	n.prepareConfirmCount[x][y] = b
}

func (n *Node) setCommitConfirmMap(x, y string, b bool){
	if _, ok := n.commitConfirmCount[x]; !ok{
		n.commitConfirmCount[x] = make(map[string]bool)
	}
	n.commitConfirmCount[x][y] = b
}

 //update(23/6)
 func (n *Node) recordPenalties(s string){
	oflag := 0
	filename, _ := filepath.Abs(fmt.Sprintf("./Penalties/%s", n.nodeID))
	write := fmt.Sprintf(s)
	if !isExist(filename + "_penalty"){
		err := ioutil.WriteFile(filename + "_penalty", []byte(write), 0644)
		oflag = 1
		if err != nil{
			log.Panicf("penalty log: error creating file: %s", err)
		}
	}
	f, err :=os.OpenFile(filename + "_penalty", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil{
		log.Panicf("penalty log: error opening file: %s", err)
	}
	defer f.Close() 

	if oflag != 1{
		if _, err = f.WriteString(write); err != nil{
			log.Panicf("error recording results: %s", err)
		}
	}
	oflag = 0
 }