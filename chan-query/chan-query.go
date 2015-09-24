package query

import (
	"github.com/ivartj/norske-irc-kanaler.com/irc"
	"fmt"
	"strconv"
	"strings"
)

var (
	ListMethod = &Method{
		name: "list",
		query: listQuery,
		description: "Employs the LIST IRC protocol command.",
	}
	JoinMethod = &Method{
		name: "join",
		query: joinQuery,
		description: "Attempts to get channel information by joining the channel.",
	 }
)

func GetMethods() []*Method {
	return []*Method{
		ListMethod,
		JoinMethod,
	}
}

func GetMethodByName(name string) (*Method, bool) {
	for _, v := range GetMethods() {
		if v.Name() == name {
			return v, true
		}
	}
	return nil, false
}

type Result struct {
	Name string
	Server string
	NumberOfUsers int
	Topic string
}

type Method struct {
	name string
	query func(*irc.Conn, string)(*Result, error)
	description string
}

func (m *Method) Name() string {
	return m.name
}

func (m *Method) Description() string {
	return m.description
}

func (m *Method) Query(conn *irc.Conn, channelName string) (*Result, error) {
	res, err := m.query(conn, channelName)
	if err != nil {
		return nil, err
	}
	res.Name = channelName
	res.Server = conn.GetServer()
	return res, nil
}

func listQuery(conn *irc.Conn, channelName string) (*Result, error) {

	conn.SendRawf("LIST %s", channelName)

	numusers := 0
	topic := ""
	received322 := false

	for {
		ev, err := conn.WaitUntil(
			"322", // RPL_LIST
			"323", // RPL_LISTEND
			"401", // ERR_NOSUCHNICK (nick/channel)
			"403", // ERR_NOSUCHCHANNEL
		)
		if err != nil {
			return nil, err
		}

		switch ev.Code {
		case "322":
			received322 = true
			if len(ev.Args) < 4 {
				return nil, fmt.Errorf("Unexpectedly short LIST response on %s", channelName)
			}
			numusers, err = strconv.Atoi(ev.Args[2])
			if err != nil {
				return nil, fmt.Errorf("Failed to parse number of channel users: %s", err.Error())
			}
			topic = ev.Args[3]

		case "323":
			if received322 == false {
				return nil, fmt.Errorf("No status data for %s received on query", channelName)
			}
			goto ret

		case "401": fallthrough
		case "403":
			return nil, fmt.Errorf("No such channel, '%s'", channelName)

		}
	}

ret:
	return &Result{
		NumberOfUsers: numusers,
		Topic: topic,
	}, nil
}

func joinQuery(conn *irc.Conn, channelName string) (*Result, error) {
	conn.SendRawf("JOIN %s", channelName)
	conn.SendRawf("TOPIC %s", channelName)

	numusers := 0
	topic := ""

	receivedTopic, receivedNames := false, false

	for {
		ev, err := conn.WaitUntil(
			"353", // RPL_NAMREPLY
			"366", // RPL_ENDOFNAMES
			"331", // RPL_NOTOPIC
			"332", // RPL_TOPIC
			"422", // ERR_NOMOTD
		)
		if err != nil {
			return nil, err
		}

		switch ev.Code {
		case "353":
			if len(ev.Args) != 0 {
				numusers += len(strings.Split(ev.Args[len(ev.Args) - 1], " "))
			}

		case "332":
			topic = ev.Args[len(ev.Args) - 1]
			receivedTopic = true

		case "366":
			receivedNames = true

		case "422": continue
		case "331":
			receivedTopic = true
		}

		if receivedNames && receivedTopic {
			break
		}
	}

	// TODO: Make it possible to add a parting message
	conn.SendRawf("PART %s", channelName)

	return &Result{
		NumberOfUsers: numusers,
		Topic: topic,
	}, nil
}

