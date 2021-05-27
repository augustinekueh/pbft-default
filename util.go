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

//TOTAL METHODS: 10
//hashcode

func createDigest(request RequestMsg) []byte{
	bmsg, err := json.Marshal(request)
	if err != nil{
		log.Panic(err)
	}
	hash := sha256.Sum256(bmsg)
	return hash[:]
}

func generateDigest(req string) []byte{
	bmsg, _ := json.Marshal(req)
	hash := sha256.Sum256(bmsg)
	return hash[:] 
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
	fmt.Println(data, addr)
	fmt.Println("breakpoint")
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

func verifyDigest(msg []byte, digest []byte) bool{
	fmt.Println(hex.EncodeToString(msg))
	fmt.Println(hex.EncodeToString(digest))
	return hex.EncodeToString(msg) == hex.EncodeToString(digest)
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

			//create public keys
			/*pubFileName := "Keys/N" + strconv.Itoa(i) + "/N" + strconv.Itoa(i) + "_RSA_PUB"
			pubFile, err := os.OpenFile(pubFileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil{
				log.Panic(err)
			}
			defer pubFile.Close()
			pubFile.Write(pub)

			//create private keys
			privFileName := "Keys/N" + strconv.Itoa(i) + "/N" + strconv.Itoa(i) + "_RSA_PRIV"
			privFile, err := os.OpenFile(privFileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil{
				log.Panic(err)
			}
			defer privFile.Close()
			privFile.Write(priv)
		}
		fmt.Println("all keys created successfully!")*/
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