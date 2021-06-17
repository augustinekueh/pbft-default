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
	jsonFile, err := os.Open("node_64.json")

	if err != nil{
		fmt.Println(err)
	}

	fmt.Println("Successfully opened node.json")

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	nodeTable := make(map[string]string)
	lowerPrimaryTable := make(map[string]string)
	firstPrimaryTable := make(map[string]string)
	totalPrimaryTable := make(map[string]string)
	json.Unmarshal(byteValue, &jnodes)
	count := 0
	tmpCount := 0
	flag := 0

	for i:=0; i<len(jnodes.JsonNodes); i++{
		nodeTable[jnodes.JsonNodes[i].ID] = jnodes.JsonNodes[i].URL
		if count == 0 || tmpCount <= 3{
			if jnodes.JsonNodes[i].ID != "C0"{
				if flag >= 4{
					lowerPrimaryTable[jnodes.JsonNodes[i].ID] = jnodes.JsonNodes[i].URL
				}
				if tmpCount <= 3{
					firstPrimaryTable[jnodes.JsonNodes[i].ID] = jnodes.JsonNodes[i].URL
					tmpCount++
					flag++
				}
			}
		}
		count++
		if count == 4{//change to determine number of primary nodes; now is only 1 if run the node_8.json
			count = 0
		}
	}
	
	for k, v := range firstPrimaryTable{
		totalPrimaryTable[k] = v
	} 

	for k, v := range lowerPrimaryTable{
		totalPrimaryTable[k] = v
	} 

	for k,v := range nodeTable{
		fmt.Println("Key: ", k, "Value: ", v)
		nodeCount++
	}
	
	fmt.Println("First Primary Table: ", firstPrimaryTable)
	fmt.Println("Lower Primary Table: ", lowerPrimaryTable)
	fmt.Println("Total Primary Table: ", totalPrimaryTable)

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
		client := newClient(termID, addr, firstPrimaryTable)
		client.Initiate()
		} else{
			log.Fatal("connection failed!")
		}
	//node
	} else if addr, ok := nodeTable[termID]; ok{	
		server := newServer(termID, addr, nodeTable, totalPrimaryTable)
		server.Initiate()
	} else {
		log.Fatal("connection failed!!")
	}
	select {}
}

func main(){
	Operation()
}
