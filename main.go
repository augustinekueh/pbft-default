package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"log"
)

//global variables
var nodeCount int
//var nodeTable map[string]string
var jnodes *JsonNodes
var ClientNode *JsonNode
//var PrimaryNode *JsonNode

//create slices of JsonNode
type JsonNodes struct{
	JsonNodes []JsonNode `json:"nodes"`
}

//structure of a json node
type JsonNode struct{
	ID string `json:"id"`
	URL string `json:"url"`
}

func main(){
	fmt.Println("Hello World")

	//Retrieve nodes' information from a json file
	jsonFile, err := os.Open("node.json")

	if err != nil{
		fmt.Println(err)
	}

	fmt.Println("Successfully opened node.json")

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	nodeTable := make(map[string]string)
	json.Unmarshal(byteValue, &jnodes)

	for i:=0; i<len(jnodes.JsonNodes); i++{
		nodeTable[jnodes.JsonNodes[i].ID] = jnodes.JsonNodes[i].URL
	}

	for k,v := range nodeTable{
		fmt.Println(k, "value is", v)
		nodeCount++
	}
	
	//pre-generate RSA keys; public and private 
	genKeys(len(jnodes.JsonNodes))
	fmt.Println("endl")
	//terminal input condition
	
	if len(os.Args)!=2{
		log.Panic("command insertion error!")
	}

	termID := os.Args[1]
	//client
	if termID == "C0"{
		if addr, ok := nodeTable[termID]; ok{
		fmt.Println(addr)
		ClientNode = &JsonNode{
			termID,
			nodeTable[termID],
		}
		client := newClient(termID, addr)
		client.Initiate()
		} else{
			fmt.Println(termID)
			fmt.Println(nodeTable[termID])
			log.Fatal("connection failed!")
		}
	//node
	} else if addr, ok := nodeTable[termID]; ok{	
		fmt.Println(addr)
		server := newServer(termID, addr, nodeTable)
		server.Initiate()
		//server.memberNodes()
	} else {
		fmt.Println(termID)
		fmt.Println(nodeTable[termID])
		log.Fatal("connection failed!!")
	}
	select {}
}