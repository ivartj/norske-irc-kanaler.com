package main

type confSetting struct {
	isSet_ bool
	value  string
}

var conf struct {
	inputFilename  confSetting
	outputFilename confSetting
}

func (s confSetting) get() string {
	return s.value
}

func (s confSetting) set(val string) {
	s.isSet_ = true
	s.value = val
}

func (s confSetting) isSet() bool {
	return s.isSet_
}
