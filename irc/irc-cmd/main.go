package main

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"github.com/ivartj/norske-irc-kanaler.com/irc"
	"io"
	"net"
	"os"
	"strings"
)

func mainUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage: %s\n", os.Args[0])
	fmt.Fprintf(w, "  [-w <password>]\n")
	fmt.Fprintf(w, "  [-U <username>]\n")
	fmt.Fprintf(w, "  [--ssl-no-verify]\n")
	fmt.Fprintf(w, "  [--ssl-fingerprint <sha1-fingerprint>]\n")
	fmt.Fprintf(w, "  <server>[:<port>] \\#channel\n")
}

func main() {
	var err error
	var password string = ""
	var sslFingerprint = ""
	useSSL := false
	user := "testuser"
	nick := "testnick"
	tok := args.NewTokenizer(os.Args)
	positionals := []string{}
	for tok.Next() {
		if tok.IsOption() {
			switch tok.Arg() {
			case "-h", "--help":
				mainUsage(os.Stdout)
				return
			case "--version":
				fmt.Println("irc-cmd version 0.1.0")
				return
			case "-n":
				nick, err = tok.TakeParameter()
				if err != nil {
					panic(err)
				}
			case "-U":
				user, err = tok.TakeParameter()
				if err != nil {
					panic(err)
				}
			case "-w":
				password, err = tok.TakeParameter()
				if err != nil {
					panic(err)
				}
			case "--ssl-no-verify":
				useSSL = true
			case "--ssl-fingerprint":
				useSSL = true
				sslFingerprint, err = tok.TakeParameter()
				if err != nil {
					panic(err)
				}
			default:
				fmt.Fprintf(os.Stderr, "Unrecognized option '%s'\n", tok.Arg())
				os.Exit(1)
			}
		} else {
			positionals = append(positionals, tok.Arg())
		}
	}
	if len(positionals) != 2 {
		mainUsage(os.Stderr)
		os.Exit(1)
	}
	server := positionals[0]
	channel := positionals[1]

	var conn net.Conn
	if useSSL {
		config := &tls.Config{InsecureSkipVerify: true}
		if sslFingerprint != "" {
			config.VerifyConnection = func(cs tls.ConnectionState) error {
				sha1bytes := sha1.Sum(cs.PeerCertificates[0].Raw)
				sha1string := hex.EncodeToString(sha1bytes[:])
				sha1string = strings.ToLower(sha1string)
				sslFingerprint = strings.ToLower(sslFingerprint)
				if sha1string != sslFingerprint {
					return fmt.Errorf("Expected SHA1 fingerprint '%s' does not match actual fingerprint '%s'", sslFingerprint, sha1string)
				}
				return nil
			}
		}
		conn, err = tls.Dial("tcp", server, config)
	} else {
		conn, err = net.Dial("tcp", server)
	}
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c, err := irc.New(conn, &irc.Config{
		Nick:     nick,
		User:     user,
		Password: password,
		Log:      os.Stderr,
	})
	if err != nil {
		panic(err)
	}
	defer c.Disconnect()

	c.SendRawf("LIST %s", channel)
	ev, err := c.WaitUntil("322") // RPL_LIST
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", ev.Args)
}
