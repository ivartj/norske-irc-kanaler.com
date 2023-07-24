package znc

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/ivartj/norske-irc-kanaler.com/irc"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Channel interface {
	Name() string
	Network() string
}

type ChannelStatus struct {
	Channel  string
	Network  string
	NumUsers int
	Topic    string
}

type Config interface {
	ZncHost() string
	ZncPort() uint
	ZncUser() string
	ZncPassword() string
	ZncTlsFingerprint() string
	ZncNetworkNames() map[string]string
}

func gatherChannelStatus(cfg Config, conn *irc.Conn, channel Channel) (*ChannelStatus, error) {
	conn.SendRawf("LIST %s", channel.Name())
	ev, err := conn.WaitUntil("322", "323") // RPL_LIST, RPL_LISTEND
	if err != nil {
		return nil, fmt.Errorf("Error while executing the IRC LIST command: %w", err)
	}
	if ev.Code == "323" {
		return nil, fmt.Errorf("Empty data in response to LIST command")
	} else {
		_, err = conn.WaitUntil("323") // RPL_LISTEND
		if err != nil {
			log.Printf("Error while while waiting for 323 RPL_LISTEND: %s\n", err)
		}
	}
	var numUsers int64 = 0
	if len(ev.Args) >= 3 {
		numUsers, err = strconv.ParseInt(ev.Args[2], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse the number of users for channel %s: %w", channel.Name(), err)
		}
	} else {
		return nil, fmt.Errorf("Number of users was not included in RPL_LIST response")
	}
	topic := ""
	if len(ev.Args) == 4 {
		topic = ev.Args[3]
	}
	return &ChannelStatus{
		Channel:  channel.Name(),
		Network:  channel.Network(),
		NumUsers: int(numUsers),
		Topic:    topic,
	}, nil
}

func newConn(cfg Config, network string) (*irc.Conn, error) {
	var netConn net.Conn
	var err error
	zncAddress := cfg.ZncHost()
	if cfg.ZncPort() != 0 { // TODO: allow use of 0 port
		zncAddress += fmt.Sprintf(":%d", cfg.ZncPort())
	} else {
		zncAddress += fmt.Sprintf(":1025") // default znc port
	}
	tlsFingerprint := cfg.ZncTlsFingerprint()
	if tlsFingerprint == "" {
		netConn, err = net.Dial("tcp", zncAddress)
	} else {
		config := &tls.Config{InsecureSkipVerify: true}
		config.VerifyConnection = func(cs tls.ConnectionState) error {
			sha1bytes := sha1.Sum(cs.PeerCertificates[0].Raw)
			sha1string := hex.EncodeToString(sha1bytes[:])
			sha1string = strings.ToLower(sha1string)
			tlsFingerprint = strings.ToLower(tlsFingerprint)
			if sha1string != tlsFingerprint {
				return fmt.Errorf("Expected SHA1 fingerprint '%s' does not match actual fingerprint '%s'", tlsFingerprint, sha1string)
			}
			return nil
		}
		netConn, err = tls.Dial("tcp", zncAddress, config)
	}
	if err != nil {
		return nil, err
	}

	zncNetworkName, ok := cfg.ZncNetworkNames()[network]
	if !ok {
		return nil, fmt.Errorf("No ZNC network name configured for network '%s'", network)
	}
	nick := cfg.ZncUser()
	user := fmt.Sprintf("%s/%s", cfg.ZncUser(), zncNetworkName)
	ircConn, err := irc.New(netConn, &irc.Config{
		Nick:     nick,
		User:     user,
		Password: cfg.ZncPassword(),
		Log:      io.Discard,
	})
	if err != nil {
		return nil, fmt.Errorf("Error connecting to ZNC for network %s: %w", network, err)
	}

	return ircConn, nil
}

func GatherNetworkStatus(cfg Config, network string, channels []Channel) (<-chan *ChannelStatus, error) {
	conn, err := newConn(cfg, network)
	if err != nil {
		return nil, err
	}

	chs := make(chan *ChannelStatus)

	go func() {
		defer conn.Disconnect()
		for _, channel := range channels {
			cs, err := gatherChannelStatus(cfg, conn, channel)
			if err != nil {
				log.Printf("Failed to get channel status for %s/%s: %s\n", channel.Network(), channel.Name(), err)
				continue
			}
			chs <- cs
		}
		close(chs)
	}()

	return chs, nil
}
