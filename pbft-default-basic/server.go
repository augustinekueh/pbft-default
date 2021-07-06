package main

import(
	"fmt"
)

var urlName = "localhost:%d"

type Server struct{
	node *Node
	url string
}

func nodeIdToPort(nodeId string)string{
	return nodeId
}

func newServer(nodeId , addr string, nodeTable map[string]string) *Server{
	server := &Server{
		newNode(string(nodeId), addr, nodeTable),
		fmt.Sprintf(urlName, nodeIdToPort(nodeId)),
	}
	return server
}

func (s *Server) Initiate(){
	s.node.Initiate()
}



