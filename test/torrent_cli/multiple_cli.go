/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main 

import (
	//"path"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"fmt"

	"github.com/anacrolix/dht"
	"github.com/anacrolix/torrent"
	"github.com/radovskyb/watcher"
	"log"
	"bytes"
)

const (
	origindataPath    = "origindata"
	encryptdataPath   = "encryptdata"
	decryptdataPath   = "decryptdata"
	org1        = "Org1"
	org2        = "Org2"
)

// Peers
var orgTestPeer0 fab.Peer
var orgTestPeer1 fab.Peer

// TestOrgsEndToEnd creates a channel with two organisations, installs chaincode
// on each of them, and finally invokes a transaction on an org2 peer and queries
// the result from an org1 peer
func main() {

	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile("config_test.yaml"))
	if err != nil {
		fmt.Println("Failed to create new SDK: %s", err)
	}

	// Channel management torrentClient is responsible for managing channels (create/update channel)
	//chMgmtClient, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg("ordererorg")).ChannelMgmt()
	//if err != nil {
	//	fmt.Println(err)
	//}

	// Create channel (or update if it already exists)
	//org1AdminUser := loadOrgUser( sdk, org1, "Admin")
	//req := chmgmt.SaveChannelRequest{ChannelID: "orgchannel", ChannelConfig: path.Join("v1.1/channel/", "orgchannel.tx"), SigningIdentity: org1AdminUser}
	//if err = chMgmtClient.SaveChannel(req); err != nil {
	//	fmt.Println(err)
	//}

	// Allow orderer to process channel creation
	time.Sleep(time.Second * 5)

	// Load specific targets for move funds test
	loadOrgPeers( sdk)

	clientConfig := torrent.Config{}
	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := sdk.NewClient(fabsdk.WithUser("User1"), fabsdk.WithOrg(org1)).Channel("orgchannel")
	if err != nil {
		fmt.Println("Failed to create new channel torrentClient for Org1 user: %s", err)
	}
	for{
		upload_response, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "dht_server", Fcn: "invoke", Args: dht_queryArgs})
		if err !=nil || string(upload_response.Payload)=="" || upload_response.Payload==nil{
			fmt.Println("another try in getting server address")
			time.Sleep(20*time.Second)
		}else{
			clientConfig.DHTConfig = dht.ServerConfig{
				StartingNodes:generateClientAddrs([]string {string(upload_response.Payload)}),
			}
			fmt.Println("finally get the server address: "+string(upload_response.Payload))
			break
		}
	}

	clientConfig.Seed = true
	clientConfig.Debug = true
	clientConfig.DisableTrackers = true
	clientConfig.ListenAddr = "0.0.0.0:6666"
	clientConfig.DataDir = encryptdataPath
	clientConfig.DisableAggressiveUpload = false
	torrentClient, _ := torrent.NewClient(&clientConfig)

	go testChaincodeEventListener("myapp",chClientOrg1User, torrentClient)
	//retrive all magnet available
	upload_response, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "myapp", Fcn: "getAllMagnet"})
	if err!=nil{
	}
	allMagnet:=bytes.Split(upload_response.Payload,[]byte(",,"))
	for _,i :=range(allMagnet){
		fmt.Println(i)
		download(torrentClient,string(i))
	}
	//dir, _ := os.Open(dataPath)
	//defer dir.Close()
	//
	//fi, _ := dir.Readdir(-1)
	//for _, x := range fi {
	//	if !x.IsDir() && x.Name() != ".torrent.bolt.db" {
	//		d := makeMagnet(dataPath, x.Name(), torrentClient)
	//		fmt.Println(d)
	//		upload_AddArgs := [][]byte{[]byte(x.Name()),[]byte("hash"),[]byte("keywords"),[]byte("Summary"),[]byte(d)}
	//		_, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "myapp", Fcn: "createFile", Args:upload_AddArgs})
	//		if err != nil {
	//			fmt.Println("Failed to add a magnetlink: %s", err)
	//		}
	//	}
	//}

	time.Sleep(time.Second * 5)
	//replace start
/*
	upload_response, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "upload", Fcn: "invoke", Args:upload_QueryArgs})
	available_magnets :=strings.Split(string(upload_response.Payload),",")
	fmt.Printf("%q\n",available_magnets)

	for _,v := range available_magnets {
		fmt.Println(v)
		download(torrentClient,v)
	}
*/
	// replace end
	w :=watcher.New()
	go func(){
		for{
			select {
			case event:=<-w.Event:
				if event.Op.String()=="CREATE"{
					key,err:=encryptFile(event.Name())
					if err!=nil{
						log.Fatalln("err in encrypt file")
						return
					}
					d:=makeMagnet(encryptdataPath, event.Name(),torrentClient)
					fmt.Println(d)
					upload_AddArgs := [][]byte{[]byte(event.Name()),[]byte("hash"),[]byte("keywords"),[]byte("Summary"),[]byte(d),[]byte(key)}
					response, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "myapp", Fcn: "createFile", Args:upload_AddArgs})
					if err != nil {
						fmt.Println("Failed to add a magnetlink: %s", err)
					}else{
						fmt.Println("username : ",string(response.Payload))
					}
				}

			case err:=<-w.Error:
				log.Fatal(err)
			case <-w.Closed:
				return
			}

		}
	}()
	if err:=w.AddRecursive(origindataPath);err!=nil{
		log.Fatalln(err)
	}
	if err:=w.Start(time.Millisecond*100);err!=nil{
		log.Fatalln(err)
	}

	select {}
}

func loadOrgPeers( sdk *fabsdk.FabricSDK) {

	org1Peers, err := sdk.Config().PeersConfig(org1)
	if err != nil {
		fmt.Println(err)
	}

	org2Peers, err := sdk.Config().PeersConfig(org2)
	if err != nil {
		fmt.Println(err)
	}

	orgTestPeer0, err = peer.New(sdk.Config(), peer.FromPeerConfig(&apiconfig.NetworkPeer{PeerConfig: org1Peers[0]}))
	if err != nil {
		fmt.Println(err)
	}

	orgTestPeer1, err = peer.New(sdk.Config(), peer.FromPeerConfig(&apiconfig.NetworkPeer{PeerConfig: org2Peers[0]}))
	if err != nil {
		fmt.Println(err)
	}
}

var dhtserver_Initargs= [][]byte{[]byte("init"), []byte("dht_server"), []byte("server:6666")}
var dht_queryArgs = [][]byte{[]byte("query"), []byte("dht_server")}

var upload_InitArgs = [][]byte{[]byte("init"),[]byte("init"),[]byte("myipaddr:port")}
var upload_QueryArgs = [][]byte{[]byte("query"), []byte("init")}
