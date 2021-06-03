package main

import (
	"fmt"
	"strconv"
)
type Moderator struct{
	consensusTable		map[string]string
	nonconsenusTable	map[string]string
}


func calculateTrust(nodeTable map[string]string){
	fmt.Println("calculating node trust...")

	for i, _ := range nodeTable{
		strnum := string(i[1:])
		if num, err := strconv.Atoi(strnum); err == nil{
			fmt.Printf("Number : %d", num)
			if num % 2 == 0{
		}
		}
		
	}
}