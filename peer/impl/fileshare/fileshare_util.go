package fileshare

import (
	"github.com/rs/xid"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"log"
	"regexp"
	"strings"
	"time"
)

func (f *FileModule) tryDownloadBlockRemote(dest string, dataRequestMsg types.DataRequestMessage,
	chunkChan *chan []byte, timeout time.Duration) ([]byte, error) {

	dataRequestMsgTrans, err := f.conf.MessageRegistry.MarshalMessage(dataRequestMsg)
	if err != nil {
		return nil, err
	}

	err = f.message.Unicast(dest, dataRequestMsgTrans)
	if err != nil {
		return nil, err
	}

	/* Either we receive the timeout or we receive the chunk */
	select {
	case chunk := <-*chunkChan:
		/* Delete the entry in the channels */
		f.message.Async.Delete(dataRequestMsg.RequestID)
		return chunk, nil
	case <-time.After(timeout):
		/* Back off and retry */
		f.message.Async.Delete(dataRequestMsg.RequestID)
		return nil, xerrors.Errorf("downloadBlock timeout")
	}
}

func (f *FileModule) downloadBlock(hash string) ([]byte, error) {
	/* Check the local storage */
	localStorage := f.conf.Storage.GetDataBlobStore().Get(hash)
	if localStorage != nil {
		return localStorage, nil
	}

	/* Check the catalog for remote peers */
	remoteStorage, ok := f.catalog.Load(hash)
	if !ok {
		/* Nothing is found in either local storage or in the catalog, return an error */
		return nil, xerrors.Errorf("Unable to locate file with hash: %v", hash)
	}
	/* If we find it in the catalog, send a request to a random remote peer */
	if len(remoteStorage.(map[string]struct{})) == 0 {
		return nil, xerrors.Errorf("Unable to locate file with hash: %v", hash)
	}
	randomPeer := f.message.SelectRandomNeighbor(remoteStorage.(map[string]struct{}))

	dataRequestMsg := types.DataRequestMessage{
		RequestID: xid.New().String(),
		Key:       hash,
	}

	/* Make a channel for DataReplyMessage handler to send back the chunk received */
	chunkChan := make(chan []byte, 1)
	f.message.Async.Store(dataRequestMsg.RequestID, chunkChan)

	/* Send the DataRequestMessage to the peer and wait for the response, retry in case of timeout */
	timeout := f.conf.BackoffDataRequest.Initial
	var retry uint
	for retry < f.conf.BackoffDataRequest.Retry {
		chunk, err := f.tryDownloadBlockRemote(randomPeer, dataRequestMsg, &chunkChan, timeout)
		if err != nil {
			timeout = timeout * time.Duration(f.conf.BackoffDataRequest.Factor)
			retry++
		} else {
			return chunk, nil
		}
	}

	/* By default, there is nothing found */
	return nil, xerrors.Errorf("Unable to locate file with hash: %v", hash)
}

func (f *FileModule) sendSearchRequest(reg regexp.Regexp, requestID string, origin string, selectNeighbor string,
	neighborBudget uint) {
	searchRequestMsg := types.SearchRequestMessage{
		RequestID: requestID,
		Origin:    origin,
		Pattern:   reg.String(),
		Budget:    neighborBudget,
	}

	searchRequestMsgTrans, err := f.conf.MessageRegistry.MarshalMessage(searchRequestMsg)
	if err != nil {
		log.Panicln("sendSearchRequest: ", f.address, err)
	}

	err = f.message.Unicast(selectNeighbor, searchRequestMsgTrans)
	if err != nil {
		log.Panicln("sendSearchRequest: ", f.address, err)
	}
}

func (f *FileModule) localFullKnown(pattern regexp.Regexp) string {
	fullKnownName := ""
	f.conf.Storage.GetNamingStore().ForEach(func(name string, metahash []byte) bool {
		found := true
		metaFile := f.conf.Storage.GetDataBlobStore().Get(string(metahash))
		if metaFile == nil {
			found = false
		} else {
			/* Get each individual block */
			chunkHexKeys := strings.Split(string(metaFile), peer.MetafileSep)
			for _, chunkHash := range chunkHexKeys {
				chunkFile := f.conf.Storage.GetDataBlobStore().Get(chunkHash)
				if chunkFile == nil {
					found = false
				}
			}
		}
		if found && pattern.MatchString(name) && fullKnownName == "" {
			fullKnownName = name
		}
		return true
	})
	return fullKnownName
}

func (f *FileModule) expandRingSearch(pattern regexp.Regexp, conf peer.ExpandingRing) string {
	f.fullKnown.Range(func(fileName, _ interface{}) bool {
		f.fullKnown.Delete(fileName)
		return true
	})

	timeout := conf.Timeout
	budget := conf.Initial
	var retry uint
	for retry < conf.Retry {
		remoteNeighborSet := f.message.RemoteNeighbor(map[string]struct{}{})
		if len(remoteNeighborSet) > 0 && budget > 0 {
			selectedNeighbors, budgets := f.message.SelectKNeighbors(budget, remoteNeighborSet)
			requestID := xid.New().String()
			for idx := range selectedNeighbors {
				selectNeighbor := selectedNeighbors[idx]
				neighborBudget := budgets[idx]
				f.sendSearchRequest(pattern, requestID, f.address, selectNeighbor, neighborBudget)
			}
		}

		/* After timeout, we check whether we have a full known name */
		time.Sleep(timeout)

		remoteFullKnownName := ""
		f.fullKnown.Range(func(fileName, _ interface{}) bool {
			if pattern.MatchString(fileName.(string)) && remoteFullKnownName == "" {
				remoteFullKnownName, _ = fileName.(string)
			}
			return true
		})
		if remoteFullKnownName != "" {
			return remoteFullKnownName
		}
		/* Retry */
		budget *= conf.Factor
		retry++
	}
	return ""
}
