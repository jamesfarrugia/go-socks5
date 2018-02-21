package main

func doAuthNone(data []byte, conn *serverConnection) {
	log.Debug("NOOP auth - pass through")
	conn.status = stsWrCmd
}

func doAuthPass(data []byte, conn *serverConnection) {
	var hIdx int = 0
	var version int = -1
	var lenUname int = -1
	var lenPassw int = -1
	var username string = ""
	var password string = ""

	if len(data) >= hIdx+1 {
		version = int(data[hIdx])
		hIdx++
	}

	if len(data) >= hIdx+1 {
		lenUname = int(data[hIdx])
		hIdx++
	}

	if len(data) >= (lenUname + hIdx) {
		username = string(data[hIdx : hIdx+lenUname])
		hIdx += lenUname
	}

	if len(data) >= hIdx+1 {
		lenPassw = int(data[hIdx])
		hIdx++
	}

	if len(data) >= (lenPassw + hIdx) {
		password = string(data[hIdx : hIdx+lenPassw])
		hIdx += lenPassw
	}

	if len(data) >= hIdx {
		if version == 1 && username == "james" && password == "james" {
			log.Info("User authenticated")
			conn.status = stsWrCmd
		} else {
			log.Info("User failed authentication")
			conn.status = stsWrAuthErr
		}
	}
}
