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
var jnodes *JsonNodes
var ClientNode *JsonNode

//create slices of JsonNode
type JsonNodes struct{
	JsonNodes []JsonNode `json:"nodes"`
}

//structure of a json node
type JsonNode struct{
	ID string `json:"id"`
	URL string `json:"url"`
}

func Operation(){
	fmt.Println("Hello World")

	//Retrieve nodes' information from a json file
	//change the total network nodes based on the numbering json file; i.e. nodes = 4 
	jsonFile, err := os.Open("node_16.json")

	if err != nil{
		fmt.Println(err)
	}

	fmt.Println("Successfully opened node.json")

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	nodeTable := make(map[string]string)
	primaryTable := make(map[string]string)
	json.Unmarshal(byteValue, &jnodes)

	for i:=0; i<len(jnodes.JsonNodes); i++{
		nodeTable[jnodes.JsonNodes[i].ID] = jnodes.JsonNodes[i].URL
	
		if jnodes.JsonNodes[i].ID == "N0"{
				primaryTable[jnodes.JsonNodes[i].ID] = jnodes.JsonNodes[i].URL
		}	
	}
	fmt.Println("Primary Table", primaryTable)
	for k,v := range nodeTable{
		fmt.Println("Key: ", k, "Value: ", v)
		nodeCount++
	}
	
	//pre-generate RSA keys; public and private 
	genKeys(len(jnodes.JsonNodes))
	
	//terminal input condition
	if len(os.Args)!=2{
		log.Panic("command insertion error!")
	}

	termID := os.Args[1]
	//client
	if termID == "C0"{
		if addr, ok := nodeTable[termID]; ok{
		//fmt.Println(addr)
		ClientNode = &JsonNode{
			termID,
			nodeTable[termID],
		}
		client := newClient(termID, addr, primaryTable)
		client.Initiate()
		} else{
			log.Fatal("connection failed!")
		}
	//node
	} else if addr, ok := nodeTable[termID]; ok{	
		server := newServer(termID, addr, nodeTable)
		server.Initiate()
	} else {
		log.Fatal("connection failed!!")
	}
	select {}
}

func main(){
	Operation()
}