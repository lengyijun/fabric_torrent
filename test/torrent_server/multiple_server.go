/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main 

import (
	"path"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"fmt"
	"os"

	"github.com/anacrolix/dht"
	"github.com/anacrolix/torrent"
	"github.com/radovskyb/watcher"
	"log"
)

const (
	origindataPath    = "origindata"
	encryptdataPath   = "encryptdata"
	decryptdataPath   = "decryptdata"
	org1        	  = "Org1"
	org2              = "Org2"
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

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg("ordererorg")).ChannelMgmt()
	if err != nil {
		fmt.Println(err)
	}

	// Create channel (or update if it already exists)
	org1AdminUser := loadOrgUser( sdk, org1, "Admin")
	req := chmgmt.SaveChannelRequest{ChannelID: "orgchannel", ChannelConfig: path.Join("v1.1/channel/", "orgchannel.tx"), SigningIdentity: org1AdminUser}
	if err = chMgmtClient.SaveChannel(req); err != nil {
		fmt.Println(err)
	}

	// Allow orderer to process channel creation
	time.Sleep(time.Second * 5)

	// Org1 resource management client (Org1 is default org)
	org1ResMgmt, err := sdk.NewClient(fabsdk.WithUser("Admin")).ResourceMgmt()
	if err != nil {
		fmt.Println("Failed to create new resource management client: %s", err)
	}

	// Org1 peers join channel
	if err = org1ResMgmt.JoinChannel("orgchannel"); err != nil {
		fmt.Println("Org1 peers failed to JoinChannel: %s", err)
	}

	// Org2 resource management client
	org2ResMgmt, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg(org2)).ResourceMgmt()
	if err != nil {
		fmt.Println(err)
	}

	// Org2 peers join channel
	if err = org2ResMgmt.JoinChannel("orgchannel"); err != nil {
		fmt.Println("Org2 peers failed to JoinChannel: %s", err)
	}

	dhtPkg, err := packager.NewCCPackage("github.com/dht_server", "../fixtures/testdata")
	if err != nil {
		fmt.Println(err)
	}

	installCCReq := resmgmt.InstallCCRequest{Name: "dht_server", Path: "github.com/dht_server", Version: "0", Package:dhtPkg}

	_, err = org1ResMgmt.InstallCC(installCCReq)
	if err != nil {
		fmt.Println(err)
	}

	_, err = org2ResMgmt.InstallCC(installCCReq)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Second * 5)

	loadPkg, err := packager.NewCCPackage("github.com/myapp", "../fixtures/testdata")
	if err != nil {
		fmt.Println(err)
	}
	installUploadReq := resmgmt.InstallCCRequest{Name: "myapp", Path: "github.com/myapp", Version: "0", Package:loadPkg}

	_, err = org1ResMgmt.InstallCC(installUploadReq)
	if err != nil {
		fmt.Println(err)
	}

	_, err = org2ResMgmt.InstallCC(installUploadReq)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Second * 5)

	loadPkg, err = packager.NewCCPackage("github.com/keyExchange", "../fixtures/testdata")
	if err != nil {
		fmt.Println(err)
	}
	installUploadReq = resmgmt.InstallCCRequest{Name: "keyExchange", Path: "github.com/keyExchange", Version: "0", Package:loadPkg}

	_, err = org1ResMgmt.InstallCC(installUploadReq)
	if err != nil {
		fmt.Println(err)
	}

	_, err = org2ResMgmt.InstallCC(installUploadReq)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Second * 5)

	// Set up chaincode policy to 'any of two msps'
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})

	err = org1ResMgmt.InstantiateCC("orgchannel", resmgmt.InstantiateCCRequest{Name: "dht_server", Path: "github.com/dht_server", Version: "0", Args:DhtServerInitArgs(), Policy: ccPolicy})
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Second * 5)

	err = org1ResMgmt.InstantiateCC("orgchannel", resmgmt.InstantiateCCRequest{Name: "myapp", Path: "github.com/myapp", Version: "0", Args:upload_InitArgs, Policy: ccPolicy})
	if err != nil {
		fmt.Println(err)
	}

	err = org1ResMgmt.InstantiateCC("orgchannel", resmgmt.InstantiateCCRequest{Name: "keyExchange", Path: "github.com/keyExchange", Version: "0", Args:[][]byte{}, Policy: ccPolicy})
	if err != nil {
		fmt.Println(err)
	}
	// Load specific targets for move funds test
	loadOrgPeers( sdk)

	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := sdk.NewClient(fabsdk.WithUser("User1"), fabsdk.WithOrg(org1)).Channel("orgchannel")
	if err != nil {
		fmt.Println("Failed to create new channel client for Org1 user: %s", err)
	}

	clientConfig := torrent.Config{}
	clientConfig.Seed = true
	clientConfig.Debug = true
	clientConfig.DisableTrackers = true
	clientConfig.ListenAddr = "0.0.0.0:6666"
	clientConfig.DHTConfig = dht.ServerConfig{
		StartingNodes: serverAddrs,
	}
	clientConfig.DataDir = encryptdataPath
	clientConfig.DisableAggressiveUpload = false
	client, _ := torrent.NewClient(&clientConfig)

	dir, _ := os.Open(origindataPath)
	defer dir.Close()

	fi, _ := dir.Readdir(-1)
	for _, x := range fi {
		if !x.IsDir() && x.Name() != ".torrent.bolt.db" {
			key,err:=encryptFile(x.Name())
			if err!=nil{
				log.Fatalln("err in encrypt file")
				return
			}
			d := makeMagnet(encryptdataPath, x.Name(), client)
			fmt.Println(d)
			upload_AddArgs := [][]byte{[]byte(x.Name()),[]byte("hash"),[]byte("keywords"),[]byte("Summary"),[]byte(d),[]byte(key)}
			response, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "myapp", Fcn: "createFile", Args:upload_AddArgs})
			if err != nil {
				fmt.Println("Failed to add a magnetlink: %s", err)
			}else{
				fmt.Println("username : ",string(response.Payload))
			}
			time.Sleep(time.Second*5)
		}
	}

	//monitor origin data path
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
					d:=makeMagnet(encryptdataPath, event.Name(), client)
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

	time.Sleep(time.Second * 5)

	//todo query file
	//query chaincode of myapp:
	// [{"Key":{"objectType":File", "attributes":["keywords", "filename", "User1@org1.example.com"]},
		// "Record":{"name":"filename","hash":"hash","keyword":"keywords","summary":"Summary","owner":"User1@org1.example.com","locktime":0,"Magnet":"magnet:?xt=urn:btih:4b6a1fe45384c3e06dad104aa068c054dfca271e\u0026dn=a.jpg"}}]


	upload_response, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "myapp", Fcn: "queryFile", Args: upload_QueryArgs})
	if err != nil {
			  fmt.Println("Failed to query funds: %s", err)
			  }
	upload_initial :=upload_response.Payload
	fmt.Println("query chaincode of myapp: ",string(upload_initial))

	testChaincodeEventListener("keyExchange",chClientOrg1User)
	select {}
}

func loadOrgUser( sdk *fabsdk.FabricSDK, orgName string, userName string) fab.IdentityContext {

	session, err := sdk.NewClient(fabsdk.WithUser(userName), fabsdk.WithOrg(orgName)).Session()
	if err != nil {
		fmt.Println(errors.Wrapf(err, "Session failed, %s, %s", orgName, userName))
	}
	return session
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

//todo get hostname dynamicly
var dhtserver_Initargs= [][]byte{[]byte("init"), []byte("dht_server"), []byte("server:6666")}
var dht_queryArgs = [][]byte{[]byte("query"), []byte("dht_server")}

var upload_InitArgs = [][]byte{[]byte("init"),[]byte("init"),[]byte("")}
var upload_QueryArgs = [][]byte{[]byte("keywords")}

func DhtServerInitArgs() [][]byte {
	return dhtserver_Initargs
}
