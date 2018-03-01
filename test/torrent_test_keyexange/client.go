package main

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/anacrolix/dht"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
	"github.com/gosuri/uiprogress"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"encoding/json"
)
type ResponseMessage struct {
	From string `json:"from"`
	To []string `json:"To"`
	File string `json:"file"`
	TxID []string `json:"tx_id"`
	Secret string `json:"secret"`
}

func generateClientAddrs(inputaddr [] string) (func  ()(addrs []dht.Addr,err error)){
	return func()(addrs []dht.Addr,err error){
		for _, s := range (inputaddr) {
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
}

func clientAddrs(inputaddr []string) (addrs []dht.Addr, err error) {
	for _, s := range (inputaddr) {
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

func torrentBar(t *torrent.Torrent) {
	bar := uiprogress.AddBar(1)
	bar.AppendCompleted()
	bar.AppendFunc(func(*uiprogress.Bar) (ret string) {
		select {
		case <-t.GotInfo():
		default:
			return "getting info"
		}
		if t.Seeding() {
			return "seeding"
		} else if t.BytesCompleted() == t.Info().TotalLength() {
			return "completed"
		} else {
			return fmt.Sprintf("downloading (%s/%s)", humanize.Bytes(uint64(t.BytesCompleted())), humanize.Bytes(uint64(t.Info().TotalLength())))
		}
	})
	bar.PrependFunc(func(*uiprogress.Bar) string {
		return t.Name()
	})
	go func() {
		<-t.GotInfo()
		tl := int(t.Info().TotalLength())
		if tl == 0 {
			bar.Set(1)
			return
		}
		bar.Total = tl
		for {
			bc := t.BytesCompleted()
			bar.Set(int(bc))
			time.Sleep(time.Second)
		}
	}()
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

func download(client * torrent.Client,magnetUrl string){
	if magnetUrl=="" {return}
	t, _ := client.AddMagnet(magnetUrl)
	torrentBar(t)
	go func() {
		<-t.GotInfo()
		t.DownloadAll()
	}()
	uiprogress.Start()
}

func testChaincodeEventListener(ccID string, listener chclient.ChannelClient) {

	eventID := "respondSecret"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *chclient.CCEvent)
	rce, err := listener.RegisterChaincodeEvent(notifier, ccID, eventID)
	if err != nil {
		fmt.Println("Failed to register cc event: %s", err)
	}


	for{
		select {
		case ccEvent := <-notifier:
			fmt.Println("respond secret event happened")
			var responseMessage=ResponseMessage{}
			json.Unmarshal(ccEvent.Payload,&responseMessage)
			fmt.Println(responseMessage)

			//responseMessage.txid is string []!
			_, err :=listener.Execute(chclient.Request{ChaincodeID: "keyExchange", Fcn: "confirmSecret", Args:[][]byte{[]byte(responseMessage.TxID[0])}})
			if err!=nil{
				fmt.Println("error in confirm secret")
			}else{
				fmt.Println("confirm secret success")
				queryResponse,err :=listener.Execute(chclient.Request{ChaincodeID: "keyExchange", Fcn: "queryRequest", Args:[][]byte{[]byte(responseMessage.TxID[0])}})
				if err!=nil {
					fmt.Println("error in query request")
				}else{
					fmt.Println("success in query request")
					fmt.Println(string(queryResponse.Payload))
				}
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
