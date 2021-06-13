package main

import(
	"fmt"
	"sort"
	//"strconv"
)

var count int = 0
var index int = 0
var whole int = 0
var match bool = false
var done bool = false
var newNodeCount int
var twoInt int = 0

var keyArr []string
var valArr []string
var primary string

//!!bring in nodeID to make the comparison. then send the relevant grouptable back
func formLayer(nodeTable map[string]string, nodeID string) (map[string]string, string){
	//have not initialized inner map		    level   group   node 
	fmt.Println("layering...")
	fmt.Println("nodeID: ", nodeID)
	fmt.Println("current nodetable: ", nodeTable)
	gnt := make(map[int]map[string]string)
	lnt := make(map[int]map[int]map[string]string)
	wnt := make(map[int]map[int]map[int]map[string]string)
	newNodeTable := make(map[string]string)
	//allocating slice for arrangement
	keys := make([]string, 0)
	oneDigitKeys := make([]string, 0)
	twoDigitKeys := make([]string, 0)
	b := 0  
	c := 0

	for a:= range nodeTable{
		//s := "N" + strconv.Itoa(a)
		if len(a) == 2{
			oneDigitKeys = append(oneDigitKeys, a)
			b++
		} else{
			twoDigitKeys = append(twoDigitKeys, a)
			c++
		}  
	}
	sort.Strings(oneDigitKeys)
	sort.Strings(twoDigitKeys)

	fmt.Println("oneDigitArrays: ", oneDigitKeys)
	fmt.Println("twoDigitArrays: ", twoDigitKeys)

	keys = append(keys, oneDigitKeys...)
	fmt.Println("first append: ", keys)
	keys = append(keys, twoDigitKeys...)

	fmt.Println("sorted nodetable: ", keys)

	for _, a := range keys{//<-- problem here; ordering issue
		if a != "C0"{
		if count == 0 && !done{
			newNodeTable = make(map[string]string)
		}
		//initialized gnt's inner map
		gnt[count] = make(map[string]string)
		gnt[count][a] = nodeTable[a] 

		keyArr = append(keyArr, a)
		valArr = append(valArr, nodeTable[a])

		if(a == nodeID){ 
			match = true
		}
		
		count++
		if count == 4 {
			count = 0
			//newNodeTable := make(map[string]string)
			lnt[index] = make(map[int]map[string]string)
			lnt[index] = gnt
			index++
			if(match){//problem if add more than 4 nodes
				//fmt.Println("breakpoint3")
				for p := 0; p <= 3; p++{
					//fmt.Println(keyArr[p])//arrangement got problem, probably need to do sorting
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
		if index == 4 {
			index = 0
			wnt[whole] = make(map[int]map[int]map[string]string)
			wnt[whole] = lnt
			whole++
		} 
	}
}
	return newNodeTable, primary
}