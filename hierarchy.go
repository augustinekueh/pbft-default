package main

import(
	"fmt"
)

var count int = 0
var index int = 0
var whole int = 0

//!!bring in nodeID to make the comparison. then send the relevant grouptable back
func formLayer(nodeTable map[string]string) map[int]map[int]map[int]map[string]string{
	//have not initialized inner map		    level   group   node 
	fmt.Println("layering...")
	fmt.Println(nodeTable)
	gnt := make(map[int]map[string]string)
	lnt := make(map[int]map[int]map[string]string)
	wnt := make(map[int]map[int]map[int]map[string]string)


	for k, v := range nodeTable{
		//initialized gnt's inner map
		gnt[count] = make(map[string]string)
		fmt.Println(count)
		gnt[count][k] = v	
		fmt.Println(gnt)
		count++
		if count == 4 {
			count = 0
			lnt[index] = make(map[int]map[string]string)
			lnt[index] = gnt
			index++
			/*if(nodeID == gnt[count] or k){
				nodeTable = gnt
				return nodeTable
			}*/
		} 
		if index == 4 {
			index = 0
			wnt[whole] = make(map[int]map[int]map[string]string)
			wnt[whole] = lnt
			whole++
		} 
	}
	return wnt
}