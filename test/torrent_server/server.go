package main

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"

	"github.com/anacrolix/dht"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"encoding/json"
)

type Request struct {
	From string `json:"from"`
	To string `json:"To"`
	File string `json:"file"`
	RequestTime int64 `json:"requestTime"`
	ResponseTime int64 `json:"responseTime"`
	ConfirmationTime int64 `json:"confirmationTime"`
}

type RequestMessage struct {
	From string `json:"from"`
	To string `json:"To"`
	File string `json:"file"`
	TxID string `json:"tx_id"`
}
func makeMagnet(dir string, name string, cl *torrent.Client) string {
	mi := metainfo.MetaInfo{}
	mi.SetDefaults()
	info := metainfo.Info{PieceLength: 1024 * 1024}
	info.BuildFromFilePath(filepath.Join(dir, name))
	mi.InfoBytes, _ = bencode.Marshal(info)
	cl.AddTorrent(&mi)
	magnet := mi.Magnet(name, mi.HashInfoBytes()).String()
	return magnet
}

func serverAddrs() (addrs []dht.Addr, err error) {
	for _, s := range []string{
	} {
		ua, err := net.ResolveUDPAddr("udp4", s)
		if err != nil {
			continue
		}
		addrs = append(addrs, dht.NewAddr(ua))
	}
	if len(addrs) == 0 {
		err = errors.New("nothing resolved")
	}
	return
}

func testChaincodeEventListener(ccID string, listener chclient.ChannelClient) {

	eventID := "requestSecret"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *chclient.CCEvent)
	rce, err := listener.RegisterChaincodeEvent(notifier, ccID, eventID)
	if err != nil {
		fmt.Println("Failed to register cc event: %s", err)
	}


	for{
		select {
		case ccEvent := <-notifier:
			fmt.Println("requestSecret happened")
			message:=RequestMessage{}
			json.Unmarshal(ccEvent.Payload,&message)
			fmt.Println(message)
			_, err :=listener.Execute(chclient.Request{ChaincodeID: "keyExchange", Fcn: "respondSecret", Args:[][]byte{[]byte(message.TxID),[]byte("mysecret1")}})
			if err!=nil{
				fmt.Println("error in respond")
			}else{
				fmt.Println("respondSecret success")
			}
			//case <-time.After(time.Second * 20):
			//	t.Fatalf("Did NOT receive CC for eventId(%s)\n", eventID)
		}

	}

	// Unregister chain code event using registration handle
	err = listener.UnregisterChaincodeEvent(rce)
	if err != nil {
		fmt.Println("Unregister cc event failed: %s", err)
	}

}
