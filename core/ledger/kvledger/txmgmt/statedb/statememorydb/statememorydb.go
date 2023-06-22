/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package statememorydb

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/ledger/internal/version"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/stateleveldb"
)

var logger = flogging.MustGetLogger("statememorydb")

var (
	dataKeyPrefix          = []byte{'d'}
	dataKeyStopper         = []byte{'e'}
	nsKeySep               = []byte{0x00}
	lastKeyIndicator       = byte(0x01)
	savePointKey           = []byte{'s'}
	maxDataImportBatchSize = 4 * 1024 * 1024
)

// VersionedDBProvider implements interface VersionedDBProvider
type VersionedDBProvider struct {
}

// NewVersionedDBProvider instantiates VersionedDBProvider
func NewVersionedDBProvider(dbPath string) (*VersionedDBProvider, error) {
	logger.Debugf("constructing VersionedDBProvider dbPath=%s", dbPath)
	fmt.Printf("### In statememorydb, constructin VersionedDBprovider dbPath=%s, but will not use it ###\n", dbPath)

	return &VersionedDBProvider{}, nil
}

// GetDBHandle gets the handle to a named database
func (provider *VersionedDBProvider) GetDBHandle(dbName string, namespaceProvider statedb.NamespaceProvider) (statedb.VersionedDB, error) {
	return newVersionedDB(dbName), nil
}

// ImportFromSnapshot loads the public state and pvtdata hashes from the snapshot files previously generated
func (provider *VersionedDBProvider) ImportFromSnapshot(
	dbName string,
	savepoint *version.Height,
	itr statedb.FullScanIterator,
) error {
	fmt.Printf("### In State Memorydb.go, ImportFromSnapshot, dbname=%s, savepoint=%v ###\n", dbName, savepoint)
	_ = newVersionedDB(dbName)
	return nil
	// return vdb.importState(itr, savepoint)
}

// BytesKeySupported returns true if a db created supports bytes as a key
func (provider *VersionedDBProvider) BytesKeySupported() bool {
	return true
}

// Close closes the underlying db
func (provider *VersionedDBProvider) Close() {
}

// Drop drops channel-specific data from the state leveldb.
// It is not an error if a database does not exist.
func (provider *VersionedDBProvider) Drop(dbName string) error {
	return nil
}

// VersionedDB implements VersionedDB interface
type versionedDB struct {
	db     map[string][]byte
	dbName string
}

func newVersionedDB(dbName string) *versionedDB {
	return &versionedDB{make(map[string][]byte), dbName}
}

// Open implements method in VersionedDB interface
func (vdb *versionedDB) Open() error {
	// do nothing because using memory
	return nil
}

// Close implements method in VersionedDB interface
func (vdb *versionedDB) Close() {
	// do nothing because using memory
}

// ValidateKeyValue implements method in VersionedDB interface
func (vdb *versionedDB) ValidateKeyValue(key string, value []byte) error {
	return nil
}

// BytesKeySupported implements method in VersionedDB interface
func (vdb *versionedDB) BytesKeySupported() bool {
	return true
}

// GetState implements method in VersionedDB interface
func (vdb *versionedDB) GetState(namespace string, key string) (*statedb.VersionedValue, error) {
	logger.Debugf("GetState(). ns=%s, key=%s", namespace, key)
	encodekey := hex.EncodeToString(encodeDataKey(namespace, key))
	fmt.Printf("### In State Memorydb.go, GetState, ns=%s, key=%s, encodekey=%s ###\n", namespace, key, encodekey)
	dbVal, _ := vdb.db[encodekey]
	if dbVal == nil {
		return nil, nil
	}
	return stateleveldb.DecodeValue(dbVal)
}

// GetVersion implements method in VersionedDB interface
func (vdb *versionedDB) GetVersion(namespace string, key string) (*version.Height, error) {
	versionedValue, err := vdb.GetState(namespace, key)
	if err != nil {
		return nil, err
	}
	if versionedValue == nil {
		return nil, nil
	}
	return versionedValue.Version, nil
}

// GetStateMultipleKeys implements method in VersionedDB interface
func (vdb *versionedDB) GetStateMultipleKeys(namespace string, keys []string) ([]*statedb.VersionedValue, error) {
	vals := make([]*statedb.VersionedValue, len(keys))
	for i, key := range keys {
		val, err := vdb.GetState(namespace, key)
		if err != nil {
			return nil, err
		}
		vals[i] = val
	}
	return vals, nil
}

// GetStateRangeScanIterator implements method in VersionedDB interface
// startKey is inclusive
// endKey is exclusive
func (vdb *versionedDB) GetStateRangeScanIterator(namespace string, startKey string, endKey string) (statedb.ResultsIterator, error) {
	// pageSize = 0 denotes unlimited page size
	return vdb.GetStateRangeScanIteratorWithPagination(namespace, startKey, endKey, 0)
}

// GetStateRangeScanIteratorWithPagination implements method in VersionedDB interface
func (vdb *versionedDB) GetStateRangeScanIteratorWithPagination(namespace string, startKey string, endKey string, pageSize int32) (statedb.QueryResultsIterator, error) {
	dataStartKey := encodeDataKey(namespace, startKey)
	dataEndKey := encodeDataKey(namespace, endKey)
	if endKey == "" {
		dataEndKey[len(dataEndKey)-1] = lastKeyIndicator
	}
	strStartKey := hex.EncodeToString(dataStartKey)
	strEndKey := hex.EncodeToString(dataEndKey)

	keylist := []string{}
	for key := range vdb.db {
		if key >= strStartKey && key <= strEndKey {
			keylist = append(keylist, key)
		}
	}
	sort.Strings(keylist)

	return newKVScanner(vdb, keylist), nil
	// return nil, errors.New("GetStateRangeScanIteratorWithPagination not supported for Memorydb")
}

// ExecuteQuery implements method in VersionedDB interface
func (vdb *versionedDB) ExecuteQuery(namespace, query string) (statedb.ResultsIterator, error) {
	return nil, errors.New("ExecuteQuery not supported for Memorydb")
}

// ExecuteQueryWithPagination implements method in VersionedDB interface
func (vdb *versionedDB) ExecuteQueryWithPagination(namespace, query, bookmark string, pageSize int32) (statedb.QueryResultsIterator, error) {
	return nil, errors.New("ExecuteQueryWithMetadata not supported for Memorydb")
}

// ApplyUpdates implements method in VersionedDB interface
func (vdb *versionedDB) ApplyUpdates(batch *statedb.UpdateBatch, height *version.Height) error {
	fmt.Printf("### In statememorydb.go/ApplyUpdates, height=%s ###\n", height)

	namespaces := batch.GetUpdatedNamespaces()
	for _, ns := range namespaces {
		updates := batch.GetUpdates(ns)
		for k, vv := range updates {
			dataKey := hex.EncodeToString(encodeDataKey(ns, k))
			fmt.Printf("### store: ns=%s, key=%s, datakey=%s ###\n", ns, k, dataKey)
			if vv.Value == nil {
				delete(vdb.db, dataKey)
			} else {
				encodedVal, err := stateleveldb.EncodeValue(vv)
				if err != nil {
					return err
				}
				vdb.db[dataKey] = encodedVal
			}
		}
	}
	return nil
}

// GetLatestSavePoint implements method in VersionedDB interface
func (vdb *versionedDB) GetLatestSavePoint() (*version.Height, error) {
	return nil, nil
}

// GetFullScanIterator implements method in VersionedDB interface.
func (vdb *versionedDB) GetFullScanIterator(skipNamespace func(string) bool) (statedb.FullScanIterator, error) {
	panic("Not implemented")
}

// importState implements method in VersionedDB interface. The function is expected to be used
// for importing the state from a previously snapshotted state. The parameter itr provides access to
// the snapshotted state.
func (vdb *versionedDB) importState(itr statedb.FullScanIterator, savepoint *version.Height) error {
	return nil
}

func (vdb *versionedDB) addDummyCounter() {
	counter, found := vdb.db["counter"]
	if found != true {
		vdb.db["counter"] = []byte("x")
	} else {
		vdb.db["counter"] = append(counter[:], []byte("x")...)
	}
}

func (vdb *versionedDB) printState() {
	fmt.Println("---- current in memory db state ----")
	for key, element := range vdb.db {
		hash := md5.Sum([]byte(element))
		fmt.Println("Key:", key, "=>", "Element:", hex.EncodeToString(hash[:]))
	}
}

func encodeDataKey(ns, key string) []byte {
	k := append(dataKeyPrefix, []byte(ns)...)
	k = append(k, nsKeySep...)
	return append(k, []byte(key)...)
}

func decodeDataKey(encodedDataKey []byte) (string, string) {
	split := bytes.SplitN(encodedDataKey, nsKeySep, 2)
	return string(split[0][1:]), string(split[1])
}

type kvScanner struct {
	vdb     *versionedDB
	keylist []string
	curNum  int
}

func newKVScanner(vdb *versionedDB, keylist []string) *kvScanner {
	return &kvScanner{
		vdb:     vdb,
		keylist: keylist,
		curNum:  0,
	}
}

func (kvs *kvScanner) Next() (*statedb.VersionedKV, error) {
	if kvs.curNum >= len(kvs.keylist) {
		return nil, nil
	}

	dbKey := kvs.keylist[kvs.curNum]
	dbVal, _ := kvs.vdb.db[dbKey]
	dbValCopy := make([]byte, len(dbVal))
	copy(dbValCopy, dbVal)
	dbKeyByte, err := hex.DecodeString(dbKey)

	ns, key := decodeDataKey(dbKeyByte)
	vv, err := stateleveldb.DecodeValue(dbValCopy)
	if err != nil {
		return nil, err
	}
	kvs.curNum++
	return &statedb.VersionedKV{
		CompositeKey: &statedb.CompositeKey{
			Namespace: ns,
			Key:       key,
		},
		VersionedValue: vv,
	}, nil
}

func (kvs *kvScanner) Close() {}

func (kvs *kvScanner) GetBookmarkAndClose() string {
	retval := ""
	if kvs.curNum < len(kvs.keylist) {
		dbKey := kvs.keylist[kvs.curNum]
		dbKeyByte, _ := hex.DecodeString(dbKey)
		_, key := decodeDataKey(dbKeyByte)
		retval = key
	}
	return retval
}
