package main

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/xerrors"
	"net"
	"strconv"
	"strings"
)

func addressValidator(ans interface{}) error {
	peerAddr, _ := ans.(string)
	ipAndPort := strings.Split(peerAddr, ":")
	if len(ipAndPort) != 2 {
		// The address given is invalid
		return xerrors.Errorf("Please enter a valid peer address, e.g., 127.0.0.1:4001")
	}

	ipAddr := ipAndPort[0]
	if net.ParseIP(ipAddr) == nil {
		return xerrors.Errorf("Please enter a valid peer address, e.g., 127.0.0.1:4001")
	}

	portNum := ipAndPort[1]
	portN, err := strconv.Atoi(portNum)
	if err != nil || portN < 0 || portN >= 65536 {
		return xerrors.Errorf("Please enter a valid peer address, e.g., 127.0.0.1:4001")
	}

	return nil
}

func hashValidator(ans interface{}) error {
	hash, _ := ans.(string)

	_, err := hex.DecodeString(hash)
	if err != nil {
		return xerrors.Errorf(
			fmt.Sprintf("Please enter a valid %s Hash value in hex decimal string, e.g., 1122...ff",
				config.PasswordHashAlgorithm.String()))
	}

	if len(hash) != config.PasswordHashAlgorithm.Size()*2 {
		return xerrors.Errorf(
			fmt.Sprintf("Please enter a valid %s Hash value, it should have %d bytes.",
				config.PasswordHashAlgorithm.String(), config.PasswordHashAlgorithm.Size()))
	}

	return nil
}

func saltValidator(ans interface{}) error {
	salt, _ := ans.(string)

	_, err := hex.DecodeString(salt)
	if err != nil {
		return xerrors.Errorf("Please enter a valid Salt value in hex decimal string, e.g., 1122...ff")
	}

	if len(salt) != config.ChordBytes*2 {
		return xerrors.Errorf(
			fmt.Sprintf("Please enter a valid Salt value, it should have %d bytes",
				config.ChordBytes))
	}

	return nil
}
