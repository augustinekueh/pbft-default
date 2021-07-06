package main

import(
	"fmt"
	"sort"
)

//global variables
var newNodeCount int
var keyArr []string
var valArr []string
var primary string

//!!bring in nodeID to make the comparison. then send the relevant grouptable back
func formLayer(nodeTable map[string]string, nodeID string) (map[string]string, string){
	fmt.Println("grouping...")
	fmt.Println("nodeID: ", nodeID)
	fmt.Println("current nodetable: ", nodeTable)
	newNodeTable := make(map[string]string)
	//allocating slice for arrangement
	keys := make([]string, 0)
	oneDigitKeys := make([]string, 0)
	twoDigitKeys := make([]string, 0)
	threeDigitKeys := make([]string, 0)
	
	b := 0  
	c := 0
	d := 0

	count :=0
	match := false
	done  := false

	for a:= range nodeTable{
		if len(a) == 2{
			oneDigitKeys = append(oneDigitKeys, a)
			b++
		} else if len(a) == 3{
			twoDigitKeys = append(twoDigitKeys, a)
			c++
		} else{
			threeDigitKeys = append(threeDigitKeys, a)
			d++
		}
	}
	sort.Strings(oneDigitKeys)
	sort.Strings(twoDigitKeys)
	sort.Strings(threeDigitKeys)

	fmt.Println("oneDigitArrays: ", oneDigitKeys)
	fmt.Println("twoDigitArrays: ", twoDigitKeys)
	fmt.Println("threeDigitArrays: ", threeDigitKeys)

	keys = append(keys, oneDigitKeys...)
	fmt.Println("first append: ", keys)
	keys = append(keys, twoDigitKeys...)
	keys = append(keys, threeDigitKeys...)

	fmt.Println("sorted nodetable: ", keys)

	for _, a := range keys{
		if a != "C0"{
		if count == 0 && !done{
			newNodeTable = make(map[string]string)
		}
		
		keyArr = append(keyArr, a)
		valArr = append(valArr, nodeTable[a])

		if(a == nodeID){ 
			match = true
		}
		
		count++
		if count == 4 {
			count = 0
			if(match){
				for p := 0; p <= 3; p++{
					newNodeTable[keyArr[p]] = valArr[p]
					newNodeCount++
					primary = keyArr[0]
					done = true
					match = false 
				}
			}
			keyArr = nil
			valArr = nil
		} 
	}
}
	return newNodeTable, primary
}