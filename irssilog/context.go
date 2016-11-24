package irssilog

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
)

type Context struct{
	logdir string
	networkNames map[string][]string
}

func New(logdir string, networkNames map[string][]string) *Context {
	return &Context{
		logdir: logdir,
		networkNames: networkNames,
	}
}

func (ctx *Context) ChannelStatus(channel, network string) (ChannelStatus, error) {

	networkNames, ok := ctx.networkNames[network]
	if !ok {
		networkNames = []string{ network }
	}

	f, err := os.Open(ctx.logdir)
	if err != nil {
		return ChannelStatus{}, fmt.Errorf("Failed to open log directory: %s", err.Error())
	}

	filenames, err := f.Readdirnames(0)
	if err != nil {
		return ChannelStatus{}, fmt.Errorf("Failed to read log directory: %s", err.Error())
	}

	latestErr := error(nil)
	networkdir := ""
	status := ChannelStatus{}
	statusFound := false
	for _, filename := range filenames {

		for _, networkname := range networkNames {

			if strings.ToLower(filename) == strings.ToLower(networkname) {
				networkdir = filename

				channelfilename := filepath.Join(ctx.logdir, networkdir, channel + ".log")
				channelfile, err := os.Open(channelfilename)
				if err != nil {
					latestErr = fmt.Errorf("Failed to open '%s': %s", channelfilename, err.Error())
					goto nextfile
				}
				defer channelfile.Close()

				candStatus, err := GetChannelStatusFromLog(channelfile)
				if err != nil {
					latestErr = fmt.Errorf("Error on reading status from '%s': %s", channelfilename, err.Error())
					goto nextfile
				}

				if !statusFound {
					status = candStatus
				} else if candStatus.Time.After(status.Time) {
					status = candStatus
				}
				statusFound = true
			}

		}

nextfile:
	}

	if !statusFound && latestErr != nil {
		return ChannelStatus{}, fmt.Errorf("No status could be retrieved for %s@%s: %s", channel, network, latestErr.Error())
	}

	if !statusFound {
		return ChannelStatus{}, fmt.Errorf("No status could be retrieved for %s@%s", channel, network) 
	}

	return status, nil
}

