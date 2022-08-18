package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

type VestingData struct {
	StartTime int64         `json:"start_time"`
	Periods   []InputPeriod `json:"periods"`
}

type InputPeriod struct {
	Coins  string `json:"coins"`
	Length int64  `json:"length_seconds"`
}

// readScheduleFile reads the file at path and unmarshals it to get the schedule.
// Returns start time, periods, and error.
func ReadScheduleFile(path string) (int64, sdkvesting.Periods, error) {
	contents, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return 0, nil, err
	}

	var data VestingData

	if err = json.Unmarshal(contents, &data); err != nil {
		return 0, nil, err
	}

	startTime := data.StartTime
	periods := make(sdkvesting.Periods, 0, len(data.Periods))

	for i, p := range data.Periods {
		if p.Length < 1 {
			return 0, nil, fmt.Errorf("invalid period length of %d in period %d, length must be greater than 0", p.Length, i)
		}

		amount, err := sdk.ParseCoinsNormalized(p.Coins)
		if err != nil {
			return 0, nil, err
		}

		period := sdkvesting.Period{Length: p.Length, Amount: amount}
		periods = append(periods, period)
	}

	return startTime, periods, nil
}
