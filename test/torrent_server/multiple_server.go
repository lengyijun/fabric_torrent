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
)

const (
	dataPath="data"
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

	loadPkg, err := packager.NewCCPackage("github.com/upload", "../fixtures/testdata")
	if err != nil {
		fmt.Println(err)
	}
	installUploadReq := resmgmt.InstallCCRequest{Name: "upload", Path: "github.com/upload", Version: "0", Package:loadPkg}

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

	err = org1ResMgmt.InstantiateCC("orgchannel", resmgmt.InstantiateCCRequest{Name: "upload", Path: "github.com/upload", Version: "0", Args:upload_InitArgs, Policy: ccPolicy})
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
	clientConfig.DataDir = dataPath
	clientConfig.DisableAggressiveUpload = false
	client, _ := torrent.NewClient(&clientConfig)

	dir, _ := os.Open(dataPath)
	defer dir.Close()

	fi, _ := dir.Readdir(-1)
	for _, x := range fi {
		if !x.IsDir() && x.Name() != ".torrent.bolt.db" {
			d := makeMagnet(dataPath, x.Name(), client)
			fmt.Println(d)
			upload_AddArgs := [][]byte{[]byte("add"), []byte(d),[]byte("myipaddr")}
			_, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "upload", Fcn: "invoke", Args:upload_AddArgs})
			if err != nil {
				fmt.Println("Failed to add a magnetlink: %s", err)
			}
			time.Sleep(time.Second*5)
		}
	}

	time.Sleep(time.Second * 5)

	upload_response, err := chClientOrg1User.Execute(chclient.Request{ChaincodeID: "upload", Fcn: "invoke", Args: upload_QueryArgs})
	if err != nil {
			  fmt.Println("Failed to query funds: %s", err)
			  }
	upload_initial :=upload_response.Payload
	fmt.Println("query chaincode of upload: ",string(upload_initial))

	/*


	// Org1 resource manager will instantiate 'example_cc' version 1 on 'orgchannel'
	err = org1ResMgmt.UpgradeCC("orgchannel", resmgmt.UpgradeCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "1", Args:ExampleCCUpgradeArgs(), Policy: org1Andorg2Policy})
	if err != nil {
		fmt.Println(err)
	}

	// Org2 user moves funds on org2 peer (cc policy fails since both Org1 and Org2 peers should participate)
	response, err = chClientOrg2User.Execute(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args:ExampleCCTxArgs()}, chclient.WithProposalProcessor(orgTestPeer1))
	if err == nil {
		fmt.Println("Should have failed to move funds due to cc policy")
	}

	// Org2 user moves funds (cc policy ok since we have provided peers for both Orgs)
	response, err = chClientOrg2User.Execute(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args:ExampleCCTxArgs()}, chclient.WithProposalProcessor(orgTestPeer0, orgTestPeer1))
	if err != nil {
		fmt.Println("Failed to move funds: %s", err)
	}

	// Assert that funds have changed value on org1 peer
	beforeTxValue, _ := strconv.Atoi(ExampleCCUpgradeB)
	expectedValue := beforeTxValue + 1
	verifyValue( chClientOrg1User, expectedValue)

	// Specify user that will be used by dynamic selection service (to retrieve chanincode policy information)
	// This user has to have privileges to query lscc for chaincode data
	mychannelUser := selection.ChannelUser{ChannelID: "orgchannel", UserName: "User1", OrgName: "Org1"}

	// Create SDK setup for channel client with dynamic selection
	sdk, err = fabsdk.New(config.FromFile("./config_test.yaml"),
		fabsdk.WithServicePkg(&DynamicSelectionProviderFactory{ChannelUsers: []selection.ChannelUser{mychannelUser}}))
	if err != nil {
		fmt.Println("Failed to create new SDK: %s", err)
	}

	// Create new client that will use dynamic selection
	chClientOrg2User, err = sdk.NewClient(fabsdk.WithUser("User1"), fabsdk.WithOrg(org2)).Channel("orgchannel")
	if err != nil {
		fmt.Println("Failed to create new channel client for Org2 user: %s", err)
	}

	// Org2 user moves funds (dynamic selection will inspect chaincode policy to determine endorsers)
	response, err = chClientOrg2User.Execute(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args:ExampleCCTxArgs()})
	if err != nil {
		fmt.Println("Failed to move funds: %s", err)
	}

	expectedValue++
	verifyValue( chClientOrg1User, expectedValue)

	*/
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
var upload_QueryArgs = [][]byte{[]byte("query"), []byte("init")}

func DhtServerInitArgs() [][]byte {
	return dhtserver_Initargs
}
