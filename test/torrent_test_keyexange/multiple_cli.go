/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main 

import (
	"path"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	"fmt"
	//"os"

	//"github.com/anacrolix/dht"
	//"github.com/anacrolix/torrent"
)

const (
	dataPath	= "data"
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

	// Load specific targets for move funds test
	loadOrgPeers( sdk)

	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := sdk.NewClient(fabsdk.WithUser("User1"), fabsdk.WithOrg(org2)).Channel("orgchannel")
	if err != nil {
		fmt.Println("Failed to create new channel torrentClient for Org1 user: %s", err)
	}
	_, err = chClientOrg1User.Execute(chclient.Request{ChaincodeID: "keyExchange", Fcn: "requestSecret", Args: requestSecret_Args})
	if err!=nil{
		fmt.Println("error in request secret")
	}else{
		//return nil
		fmt.Println("request secret success")
	}

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

var requestSecret_Args = [][]byte{[]byte("keywords"),[]byte("a.jpg"),[]byte("User1@org1.example.com")}

var upload_InitArgs = [][]byte{[]byte("init"),[]byte("init"),[]byte("myipaddr:port")}
var upload_QueryArgs = [][]byte{[]byte("query"), []byte("init")}
