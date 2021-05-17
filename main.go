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
var nodeTable map[string]string
var jnodes JsonNodes

//create slices of JsonNode
type JsonNodes struct{
	JsonNodes []JsonNode `json:"jnodes"`
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

	//terminal input condition
	if len(os.Args)!=2{
		log.Panic("command insertion error!")
	}

	termID := os.Args[1]
	//client
	if termID == "C1"{
		if addr, ok := nodeTable[termID]; ok{
		client := newClient(termID, addr)
		client.Initiate()
		} else{
			log.Fatal("connection failed!")
		}
	//node
	} else if addr, ok := nodeTable[termID]; ok{
		node := newNode(termID, addr)
		go node.Initiate()
	} else {
		log.Fatal("connection failed!")
	}
	select {}
}