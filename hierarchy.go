package main

import(
	"fmt"
	"sort"
)

var count int = 0
var index int = 0
var whole int = 0
var match bool = false
var done bool = false
var newNodeCount int
//var lnds LayerNodes
//var newNodeTable map[string]string

var keyArr []string
var valArr []string
var primary string

// type LayerNodes struct{
// 	LayerNodes []LayerNode `json:"nodes"`
// }

// type LayerNode struct{
// 	ID string `json:"id"`
// 	URL string `json:"url"`
// }

// func (n nodeTable) sortMe()(arrange []string){
// 	for a, _ := range nodeTable{
// 		arrange = append(arrange, a)
// 	} 
// 	sort.Strings(arrange)
// 	return
// }

//!!bring in nodeID to make the comparison. then send the relevant grouptable back
func formLayer(nodeTable map[string]string, nodeID string) (map[string]string, string){
	//have not initialized inner map		    level   group   node 
	fmt.Println("layering...")
	fmt.Println("nodeID: ", nodeID)
	fmt.Println(nodeTable)
	gnt := make(map[int]map[string]string)
	lnt := make(map[int]map[int]map[string]string)
	wnt := make(map[int]map[int]map[int]map[string]string)
	newNodeTable := make(map[string]string)
	keys := make([]string, len(nodeTable))
	b := 0  
	for a:= range nodeTable{
		keys[b] = a
		b++
	}
	sort.Strings(keys)

	for _, a := range keys{//<-- problem here; ordering issue
		//fmt.Println("ididid: ", a)
		if a != "C0"{
		if count == 0 && !done{
			newNodeTable = make(map[string]string)
		}
		//initialized gnt's inner map
		gnt[count] = make(map[string]string)
		//fmt.Println(count)
		gnt[count][a] = nodeTable[a] 
		//fmt.Println("breakpoint")
		keyArr = append(keyArr, a)
		valArr = append(valArr, nodeTable[a])
		//sort.Strings(keyArr)
		//sort.Strings(valArr)
		//fmt.Println("breakpoint2")
		//fmt.Println(gnt)
		if(a == nodeID){ 
			//temp := count
			match = true
		}
		//fmt.Println(keyArr)
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