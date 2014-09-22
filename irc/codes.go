package irc

import "fmt"

var CodeMap map[string]string = map[string]string{

// errors
	"401": "No such nick/channel",
	"402": "No such server",
	"403": "No such channel",
	"404": "Cannot send to channel",
	"405": "Joined too many channels",
	"406": "There was no such nickname",
	"407": "Too many targets",
	"408": "No such service",
	"409": "No origin specified",

	"411": "No recipient given",
	"412": "No text to send",
	"413": "No top-level domain specified",
	"414": "Wildcard in toplevel domain",
	"415": "Bad server/host mask",

	"421": "Unknown command",
	"422": "No message of the day",
	"423": "No administrative info available",
	"424": "File error",

	"431": "No nickname given",
	"432": "Erroneous nickname",
	"433": "Nickname already in use",

	"436": "Nickname collision KILL",
	"437": "Nick/channel is temporarily unavailable",

	"441": "User is not in channel",
	"442": "Client not in channel",
	"443": "Client already in channel",
	"444": "User not logged in",
	"445": "SUMMON has been disabled",
	"446": "USERS has been disabled",

	"451": "You have not registered",

	"461": "Not enough parameters",
	"462": "Unauthorized command (already registered)",
	"463": "Client host is not among the privileged",
	"464": "Password incorrect",
	"465": "Client is banned",
	"466": "Client is getting banned",
	"467": "Channel key already set",

	"471": "Cannot join channel (+l)",
	"472": "Unknown channel mode",
	"473": "Cannot join channel (+i)",
	"474": "Cannot join channel (+b)",
	"475": "Cannot join channel (+k)",
	"476": "Bad channel mask",
	"477": "Channel does not support modes",
	"478": "Channel list is full",

	"481": "Client is not IRC operator",
	"482": "Client is not channel operator",
	"483": "Can't kill server",
	"484": "Client's connection is restricted",
	"485": "Client is not the original channel operator",

	"491": "No O-lines for host",

	"501": "Unknown MODE flag",
	"502": "Cannot change mode for other users",

}

func CodeIsError(code string) bool {
	num := 0
	fmt.Sscanf(code, "%d", &num)
	return num >= 400 && num <= 599
}

func CodeString(code string) string {
	str, ok := CodeMap[code]
	if !ok {
		return fmt.Sprintf("Unknown code (%s)", code)
	}
	return fmt.Sprintf("%s (%s)", str, code)
}
