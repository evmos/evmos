package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StoreUpgrades defines a series of transformations to apply the multistore db upon load
type StoreUpgrades struct {
	Added    []string                 `json:"added"`
	Renamed  []storetypes.StoreRename `json:"renamed"`
	Deleted  []string                 `json:"deleted"`
	Replaced []storetypes.StoreRename `json:"replaced"`
}

// IsAdded returns true if the given key should be added
func (s *StoreUpgrades) IsAdded(key string) bool {
	if s == nil {
		return false
	}
	for _, added := range s.Added {
		if key == added {
			return true
		}
	}
	return false
}

// IsDeleted returns true if the given key should be deleted
func (s *StoreUpgrades) IsDeleted(key string) bool {
	if s == nil {
		return false
	}
	for _, d := range s.Deleted {
		if d == key {
			return true
		}
	}
	return false
}

// ReplacedFrom returns the oldKey if it was replaced
// Returns "" if it was not renamed
func (s *StoreUpgrades) ReplacedFrom(key string) string {
	if s == nil {
		return ""
	}
	for _, re := range s.Replaced {
		if re.NewKey == key {
			return re.OldKey
		}
	}
	return ""
}

// RenamedFrom returns the oldKey if it was renamed
// Returns "" if it was not renamed
func (s *StoreUpgrades) RenamedFrom(key string) string {
	if s == nil {
		return ""
	}
	for _, re := range s.Renamed {
		if re.NewKey == key {
			return re.OldKey
		}
	}
	return ""
}

// UpgradeStoreLoader is used to prepare baseapp with a fixed StoreLoader
// pattern. This is useful for custom upgrade loading logic.
func UpgradeStoreLoader(upgradeHeight int64, storeUpgrades *StoreUpgrades) baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {
		if upgradeHeight == ms.LastCommitID().Version+1 {
			storeUpgrade := &storetypes.StoreUpgrades{}

			if len(storeUpgrades.Replaced) > 0 {
				// delete and readd by calling
				// copy sliced: added, deleted, renamed
				// add all the Replaced[i].OldKeys to the Deleted slice

				storeUpgradeReplace := &storetypes.StoreUpgrades{}
				copy(storeUpgradeReplace.Added, storeUpgrades.Added)
				copy(storeUpgradeReplace.Deleted, storeUpgrades.Deleted)
				copy(storeUpgradeReplace.Renamed, storeUpgrades.Renamed)

				for _, s := range storeUpgrades.Replaced {
					storeUpgradeReplace.Deleted = append(storeUpgrade.Deleted, s.OldKey)
					storeUpgrade.Added = append(storeUpgrade.Added, s.NewKey)
				}

				if err := ms.LoadLatestVersionAndUpgrade(storeUpgradeReplace); err != nil {
					return err
				}

				// add all the Replaced[i].NewKeys to the Added slice
				// storeUpgrade.Added = ...
			}

			// Check if the current commit version and upgrade height matches
			if len(storeUpgrades.Renamed) > 0 || len(storeUpgrades.Deleted) > 0 || len(storeUpgrades.Added) > 0 {
				return ms.LoadLatestVersionAndUpgrade(storeUpgrade)
			}
		}

		// Otherwise load default store loader
		return baseapp.DefaultStoreLoader(ms)
	}
}
