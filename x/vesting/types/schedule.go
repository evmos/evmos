// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// ReadSchedule returns the value of a schedule at readTime.
//
// A "schedule" is an increasing step function of Coins over time. It's
// specified as an absolute start time and a sequence of relative periods, with
// each step at the end of a period. A schedule may also give the time and total
// value at the last step, which can speed evaluation of the step function after
// the last step.
func ReadSchedule(
	startTime, endTime int64,
	periods sdkvesting.Periods,
	totalCoins sdk.Coins,
	readTime int64,
) sdk.Coins {
	// return empty coins if the read time is before or equal start time
	if readTime <= startTime {
		return sdk.NewCoins()
	}
	// return the total coins when the read time is equal or after
	// end time
	if readTime >= endTime {
		return totalCoins
	}

	coins := sdk.Coins{} // sum of amounts for events before readTime
	elapsedTime := startTime

	for _, period := range periods {
		if readTime < elapsedTime+period.Length {
			// we're reading before the next event
			break
		}
		coins = coins.Add(period.Amount...)
		elapsedTime += period.Length
	}

	return coins
}

// ReadPastPeriodCount returns the amount of passed periods before read time
func ReadPastPeriodCount(
	startTime, endTime int64,
	periods sdkvesting.Periods,
	readTime int64,
) int {
	passedPeriods := 0

	// return 0 if the read time is before or equal start time
	if readTime <= startTime {
		return 0
	}

	// return all the periods when the read time is equal or after
	// end time
	if readTime >= endTime {
		return len(periods)
	}

	elapsedTime := startTime

	// for every period, add the period length to the elapsed time until
	// the read time is before the next period
	for _, period := range periods {
		if readTime < elapsedTime+period.Length {
			// we're reading before the next event
			break
		}
		passedPeriods++
		elapsedTime += period.Length
	}

	return passedPeriods
}

// DisjunctPeriods returns the union of two vesting period schedules.
// The returned schedule is the union of the vesting events.
// Simultaneous events are combined into a single event.
//
// Input schedules A and B are defined by their start times and periods.
// From this, the combined schedule with new start and end times,
// as well as the joined periods.
func DisjunctPeriods(
	startTimePeriodsA, startTimePeriodsB int64,
	periodsA, periodsB sdkvesting.Periods,
) (startTime, endTime int64, periods sdkvesting.Periods) {
	var (
		// These time vars store the time of the last merged event for each schedule.
		// When processing the next event in a schedule, the times are always relative to the previous event.
		timePeriodA  = startTimePeriodsA
		timePeriodsB = startTimePeriodsB

		// Initialize the period indices for both schedules.
		// These will be used to keep track of the processed events in each schedule.
		idxPeriodsA = 0
		idxPeriodsB = 0

		// Store the lengths of both schedules before adjusting any periods.
		lenPeriodsA = len(periodsA)
		lenPeriodsB = len(periodsB)
	)

	// The start time of the resulting schedule is determined
	// by the earlier of the two start times.
	startTime = Min64(startTimePeriodsA, startTimePeriodsB)

	// The end time is initially set to the start time. But will be updated as
	// the schedule events are processed.
	endTime = startTime

	// emit creates a new period, that spans the amount of seconds of the length
	// of the currently processed period minus the most recent endTime
	// (which is the end time of the last processed event).
	// This newly created period is appended to the slice of resulting periods.
	emit := func(nextTime int64, amount sdk.Coins) {
		period := sdkvesting.Period{
			Length: nextTime - endTime,
			Amount: amount,
		}
		periods = append(periods, period)
		endTime = nextTime
	}

	// consumeA processes the next period from schedule A.
	//
	// It appends the resulting periods with a new period of the same amount,
	// spanning the length of this processed period minus the most recent endTime.
	// It increments the period index and updates the last event time for A.
	consumeA := func(nextPeriodA int64) {
		emit(nextPeriodA, periodsA[idxPeriodsA].Amount)
		timePeriodA = nextPeriodA
		idxPeriodsA++
	}

	// consumeB processes the next period from schedule B.
	//
	// It appends the resulting periods with a new period of the same amount,
	// spanning the length of this processed period minus the most recent endTime.
	// It increments the period index and updates the last event time for B.
	consumeB := func(nextPeriodB int64) {
		emit(nextPeriodB, periodsB[idxPeriodsB].Amount)
		timePeriodsB = nextPeriodB
		idxPeriodsB++
	}

	// consumeBoth process the next periods from both schedules A and B.
	//
	// It adds the vesting amounts of both periods and appends the resulting periods with,
	// with a corresponding event of the length of the processed event.
	// It increments the period indices and updates the last event times for both schedules.
	consumeBoth := func(nextTime int64) {
		emit(nextTime, periodsA[idxPeriodsA].Amount.Add(periodsB[idxPeriodsB].Amount...))
		timePeriodA = nextTime
		timePeriodsB = nextTime
		idxPeriodsA++
		idxPeriodsB++
	}

	// Processing the schedules
	//
	// While there are more events in both schedules, handle the next one, merge
	// if concurrent
	for idxPeriodsA < lenPeriodsA && idxPeriodsB < lenPeriodsB {
		endTimeOfNextPeriodA := timePeriodA + periodsA[idxPeriodsA].Length  // next periodsA event in absolute time
		endTimeOfNextPeriodB := timePeriodsB + periodsB[idxPeriodsB].Length // next periodsB event in absolute time
		switch {
		case endTimeOfNextPeriodA < endTimeOfNextPeriodB:
			// if the next event in schedule A is before the next event in schedule B,
			// process the next event in schedule A
			consumeA(endTimeOfNextPeriodA)
		case endTimeOfNextPeriodA > endTimeOfNextPeriodB:
			// if the next event in schedule B is before the next event in schedule A,
			// process the next event in schedule B
			consumeB(endTimeOfNextPeriodB)
		default:
			// if the next event in schedule A and B are at the same time,
			// process both events at the same time
			consumeBoth(endTimeOfNextPeriodA)
		}
	}

	// consume remaining events in schedule Periods A
	for idxPeriodsA < lenPeriodsA {
		nextPeriodA := timePeriodA + periodsA[idxPeriodsA].Length
		consumeA(nextPeriodA)
	}
	// consume remaining events in schedule PeriodsB
	for idxPeriodsB < lenPeriodsB {
		nextPeriodB := timePeriodsB + periodsB[idxPeriodsB].Length
		consumeB(nextPeriodB)
	}

	return startTime, endTime, periods
}

// ConjunctPeriods returns the combination of two period schedules.
// The resulting schedule is the result is the minimum of the two schedules.
// It returns the resulting periods start and end times as well as the conjuncted
// periods.
func ConjunctPeriods(
	startTimePeriodA, startTimePeriodB int64,
	periodsA, periodsB sdkvesting.Periods,
) (startTime, endTime int64, conjunctionPeriods sdkvesting.Periods) {
	var (
		// These amount variables are keeping track of the amounts of coins in the different
		// vesting schedules.
		resultingAmount, totalAmountPeriodsA, totalAmountPeriodsB sdk.Coins

		// These time vars store the time of the last merged event for each schedule.
		// When processing the next event in a schedule, the times are always relative to the previous event.
		timePeriodsA = startTimePeriodA
		timePeriodsB = startTimePeriodB

		// Initialize the period indices for both schedules.
		// These will be used to keep track of the processed events in each schedule.
		idxPeriodsA = 0
		idxPeriodsB = 0

		// Store the lengths of both schedules before adjusting any periods.
		lenPeriodsA = len(periodsA)
		lenPeriodsB = len(periodsB)
	)

	// Initialize conjunction peridos
	conjunctionPeriods = sdkvesting.Periods{}

	// The start time of the resulting schedule is determined
	// by the earlier of the two start times.
	startTime = Min64(startTimePeriodA, startTimePeriodB)
	endTimeOfLastProcessedPeriod := startTime

	// emit creates a new period, that spans the amount of seconds between
	// the end of the processed period and the most recent stored time value
	// (=end of the last processed period).
	// This time value is updated to the current periods length.
	emit := func(endTimeOfCurrentPeriod int64, coins sdk.Coins) {
		period := sdkvesting.Period{
			Length: endTimeOfCurrentPeriod - endTimeOfLastProcessedPeriod,
			Amount: coins,
		}
		conjunctionPeriods = append(conjunctionPeriods, period)
		endTimeOfLastProcessedPeriod = endTimeOfCurrentPeriod
		resultingAmount = resultingAmount.Add(coins...)
	}

	// consumeA processes the next period from schedule A.
	//
	// It adds the amount of the processed period to the total processed amount for all periods of A.
	// It then calculates the minimum of the total amount of schedules A and B.
	// If the minimum of A and B is smaller than the resulting amount tallied so far,
	// it creates a new Period containing the difference between the minimum and the resulting amount.
	//
	// Additionally, it increments the period index and updates the last event time for schedule A.
	consumeA := func(endTimeOfCurrentPeriod int64) {
		totalAmountPeriodsA = totalAmountPeriodsA.Add(periodsA[idxPeriodsA].Amount...)
		min := totalAmountPeriodsA.Min(totalAmountPeriodsB)
		if resultingAmount.IsAllLTE(min) {
			diff := min.Sub(resultingAmount...)
			if !diff.IsZero() {
				emit(endTimeOfCurrentPeriod, diff)
			}
		}
		timePeriodsA = endTimeOfCurrentPeriod
		idxPeriodsA++
	}

	// consumeB processes the next period from schedule B.
	//
	// It adds the amount of the processed period to the total processed amount for all periods of B.
	// It then calculates the minimum of the total amount of schedules A and B.
	// If the minimum of A and B is smaller than the resulting amount tallied so far,
	// it creates a new Period containing the difference between the minimum and the resulting amount.
	//
	// Additionally, it increments the period index and updates the last event time for schedule B.
	consumeB := func(nextTime int64) {
		totalAmountPeriodsB = totalAmountPeriodsB.Add(periodsB[idxPeriodsB].Amount...)
		min := totalAmountPeriodsA.Min(totalAmountPeriodsB)
		if resultingAmount.IsAllLTE(min) {
			diff := min.Sub(resultingAmount...)
			if !diff.IsZero() {
				emit(nextTime, diff)
			}
		}
		timePeriodsB = nextTime
		idxPeriodsB++
	}

	// consumeBoth processes the next periods from both schedules A and B.
	//
	// It adds the amount of the processed period to the total processed amount for all periods of A and B.
	// It then calculates the minimum of the total amount of schedules A and B.
	// If the minimum of A and B is smaller than the resulting amount tallied so far,
	// it creates a new Period containing the difference between the minimum and the resulting amount.
	//
	// Additionally, it increments the period indices and updates the last event times for both schedules.
	consumeBoth := func(nextTime int64) {
		totalAmountPeriodsA = totalAmountPeriodsA.Add(periodsA[idxPeriodsA].Amount...)
		totalAmountPeriodsB = totalAmountPeriodsB.Add(periodsB[idxPeriodsB].Amount...)
		min := totalAmountPeriodsA.Min(totalAmountPeriodsB)
		if resultingAmount.IsAllLTE(min) {
			diff := min.Sub(resultingAmount...)
			if !diff.IsZero() {
				emit(nextTime, diff)
			}
		}
		timePeriodsA = nextTime
		timePeriodsB = nextTime
		idxPeriodsA++
		idxPeriodsB++
	}

	// Processing the schedules
	//
	// While there are more events in both schedules, handle the next one, merge
	// if concurrent
	for idxPeriodsA < lenPeriodsA && idxPeriodsB < lenPeriodsB {
		nextPeriodA := timePeriodsA + periodsA[idxPeriodsA].Length // next periods A event in absolute time
		nextPeriodB := timePeriodsB + periodsB[idxPeriodsB].Length // next periods B event in absolute time
		switch {
		case nextPeriodA < nextPeriodB:
			// if the next event in schedule A is before the next event in schedule B,
			// process the next event in schedule A
			consumeA(nextPeriodA)
		case nextPeriodA > nextPeriodB:
			// if the next event in schedule B is before the next event in schedule A,
			// process the next event in schedule B
			consumeB(nextPeriodB)
		default:
			// if the next event in schedule A and B are at the same time,
			// process both events at the same time
			consumeBoth(nextPeriodA)
		}
	}

	// consume remaining events in schedule A
	for idxPeriodsA < lenPeriodsA {
		nextPeriodA := timePeriodsA + periodsA[idxPeriodsA].Length
		consumeA(nextPeriodA)
	}

	// consume remaining events in schedule B
	for idxPeriodsB < lenPeriodsB {
		nextPeriodB := timePeriodsB + periodsB[idxPeriodsB].Length
		consumeB(nextPeriodB)
	}

	return startTime, endTimeOfLastProcessedPeriod, conjunctionPeriods
}

// AlignSchedules extends the first period's length to align the two given periods
// to the same start time. The earliest start time is chosen.
// It returns the aligned new start and end times of the periods.
func AlignSchedules(
	startTimePeriodA,
	startTimePeriodB int64,
	periodsA, periodsB sdkvesting.Periods,
) (startTime, endTime int64) {
	startTime = Min64(startTimePeriodA, startTimePeriodB)

	// add the difference time between the two periods
	if len(periodsA) > 0 {
		periodsA[0].Length += startTimePeriodA - startTime
	}

	if len(periodsB) > 0 {
		periodsB[0].Length += startTimePeriodB - startTime
	}

	endPeriodsA := startTime + periodsA.TotalLength()
	endPeriodsB := startTime + periodsB.TotalLength()

	// the end time between the 2 periods is the max length
	endTime = Max64(endPeriodsA, endPeriodsB)

	return startTime, endTime
}
