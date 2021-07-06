package main

import(
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	//"strconv"
)

func createDigest(request RequestMsg) []byte{
	bmsg, err := json.Marshal(request)
	if err != nil{
		log.Panic(err)
	}
	hash := sha256.Sum256(bmsg)
	return hash[:]
}

func generateDigest(req string) []byte{
	bmsg, err := json.Marshal(req)
	if err != nil{
		log.Panic(err)
	}
	hash := sha256.Sum256(bmsg)
	return hash[:] 
}

func verifyDigest(msg []byte, digest []byte) bool{
	fmt.Println(hex.EncodeToString(msg))
	fmt.Println(hex.EncodeToString(digest))
	return hex.EncodeToString(msg) == hex.EncodeToString(digest)
}

//sign message using a private key
func signMessage(data []byte, keyBytes []byte) ([]byte, error){
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil{
		panic(errors.New("private key error"))
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil{
		fmt.Println("ParsePKCS1PrivateKey err", err)
		panic(err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil{
		fmt.Printf("Error from signing: %s\n", err)
		panic(err)
	}
	return signature, err
}

//verify signature using a public key
func (n *Node) verifySignature(data, sig, keyBytes []byte) bool{
	block, _ := pem.Decode(keyBytes)
	if block == nil{
		panic(errors.New("public key error"))
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil{
		panic(err)
	}

	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], sig)
	if err != nil{
		panic(err)
	}
	return true
}

func send(data []byte, addr string){
	conn, err := net.Dial("tcp", addr)
	if err != nil{
		log.Println("connect error:", err)
		return 
	}
	_, err = conn.Write(data)
	if err != nil{
		log.Fatal(err)
	}
	conn.Close()
}

func countTotalFaultNodes() int{
	return (nodeCount-1) / 3
}

func countTotalMsgAmount() int{
	f := countTotalFaultNodes()
	return f + 1
}

//generate keys beforehand 
func genKeys(nodes int){
	if !isExist("./Keys"){
		fmt.Println("creating public and private keys...")
		fmt.Printf("Total nodes: %d\n", nodes)
		err := os.Mkdir("Keys", 0700)
		if err != nil{
			log.Panic()
		}

		//for client
		clientFile, _ := filepath.Abs("./Keys/C0")
		if !isExist(clientFile + "_priv") && !isExist(clientFile + "_pub"){
			pub, priv := genPair()
			err := ioutil.WriteFile(clientFile + "_priv", priv, 0644)
			if err != nil{
				panic(err)
			}
			ioutil.WriteFile(clientFile + "_pub", pub, 0644)
			if err != nil{
				panic(err)
			}
		}

		//make directories for keys
		for i :=0; i < nodes; i++{
			filename, _ := filepath.Abs(fmt.Sprintf("./Keys/N%d", i))
			if !isExist(filename + "_priv") && !isExist(filename + "_pub"){
				pub, priv := genPair()
				err := ioutil.WriteFile(filename + "_priv", priv, 0644)
				if err != nil{
					panic(err)
				}
				ioutil.WriteFile(filename + "_pub", pub, 0644)
				if err != nil{
					panic(err)
				}
			}
		}
	}
}
//sub-method from genRSAKeys
func genPair() (pubKey, privKey []byte){

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil{
		panic(err)
	}

	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := &pem.Block{
		Type: "PRIVATE KEY",
		Bytes: derStream,
	}

	privKey = pem.EncodeToMemory(privBlock)
	publicKey := &privateKey.PublicKey
	
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil{
		panic(err)
	}
	pubBlock := &pem.Block{
		Type: "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubKey = pem.EncodeToMemory(pubBlock)
	return pubKey, privKey
}

func getPubKey(memberID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + memberID + "_pub")

	if err != nil{
		log.Panic(err)
	}
	return key
}

func getPrivKey(memberID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + memberID + "_priv")

	if err != nil{
		log.Panic(err)
	}
	return key
}

//search for filepath and return a bool
func isExist(path string) bool{
	_, err := os.Stat(path)
	if err != nil{
		if os.IsExist(err){
			return true
		}
		if os.IsNotExist(err){
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

// func updateNodeTable(nodeTable map[string]string)map[string]string{
// 	fmt.Println("calculating node trust...")

// 	consensusTable := make(map[string]string)
// 	for k, v := range nodeTable{
// 		if k != "C0"{
// 			strnum := string(k[1:])
// 			if num, err := strconv.Atoi(strnum); err == nil{
// 				fmt.Printf("Consensus Nodes: N%d\n", num)
// 				if num % 2 == 0{
// 					consensusTable[k] = v
// 				} 
// 			} else{
// 				log.Panic(err)
// 			}
// 		}
// 	}
// 	return consensusTable
// }