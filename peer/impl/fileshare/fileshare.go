package fileshare

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"github.com/rs/xid"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/types"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

func NewFileModule(conf *peer.Configuration, message *message.MessageModule) *FileModule {
	var catalog, fullKnown sync.Map
	file := FileModule{
		address:   conf.Socket.GetAddress(),
		conf:      conf,
		message:   message,
		catalog:   &catalog,
		fullKnown: &fullKnown,
	}

	/* File sharing callbacks */
	conf.MessageRegistry.RegisterMessageCallback(types.DataRequestMessage{}, file.execDataRequestMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.DataReplyMessage{}, file.execDataReplyMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.SearchRequestMessage{}, file.execSearchRequestMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.SearchReplyMessage{}, file.execSearchReplyMessage)

	return &file
}

type FileModule struct {
	address           string
	conf              *peer.Configuration // The configuration contains Socket and MessageRegistry
	message           *message.MessageModule
	catalog           *sync.Map // The catalog for file sharing of the peer
	catalogUpdateLock sync.Mutex
	fullKnown         *sync.Map // Full known file names for remote peers
}

func (f *FileModule) Upload(data io.Reader) (string, error) {
	var metafileKey []byte
	var metafileValue []string

	computeHash := func(content []byte) ([]byte, string) {
		h := crypto.SHA256.New()
		h.Write(content)
		hashSlice := h.Sum(nil)
		hashHex := hex.EncodeToString(hashSlice)
		return hashSlice, hashHex
	}

	for {
		/* Read content out chunk-by-chunk */
		var buf = make([]byte, f.conf.ChunkSize)
		readLen, err := data.Read(buf)
		if err == io.EOF {
			break
		}

		/* Compute the hash of the chunk, and store the chunk into local storage */
		hashSlice, hashHex := computeHash(buf[:readLen])
		f.conf.Storage.GetDataBlobStore().Set(hashHex, buf[:readLen])

		/* Update the information in meta file */
		metafileKey = append(metafileKey, hashSlice...)
		metafileValue = append(metafileValue, hashHex)
	}

	/* Compute the mata file's hash and content, and store it locally */
	_, metahashHex := computeHash(metafileKey)
	f.conf.Storage.GetDataBlobStore().Set(metahashHex, []byte(strings.Join(metafileValue, peer.MetafileSep)))

	return metahashHex, nil
}

func (f *FileModule) Download(metahash string) ([]byte, error) {
	var file []byte

	/* First get the meta file */
	metaFile, err := f.downloadBlock(metahash)
	if err != nil {
		return nil, err
	}

	/* Get each individual block */
	chunkHexKeys := strings.Split(string(metaFile), peer.MetafileSep)
	for _, chunkHash := range chunkHexKeys {
		chunkData, err := f.downloadBlock(chunkHash)
		if err != nil {
			return nil, err
		}
		file = append(file, chunkData...)
	}

	/* Upload to our local storage as well */
	_, _ = f.Upload(bytes.NewReader(file))

	return file, nil
}

func (f *FileModule) Tag(name string, mh string) error {
	f.conf.Storage.GetNamingStore().Set(name, []byte(mh))
	return nil
}

func (f *FileModule) Resolve(name string) string {
	return string(f.conf.Storage.GetNamingStore().Get(name))
}

func (f *FileModule) GetCatalog() peer.Catalog {
	/* Make a copy of the catalog */
	var copyCatalog = make(peer.Catalog)

	f.catalog.Range(func(hash, peers interface{}) bool {
		copyCatalog[hash.(string)], _ = peers.(map[string]struct{})
		return true
	})

	return copyCatalog
}

func (f *FileModule) UpdateCatalog(key string, peer string) {
	/* Update or create the entry */
	f.catalogUpdateLock.Lock()
	fileHosts, ok := f.catalog.Load(key)
	if !ok {
		m := make(map[string]struct{})
		m[peer] = struct{}{}
		f.catalog.Store(key, m)
	} else {
		fileHosts.(map[string]struct{})[peer] = struct{}{}
		f.catalog.Store(key, fileHosts)
	}
	f.catalogUpdateLock.Unlock()
}

func (f *FileModule) SearchAll(reg regexp.Regexp, budget uint, timeout time.Duration) (names []string, err error) {
	/* Check for remote peers, and wait for responses */
	remoteNeighborSet := f.message.RemoteNeighbor(map[string]struct{}{})
	if len(remoteNeighborSet) > 0 && budget > 0 {
		selectedNeighbors, budgets := f.message.SelectKNeighbors(budget, remoteNeighborSet)
		requestID := xid.New().String()
		for idx := range selectedNeighbors {
			selectNeighbor := selectedNeighbors[idx]
			neighborBudget := budgets[idx]
			f.sendSearchRequest(reg, requestID, f.address, selectNeighbor, neighborBudget)
		}
	}

	time.Sleep(timeout)

	/* Search in the local name store, the filenames should already be in the local name store */
	var matchNames []string
	f.conf.Storage.GetNamingStore().ForEach(func(name string, metahash []byte) bool {
		if reg.MatchString(name) {
			matchNames = append(matchNames, name)
		}
		return true
	})

	return matchNames, nil
}

func (f *FileModule) SearchFirst(pattern regexp.Regexp, conf peer.ExpandingRing) (string, error) {
	/* Check that local already have a full file */
	localFullKnownName := f.localFullKnown(pattern)
	if localFullKnownName != "" {
		return localFullKnownName, nil
	}
	/* Check for the remote peers */
	remoteFUllKnownName := f.expandRingSearch(pattern, conf)
	return remoteFUllKnownName, nil
}
