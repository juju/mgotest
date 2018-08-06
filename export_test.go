package mgotest

func ResetGlobalState() {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	if session != nil {
		session.Close()
		session = nil
	}
	dialError = nil
}

var DialTimeout = &dialTimeout
