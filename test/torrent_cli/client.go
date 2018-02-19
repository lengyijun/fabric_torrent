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
)

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
