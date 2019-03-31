package nttp

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

func BenchmarkNTTStream_Read(b *testing.B) {
	in, err := os.Open("benchmark/citylots.ntt")
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()
	ntt := newNTTReadWriter(in)
	_, err = ioutil.ReadAll(ntt)
	if err != nil {
		log.Fatal(err)
	}
}

func BenchmarkNTTStream_Write(b *testing.B) {
	in, err := os.Open("benchmark/citylots.json")
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenFile("benchmark/citylots.ntt", os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	ntt := newNTTReadWriter(out)
	tr := io.TeeReader(in, ntt)
	_, err = ioutil.ReadAll(tr)
	if err != nil {
		log.Fatal(err)
	}
}

func TestTCPConn(t *testing.T) {
	go func() {
		l, err := net.Listen("tcp", ":56789")
		if err != nil {
			log.Fatal(err)
			return
		}
		defer l.Close()
		for {
			conn, err := l.Accept()
			log.Println("Connection Accepted")
			if err != nil {
				log.Fatal(err)
				return
			}
			go func() {
				defer conn.Close()
				buf := make([]byte, 4096)
				for {
					nread, err := conn.Read(buf)
					if err != nil && err != io.EOF {
						log.Fatal(err)
						return
					}
					log.Println(nread)
					if err == io.EOF {
						break
					}
				}
			}()
		}
	}()

	conn, err := net.Dial("tcp", ":56789")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()
	conn.Write([]byte{0x01, 0x03, 0x04})
	conn.Write([]byte{0x04, 0x05})

	time.Sleep(1 * time.Second)
}
