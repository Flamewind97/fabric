/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endorser

import (
	commonledger "github.com/hyperledger/fabric/common/ledger"
	"github.com/hyperledger/fabric/core/ledger"
)

type MaliciousPeerWrapper struct {
	txSimulator ledger.TxSimulator
}

// GetState gets the value for given namespace and key. For a chaincode, the namespace corresponds to the chaincodeId
func (m *MaliciousPeerWrapper) GetState(namespace string, key string) ([]byte, error) {
	maliciousLogger.Warningf("Performing GetState start: namespace=%s, key=%s", namespace, key)
	res, err := m.txSimulator.GetState(namespace, key)
	maliciousLogger.Warningf("Performing GetState end: res=%s, err=%s", res, err)
	return res, err
}

// GetStateRangeScanIterator returns an iterator that contains all the key-values between given key ranges.
// startKey is included in the results and endKey is excluded. An empty startKey refers to the first available key
// and an empty endKey refers to the last available key. For scanning all the keys, both the startKey and the endKey
// can be supplied as empty strings. However, a full scan should be used judiciously for performance reasons.
// The returned ResultsIterator contains results of type *KV which is defined in fabric-protos/ledger/queryresult.
func (m *MaliciousPeerWrapper) GetStateRangeScanIterator(namespace string, startKey string, endKey string) (commonledger.ResultsIterator, error) {
	return m.txSimulator.GetStateRangeScanIterator(namespace, startKey, endKey)
}

// GetPrivateDataHash gets the hash of the value of a private data item identified by a tuple <namespace, collection, key>
// Function `GetPrivateData` is only meaningful when it is invoked on a peer that is authorized to have the private data
// for the collection <namespace, collection>. However, the function `GetPrivateDataHash` can be invoked on any peer
// to get the hash of the current value
func (m *MaliciousPeerWrapper) GetPrivateDataHash(namespace string, collection string, key string) ([]byte, error) {
	return m.txSimulator.GetPrivateDataHash(namespace, collection, key)
}

// GetStateMetadata returns the metadata for given namespace and key
func (m *MaliciousPeerWrapper) GetStateMetadata(namespace string, key string) (map[string][]byte, error) {
	return m.txSimulator.GetStateMetadata(namespace, key)
}

// GetStateMultipleKeys gets the values for multiple keys in a single call
func (m *MaliciousPeerWrapper) GetStateMultipleKeys(namespace string, keys []string) ([][]byte, error) {
	return m.txSimulator.GetStateMultipleKeys(namespace, keys)
}

// GetStateRangeScanIteratorWithPagination returns an iterator that contains all the key-values between given key ranges.
// startKey is included in the results and endKey is excluded. An empty startKey refers to the first available key
// and an empty endKey refers to the last available key. For scanning all the keys, both the startKey and the endKey
// can be supplied as empty strings. However, a full scan should be used judiciously for performance reasons.
// The page size parameter limits the number of returned results.
// The returned ResultsIterator contains results of type *KV which is defined in fabric-protos/ledger/queryresult.
func (m *MaliciousPeerWrapper) GetStateRangeScanIteratorWithPagination(namespace string, startKey string, endKey string, pageSize int32) (ledger.QueryResultsIterator, error) {
	return m.txSimulator.GetStateRangeScanIteratorWithPagination(namespace, startKey, endKey, pageSize)
}

// ExecuteQuery executes the given query and returns an iterator that contains results of type specific to the underlying data store.
// Only used for state databases that support query
// For a chaincode, the namespace corresponds to the chaincodeId
// The returned ResultsIterator contains results of type *KV which is defined in fabric-protos/ledger/queryresult.
func (m *MaliciousPeerWrapper) ExecuteQuery(namespace string, query string) (commonledger.ResultsIterator, error) {
	return m.txSimulator.ExecuteQuery(namespace, query)
}

// ExecuteQueryWithPagination executes the given query and returns an iterator that contains results of type specific to the underlying data store.
// The bookmark and page size parameters are associated with the pagination.
// Only used for state databases that support query
// For a chaincode, the namespace corresponds to the chaincodeId
// The returned ResultsIterator contains results of type *KV which is defined in fabric-protos/ledger/queryresult.
func (m *MaliciousPeerWrapper) ExecuteQueryWithPagination(namespace string, query string, bookmark string, pageSize int32) (ledger.QueryResultsIterator, error) {
	return m.txSimulator.ExecuteQueryWithPagination(namespace, query, bookmark, pageSize)
}

// GetPrivateData gets the value of a private data item identified by a tuple <namespace, collection, key>
func (m *MaliciousPeerWrapper) GetPrivateData(namespace string, collection string, key string) ([]byte, error) {
	return m.txSimulator.GetPrivateData(namespace, collection, key)
}

// GetPrivateDataMetadata gets the metadata of a private data item identified by a tuple <namespace, collection, key>
func (m *MaliciousPeerWrapper) GetPrivateDataMetadata(namespace string, collection string, key string) (map[string][]byte, error) {
	return m.txSimulator.GetPrivateDataMetadata(namespace, collection, key)
}

// GetPrivateDataMetadataByHash gets the metadata of a private data item identified by a tuple <namespace, collection, keyhash>
func (m *MaliciousPeerWrapper) GetPrivateDataMetadataByHash(namespace string, collection string, keyhash []byte) (map[string][]byte, error) {
	return m.txSimulator.GetPrivateDataMetadataByHash(namespace, collection, keyhash)
}

// GetPrivateDataMultipleKeys gets the values for the multiple private data items in a single call
func (m *MaliciousPeerWrapper) GetPrivateDataMultipleKeys(namespace string, collection string, keys []string) ([][]byte, error) {
	return m.txSimulator.GetPrivateDataMultipleKeys(namespace, collection, keys)
}

// GetPrivateDataRangeScanIterator returns an iterator that contains all the key-values between given key ranges.
// startKey is included in the results and endKey is excluded. An empty startKey refers to the first available key
// and an empty endKey refers to the last available key. For scanning all the keys, both the startKey and the endKey
// can be supplied as empty strings. However, a full scan shuold be used judiciously for performance reasons.
// The returned ResultsIterator contains results of type *KV which is defined in fabric-protos/ledger/queryresult.
func (m *MaliciousPeerWrapper) GetPrivateDataRangeScanIterator(namespace string, collection string, startKey string, endKey string) (commonledger.ResultsIterator, error) {
	return m.txSimulator.GetPrivateDataRangeScanIterator(namespace, collection, startKey, endKey)
}

// ExecuteQuery executes the given query and returns an iterator that contains results of type specific to the underlying data store.
// Only used for state databases that support query
// For a chaincode, the namespace corresponds to the chaincodeId
// The returned ResultsIterator contains results of type *KV which is defined in fabric-protos/ledger/queryresult.
func (m *MaliciousPeerWrapper) ExecuteQueryOnPrivateData(namespace string, collection string, query string) (commonledger.ResultsIterator, error) {
	return m.txSimulator.ExecuteQueryOnPrivateData(namespace, collection, query)
}

// Done releases resources occupied by the QueryExecutor
func (m *MaliciousPeerWrapper) Done() {
	m.txSimulator.Done()
}

// SetState sets the given value for the given namespace and key. For a chaincode, the namespace corresponds to the chaincodeId
func (m *MaliciousPeerWrapper) SetState(namespace string, key string, value []byte) error {
	return m.txSimulator.SetState(namespace, key, value)
}

// DeleteState deletes the given namespace and key
func (m *MaliciousPeerWrapper) DeleteState(namespace string, key string) error {
	return m.txSimulator.DeleteState(namespace, key)
}

// SetMultipleKeys sets the values for multiple keys in a single call
func (m *MaliciousPeerWrapper) SetStateMultipleKeys(namespace string, kvs map[string][]byte) error {
	return m.txSimulator.SetStateMultipleKeys(namespace, kvs)
}

// SetStateMetadata sets the metadata associated with an existing key-tuple <namespace, key>
func (m *MaliciousPeerWrapper) SetStateMetadata(namespace string, key string, metadata map[string][]byte) error {
	return m.txSimulator.SetStateMetadata(namespace, key, metadata)
}

// DeleteStateMetadata deletes the metadata (if any) associated with an existing key-tuple <namespace, key>
func (m *MaliciousPeerWrapper) DeleteStateMetadata(namespace string, key string) error {
	return m.txSimulator.DeleteStateMetadata(namespace, key)
}

// ExecuteUpdate for supporting rich data model (see comments on QueryExecutor above)
func (m *MaliciousPeerWrapper) ExecuteUpdate(query string) error {
	return m.txSimulator.ExecuteUpdate(query)
}

// SetPrivateData sets the given value to a key in the private data state represented by the tuple <namespace, collection, key>
func (m *MaliciousPeerWrapper) SetPrivateData(namespace string, collection string, key string, value []byte) error {
	return m.txSimulator.SetPrivateData(namespace, collection, key, value)
}

// SetPrivateDataMultipleKeys sets the values for multiple keys in the private data space in a single call
func (m *MaliciousPeerWrapper) SetPrivateDataMultipleKeys(namespace string, collection string, kvs map[string][]byte) error {
	return m.txSimulator.SetPrivateDataMultipleKeys(namespace, collection, kvs)
}

// DeletePrivateData deletes the given tuple <namespace, collection, key> from private data
func (m *MaliciousPeerWrapper) DeletePrivateData(namespace string, collection string, key string) error {
	return m.txSimulator.DeletePrivateData(namespace, collection, key)
}

// SetPrivateDataMetadata sets the metadata associated with an existing key-tuple <namespace, collection, key>
func (m *MaliciousPeerWrapper) SetPrivateDataMetadata(namespace string, collection string, key string, metadata map[string][]byte) error {
	return m.txSimulator.SetPrivateDataMetadata(namespace, collection, key, metadata)
}

// DeletePrivateDataMetadata deletes the metadata associated with an existing key-tuple <namespace, collection, key>
func (m *MaliciousPeerWrapper) DeletePrivateDataMetadata(namespace string, collection string, key string) error {
	return m.txSimulator.DeletePrivateDataMetadata(namespace, collection, key)
}

// GetTxSimulationResults encapsulates the results of the transaction simulation.
// This should contain enough detail for
//   - The update in the state that would be caused if the transaction is to be committed
//   - The environment in which the transaction is executed so as to be able to decide the validity of the environment
//     (at a later time on a different peer) during committing the transactions
//
// Different ledger implementation (or configurations of a single implementation) may want to represent the above two pieces
// of information in different way in order to support different data-models or optimize the information representations.
// Returned type 'TxSimulationResults' contains the simulation results for both the public data and the private data.
// The public data simulation results are expected to be used as in V1 while the private data simulation results are expected
// to be used by the gossip to disseminate this to the other endorsers (in phase-2 of sidedb)
func (m *MaliciousPeerWrapper) GetTxSimulationResults() (*ledger.TxSimulationResults, error) {
	return m.txSimulator.GetTxSimulationResults()
}
