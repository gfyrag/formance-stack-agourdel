package core

import (
	"sort"

	"github.com/formancehq/stack/libs/go-libs/collectionutils"
	"github.com/formancehq/stack/libs/go-libs/metadata"
)

const (
	formanceNamespace         = "com.formance.spec/"
	revertKey                 = "state/reverts"
	revertedKey               = "state/reverted"
	MetaTargetTypeAccount     = "ACCOUNT"
	MetaTargetTypeTransaction = "TRANSACTION"
)

func SpecMetadata(name string) string {
	return formanceNamespace + name
}

func MarkReverts(m metadata.Metadata, txID string) metadata.Metadata {
	return m.Merge(RevertMetadata(txID))
}

func RevertedMetadataSpecKey() string {
	return SpecMetadata(revertedKey)
}

func RevertMetadataSpecKey() string {
	return SpecMetadata(revertKey)
}

func ComputeMetadata(key, value string) metadata.Metadata {
	return metadata.Metadata{
		key: value,
	}
}

func RevertedMetadata(by string) metadata.Metadata {
	return ComputeMetadata(RevertedMetadataSpecKey(), by)
}

func RevertMetadata(tx string) metadata.Metadata {
	return ComputeMetadata(RevertMetadataSpecKey(), tx)
}

func IsReverted(m metadata.Metadata) bool {
	if _, ok := m[RevertedMetadataSpecKey()]; ok {
		return true
	}
	return false
}

func hashStringMetadata(buf *buffer, m metadata.Metadata) {
	if len(m) == 0 {
		return
	}
	keysOfAccount := collectionutils.Keys(m)
	if len(m) > 1 {
		sort.Strings(keysOfAccount)
	}
	for _, key := range keysOfAccount {
		buf.writeString(key)
		buf.writeString(m[key])
	}
}
