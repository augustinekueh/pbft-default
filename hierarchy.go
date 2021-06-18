package main

import(
	"fmt"
	"sort"
	//"strconv"
)

//NEW LAYER STRUCTURE (PRO VERSION)
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
var assignFlag bool = false
var indexPosition int = 0
var indexPrimary int = 0 
var broadcastAddr string 

//!!bring in nodeID to make the comparison. then send the relevant grouptable back
func formLayer(nodeTable map[string]string, nodeID string, totalPrimaryTable map[string]string) (map[string]string, string, string){
	//have not initialized inner map		    level   group   node 
	fmt.Println("layering...")
	fmt.Println("nodeID: ", nodeID)
	fmt.Println("current nodetable: ", nodeTable)
	gnt := make(map[int]map[string]string)
	lnt := make(map[int]map[int]map[string]string)
	wnt := make(map[int]map[int]map[int]map[string]string)
	newNodeTable := make(map[string]string)
	//NEW!
	indexTotalPrimaryTable := make(map[int]string)
	//allocating slice for arrangement
	keys := make([]string, 0)
	pkeys := make([]string, 0)

	oneDigitKeys := make([]string, 0)
	twoDigitKeys := make([]string, 0)
	onePDigitKeys := make([]string, 0)
	twoPDigitKeys := make([]string, 0)
	threeDigitKeys := make([]string, 0)
	threePDigitKeys := make([]string, 0)
	b := 0  
	c := 0
	d := 0

	for a:= range nodeTable{
		//s := "N" + strconv.Itoa(a)
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

	for a:= range totalPrimaryTable{
		//s := "N" + strconv.Itoa(a)
		if len(a) == 2{
			onePDigitKeys = append(onePDigitKeys, a)
			b++
		} else if len(a) == 3{
			twoPDigitKeys = append(twoPDigitKeys, a)
			c++
		} else{
			threePDigitKeys = append(threePDigitKeys, a)
			d++
		}
	}

	sort.Strings(oneDigitKeys)
	sort.Strings(twoDigitKeys)
	sort.Strings(threeDigitKeys)

	sort.Strings(onePDigitKeys)
	sort.Strings(twoPDigitKeys)
	sort.Strings(threePDigitKeys)

	fmt.Println("oneDigitArrays: ", oneDigitKeys)
	fmt.Println("twoDigitArrays: ", twoDigitKeys)
	fmt.Println("threeDigitArrays: ", threeDigitKeys)

	keys = append(keys, oneDigitKeys...)
	fmt.Println("first append: ", keys)
	keys = append(keys, twoDigitKeys...)
	keys = append(keys, threeDigitKeys...)

	//!NEW
	pkeys = append(pkeys, onePDigitKeys...)
	fmt.Println("first append: ", pkeys)
	pkeys = append(pkeys, twoPDigitKeys...)
	pkeys = append(pkeys, threePDigitKeys...)

	fmt.Println("sorted nodetable: ", keys)
	fmt.Println("sorted primary table: ", pkeys)

	for _, a := range pkeys{
		indexTotalPrimaryTable[indexPrimary] = a 
		indexPrimary++
	}

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
					//for fixing the broadcasting (preprepare packet) of top-level primary nodes, but currently not much use
					// if assignFlag == false{
					// 	primary = nodeID
					// } else{
					// 	primary = keyArr[0]
					// }
					primary = keyArr[0]
					done = true
					match = false 
					
				}
			}
			if(assignFlag){
				if indexTotalPrimaryTable[indexPosition] == nodeID{
					broadcastAddr = valArr[0] 
					fmt.Println("Broadcast Address (Hierarchy): ", valArr[0])
					//put a number instead of keyArr[p] and put zero instead of p for valArr??
				}
				indexPosition++
			}
			keyArr = nil
			valArr = nil
			assignFlag = true
		} 
		if index == 4 {
			index = 0
			wnt[whole] = make(map[int]map[int]map[string]string)
			wnt[whole] = lnt
			whole++
		} 
	}
}
	assignFlag = false
	indexPosition = 0
	return newNodeTable, primary, broadcastAddr
}