/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pvtstatepurgemgmt

import (
	fmt "fmt"
	math "math"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/privacyenabledstate"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/hyperledger/fabric/core/ledger/pvtdatapolicy"
	"github.com/hyperledger/fabric/core/ledger/util"
)

type expiryScheduleBuilder struct {
	btlPolicy       pvtdatapolicy.BTLPolicy
	scheduleEntries map[expiryInfoKey]*PvtdataKeys
}

func newExpiryScheduleBuilder(btlPolicy pvtdatapolicy.BTLPolicy) *expiryScheduleBuilder {
	return &expiryScheduleBuilder{btlPolicy, make(map[expiryInfoKey]*PvtdataKeys)}
}

func (builder *expiryScheduleBuilder) add(ns, coll, key string, keyHash []byte, versionedValue *statedb.VersionedValue) error {
	fmt.Printf("--- In expiry_schedule_builder.go, add ---\n")
	committingBlk := versionedValue.Version.BlockNum
	expiryBlk, err := builder.btlPolicy.GetExpiringBlock(ns, coll, committingBlk)
	if err != nil {
		fmt.Printf("--- In expiry_schedule_builder.go, getExpiringBlock failed, err %s---\n", err)
		return err
	}
	if isDelete(versionedValue) || neverExpires(expiryBlk) {
		fmt.Printf("--- In expiry_schedule_builder.go, isDelete or neverExpires failed, err %s---\n", err)
		return nil
	}
	expinfoKey := expiryInfoKey{committingBlk: committingBlk, expiryBlk: expiryBlk}
	pvtdataKeys, ok := builder.scheduleEntries[expinfoKey]
	if !ok {
		pvtdataKeys = newPvtdataKeys()
		builder.scheduleEntries[expinfoKey] = pvtdataKeys
	}
	fmt.Printf("--- In expiry_schedule_builder.go, pvtdataKeys.add, ok = %t ---\n", ok)
	pvtdataKeys.add(ns, coll, key, keyHash)
	return nil
}

func (builder *expiryScheduleBuilder) getExpiryInfo() []*expiryInfo {
	fmt.Printf("--- In expiry_schedule_builder.go, getExpiryInfo ---\n")
	var listExpinfo []*expiryInfo
	for expinfoKey, pvtdataKeys := range builder.scheduleEntries {
		expinfoKeyCopy := expinfoKey
		listExpinfo = append(listExpinfo, &expiryInfo{expiryInfoKey: &expinfoKeyCopy, pvtdataKeys: pvtdataKeys})
	}
	return listExpinfo
}

func buildExpirySchedule(
	btlPolicy pvtdatapolicy.BTLPolicy,
	pvtUpdates *privacyenabledstate.PvtUpdateBatch,
	hashedUpdates *privacyenabledstate.HashedUpdateBatch) ([]*expiryInfo, error) {

	fmt.Printf("--- In expiry_schedule_builder.go, buildExpirySchedule ---\n")
	hashedUpdateKeys := hashedUpdates.ToCompositeKeyMap()
	expiryScheduleBuilder := newExpiryScheduleBuilder(btlPolicy)

	logger.Debugf("Building the expiry schedules based on the update batch")

	// Iterate through the private data updates and for each key add into the expiry schedule
	// i.e., when these private data key and it's hashed-keys are going to be expired
	// Note that the 'hashedUpdateKeys'  may be superset of the pvtUpdates. This is because,
	// the peer may not receive all the private data either because the peer is not eligible for certain private data
	// or because we allow proceeding with the missing private data data
	for pvtUpdateKey, vv := range pvtUpdates.ToCompositeKeyMap() {
		keyHash := util.ComputeStringHash(pvtUpdateKey.Key)
		hashedCompisiteKey := privacyenabledstate.HashedCompositeKey{
			Namespace:      pvtUpdateKey.Namespace,
			CollectionName: pvtUpdateKey.CollectionName,
			KeyHash:        string(keyHash),
		}
		logger.Debugf("Adding expiry schedule for key and key hash [%s]", &hashedCompisiteKey)
		if err := expiryScheduleBuilder.add(pvtUpdateKey.Namespace, pvtUpdateKey.CollectionName, pvtUpdateKey.Key, keyHash, vv); err != nil {
			fmt.Printf("--- In expiry_schedule_builder.go, buildExpirySchedule, pvtUpdates expiryScheduleBuilder add failed, err %s ---\n", err)
			return nil, err
		}
		delete(hashedUpdateKeys, hashedCompisiteKey)
	}

	// Add entries for the leftover key hashes i.e., the hashes corresponding to which there is not private key is present
	for hashedUpdateKey, vv := range hashedUpdateKeys {
		logger.Debugf("Adding expiry schedule for key hash [%s]", &hashedUpdateKey)
		if err := expiryScheduleBuilder.add(hashedUpdateKey.Namespace, hashedUpdateKey.CollectionName, "", []byte(hashedUpdateKey.KeyHash), vv); err != nil {
			fmt.Printf("--- In expiry_schedule_builder.go, buildExpirySchedule, hashedUpdateKeys expiryScheduleBuilder add failed, err %s ---\n", err)
			return nil, err
		}
	}
	return expiryScheduleBuilder.getExpiryInfo(), nil
}

func isDelete(versionedValue *statedb.VersionedValue) bool {
	return versionedValue.Value == nil
}

func neverExpires(expiryBlk uint64) bool {
	return expiryBlk == math.MaxUint64
}
