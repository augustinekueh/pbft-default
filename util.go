package main

import(
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

//rewrite
func createDigest(request RequestMsg) []byte{
	bmsg, err := json.Marshal(request)
	if err != nil{
		log.Panic(err)
	}
	hash := sha256.Sum256(bmsg)
	return hash[:]
}

//sign message using a private key
func (n *Node) signMessage(data []byte, keyBytes []byte) ([]byte, error){
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

func send(data []byte, addr string){
	conn, err := net.Dial("tcp", addr)
	if err != nil{
		log.Println("connect error", err)
		return 
	}
	_, err = conn.Write(data)
	if err != nil{
		log.Fatal(err)
	}
	conn.Close()
}

/*func countTotalFaultNodes() int{
	
}
*/

func countTotalMsgAmount() int{
	//f := countToleratefaultNode()
	return 0
}

/*
func verifyDigest(msg interface{}, digest string) bool{
	return hex.EncodeToString(createDigest(msg)) == digest
}
*/

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


func genRSAkeys(nodes int){
	if !isExist("./Keys"){
		fmt.Println("creating public and private keys...")
		err := os.Mkdir("Keys", 0644)
		if err != nil{
			log.Panic()
		}

		//make directories for keys
		for i :=0; i <= nodes; i++{
			if !isExist("./Keys/N" + strconv.Itoa(i)){
				err := os.Mkdir("./Keys/N" + strconv.Itoa(i), 0644)
				if err != nil{
					log.Panic()
				}
			}

			//create public keys
			pub, priv := getKeyPair()
			pubFileName := "Keys/N" + strconv.Itoa(i) + "/N" + strconv.Itoa(i) + "_RSA_PUB"
			pubFile, err := os.OpenFile(pubFileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil{
				log.Panic(err)
			}
			defer pubFile.Close()
			pubFile.Write(pub)

			privFileName := "Keys/N" + strconv.Itoa(i) + "/N" + strconv.Itoa(i) + "_RSA_PRIV"
			privFile, err := os.OpenFile(privFileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil{
				log.Panic(err)
			}
			defer privFile.Close()
			privFile.Write(priv)
		}
		fmt.Println("all keys created successfully!")
	}
}


func getKeyPair() (pubKey, privKey []byte){

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
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
	pubBlock := &pem.Block{
		Type: "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubKey = pem.EncodeToMemory(pubBlock)
	return
}

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