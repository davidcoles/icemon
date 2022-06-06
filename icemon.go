package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const BUFFSZ = 8192

var verbose bool
var servers bool
var client *http.Client = &http.Client{}

func main() {

	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.BoolVar(&servers, "s", false, "servers")
	flag.Parse()

	args := flag.Args()

	if servers {
		for _, ip := range args[1:] {
			go bar(ip, args[0])
		}
	} else {
		for _, mp := range args[1:] {
			go bar(args[0], mp)
		}
	}

	for {
		time.Sleep(10 * time.Second)
	}
}

func bar(ip, mp string) {

	for {
		stream("http://" + ip + "/" + mp)
		time.Sleep(10 * time.Second)
	}
}

func stream(endpoint string) {

	req, err := http.NewRequest("GET", endpoint, nil)
	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(endpoint, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(endpoint, resp.StatusCode)
		return
	}

	log.Println(endpoint, "started")

	metaint := uint(0)

	for k, v := range resp.Header {
		if k == "Icy-Metaint" {
			i, err := strconv.Atoi(v[0])

			if err != nil || i < 0 {
				log.Println(endpoint, "icy-metaint must be a positive integer", v[0])
				return
			}

			metaint = uint(i)
		}
	}

	demux := demuxmeta(metaint)

	for {
		var buff [BUFFSZ]byte

		if nread, _ := io.ReadFull(resp.Body, buff[:]); nread != BUFFSZ || err != nil {
			log.Println(endpoint, "finished", nread, BUFFSZ, err)
			return
		}

		demux(buff[:], func(b []byte, m bool) {
			if m && verbose {
				log.Println(endpoint, string(b))
			}
		})
	}
}

func demuxmeta(mint uint) func([]byte, func([]byte, bool)) {

	if mint == 0 {
		return func(buff []byte, f func([]byte, bool)) {
			f(buff, false)
		}
	}

	stat := 0               // state: 0 - data, 1 - metaint byte, 2 - metadata
	todo := mint            // remaining bytes for this state
	meta := make([]byte, 0) // metadata buffer

	return func(buff []byte, f func([]byte, bool)) {
		for len(buff) > 0 {
			switch stat {
			case 0: // not in metadata
				if uint(len(buff)) < todo {
					f(buff, false)
					todo -= uint(len(buff))
					return
				}
				f(buff[:todo], false)
				buff = buff[todo:] // may be empty - caught by for{} condition
				todo = 0
				stat = 1

			case 1: // read metalen byte
				todo = uint(buff[0]) << 4    // * 16
				meta = make([]byte, 0, 4096) // greater then 255*16
				buff = buff[1:]
				stat = 2

			case 2: // in metadata
				if uint(len(buff)) < todo {
					meta = append(meta, buff[:]...)
					todo -= uint(len(buff))
					return
				}
				meta := append(meta, buff[:todo]...)
				f(meta, true)
				buff = buff[todo:]
				todo = mint
				stat = 0
			}
		}
	}
}
