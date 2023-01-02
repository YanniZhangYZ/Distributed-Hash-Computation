package impl

import (
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"log"
	"regexp"
	"strings"
)

func (f *FileModule) execDataRequestMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	dataRequestMsg, ok := msg.(*types.DataRequestMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Check the local storage */
	localStorage := f.conf.Storage.GetDataBlobStore().Get(dataRequestMsg.Key)
	dataReplyMsg := types.DataReplyMessage{
		RequestID: dataRequestMsg.RequestID,
		Key:       dataRequestMsg.Key,
		Value:     localStorage,
	}

	dataReplyMsgTrans, err := f.conf.MessageRegistry.MarshalMessage(dataReplyMsg)
	if err != nil {
		return err
	}

	err = f.message.unicast(pkt.Header.Source, dataReplyMsgTrans)
	if err != nil {
		log.Println("ExecDataRequestMessage: ", err)
	}

	return nil
}

func (f *FileModule) execDataReplyMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	dataReplyMsg, ok := msg.(*types.DataReplyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Load the channel from the map and send the chunks */
	chunkChan, _ := f.message.async.Load(dataReplyMsg.RequestID)
	chunkChan.(chan []byte) <- dataReplyMsg.Value
	return nil
}

func (f *FileModule) execSearchRequestMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	searchRequestMsg, ok := msg.(*types.SearchRequestMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Ignore duplicate searches */
	_, ok = f.message.seenRequest.Load(searchRequestMsg.RequestID)
	if ok {
		return nil
	}
	f.message.seenRequest.Store(searchRequestMsg.RequestID, struct{}{})

	/* Find out local match files, only if the peer has the corresponding metafile in its blob store */
	reg := regexp.MustCompile(searchRequestMsg.Pattern)
	var matchFiles []types.FileInfo
	f.conf.Storage.GetNamingStore().ForEach(func(name string, metahash []byte) bool {
		localMeta := f.conf.Storage.GetDataBlobStore().Get(string(metahash))
		if localMeta == nil || !reg.MatchString(name) {
			return true
		}

		var chunks [][]byte
		chunkHexKeys := strings.Split(string(localMeta), peer.MetafileSep)
		for _, chunkHash := range chunkHexKeys {
			localChunk := f.conf.Storage.GetDataBlobStore().Get(chunkHash)
			if localChunk == nil {
				chunks = append(chunks, nil)
			} else {
				chunks = append(chunks, []byte(chunkHash))
			}
		}

		matchFiles = append(matchFiles,
			types.FileInfo{
				Name:     name,
				Metahash: string(metahash),
				Chunks:   chunks,
			})

		return true
	})

	/* Send back the reply without using routing tables */
	searchReplyMsg := types.SearchReplyMessage{
		RequestID: searchRequestMsg.RequestID,
		Responses: matchFiles,
	}
	searchReplyMsgTrans, _ := f.conf.MessageRegistry.MarshalMessage(searchReplyMsg)
	err := f.message.sendDirectMsg(pkt.Header.Source, searchRequestMsg.Origin, searchReplyMsgTrans)
	if err != nil {
		return err
	}

	/* Forward the search if budgets allow */
	if searchRequestMsg.Budget == 1 {
		return nil
	}

	/* Exclude the node that we receive the packet */
	previousTargets := map[string]struct{}{}
	previousTargets[pkt.Header.Source] = struct{}{}
	remoteNeighborSet := f.message.remoteNeighbor(previousTargets)
	if len(remoteNeighborSet) > 0 {
		selectedNeighbors, budgets := f.message.selectKNeighbors(searchRequestMsg.Budget-1, remoteNeighborSet)
		for idx := range selectedNeighbors {
			selectNeighbor := selectedNeighbors[idx]
			neighborBudget := budgets[idx]
			f.sendSearchRequest(*reg, searchRequestMsg.RequestID, searchRequestMsg.Origin, selectNeighbor, neighborBudget)
		}
	}

	return nil
}

func (f *FileModule) execSearchReplyMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	searchReplyMsg, ok := msg.(*types.SearchReplyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	for _, fileInfo := range searchReplyMsg.Responses {
		/* Update the naming store and catalog */
		f.conf.Storage.GetNamingStore().Set(fileInfo.Name, []byte(fileInfo.Metahash))
		f.updateCatalog(fileInfo.Metahash, pkt.Header.Source)

		remoteFullKnown := true
		for _, chunkHash := range fileInfo.Chunks {
			if chunkHash != nil {
				f.updateCatalog(string(chunkHash), pkt.Header.Source)
			} else {
				remoteFullKnown = false
			}
		}

		/* Update the full known names */
		if remoteFullKnown {
			f.fullKnown.Store(fileInfo.Name, struct{}{})
		}
	}
	return nil
}
