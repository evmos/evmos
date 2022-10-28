package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// A "schedule" is an increasing step function of Coins over time. It's
// specified as an absolute start time and a sequence of relative periods, with
// each step at the end of a period. A schedule may also give the time and total
// value at the last step, which can speed evaluation of the step function after
// the last step.
//
// ReadSchedule returns the value of a schedule at readTime.
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

// DisjunctPeriods returns the union of two vesting period schedules. The
// returned schedule is the union of the vesting events, with simultaneous
// events combined into a single event. Input schedules P and Q are defined by
// their start times and periods. Returns new start time, new end time, and
// merged vesting events, relative to the new start time.
func DisjunctPeriods(
	startTimePeriodsA, startTimePeriodsB int64,
	periodsA, periodsB sdkvesting.Periods,
) (startTime, endTime int64, periods sdkvesting.Periods) {
	timePeriodA := startTimePeriodsA  // time of last merged periods A event, next p event is relative to this time
	timePeriodsB := startTimePeriodsB // time of last merged periods B event, next periodsB event is relative to this time

	idxPeriodsA := 0 // periods A indexes before this have been merged
	idxPeriodsB := 0 // periods B indexes before this have been merged

	lenPeriodsA := len(periodsA)
	lenPeriodsB := len(periodsB)
	startTime = Min64(startTimePeriodsA, startTimePeriodsB) // we pick the earlier time
	endTime = startTime                                     // time of last merged event, or the start time
	periods = sdkvesting.Periods{}                          // merged periods

	// emit adds an output period and updates the last event time
	emit := func(nextTime int64, amount sdk.Coins) {
		period := sdkvesting.Period{
			Length: nextTime - endTime,
			Amount: amount,
		}
		periods = append(periods, period)
		endTime = nextTime
	}

	// consumeA emits the next period from A, updating indexes
	consumeA := func(nextPeriodA int64) {
		emit(nextPeriodA, periodsA[idxPeriodsA].Amount)
		timePeriodA = nextPeriodA
		idxPeriodsA++
	}

	// consumeB emits the next period from B, updating indexes
	consumeB := func(nextPeriodB int64) {
		emit(nextPeriodB, periodsB[idxPeriodsB].Amount)
		timePeriodsB = nextPeriodB
		idxPeriodsB++
	}

	// consumeBoth emits a merge of the next periods from periods A and B, updating indexes
	consumeBoth := func(nextTime int64) {
		emit(nextTime, periodsA[idxPeriodsA].Amount.Add(periodsB[idxPeriodsB].Amount...))
		timePeriodA = nextTime
		timePeriodsB = nextTime
		idxPeriodsA++
		idxPeriodsB++
	}

	// while there are more events in both schedules, handle the next one, merge
	// if concurrent
	for idxPeriodsA < lenPeriodsA && idxPeriodsB < lenPeriodsB {
		nextPeriodA := timePeriodA + periodsA[idxPeriodsA].Length  // next periodsA event in absolute time
		nextPeriodB := timePeriodsB + periodsB[idxPeriodsB].Length // next periodsB event in absolute time
		switch {
		case nextPeriodA < nextPeriodB:
			consumeA(nextPeriodA)
		case nextPeriodA > nextPeriodB:
			consumeB(nextPeriodB)
		default:
			consumeBoth(nextPeriodA)
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

// ConjunctPeriods returns the combination of two period schedules where the
// result is the minimum of the two schedules.
// It returns the resulting periods start and end times as well as the resulting
// conjunction periods.
// TODO: rename and add comprehensive comments, this is currently not maintainable
func ConjunctPeriods(
	startTimePeriodA, startTimePeriodB int64,
	periodsA, periodsB sdkvesting.Periods,
) (startTime, endTime int64, conjunctionPeriods sdkvesting.Periods) {
	timePeriodsA := startTimePeriodA
	timePeriodsB := startTimePeriodB
	idxPeriodsA := 0
	idxPeriodsB := 0
	lenPeriodsA := len(periodsA)
	lenPeriodsB := len(periodsB)
	startTime = Min64(startTimePeriodA, startTimePeriodB)
	time := startTime

	conjunctionPeriods = sdkvesting.Periods{}
	amount := sdk.Coins{}

	totalAmountPeriodsA := amount
	totalAmountPeriodsB := amount

	// emit adds an output period and updates the last event time
	emit := func(nextTime int64, coins sdk.Coins) {
		period := sdkvesting.Period{
			Length: nextTime - time,
			Amount: coins,
		}
		conjunctionPeriods = append(conjunctionPeriods, period)
		time = nextTime
		amount = amount.Add(coins...)
	}

	// consumeA processes the next event in P and emits an event
	// if the minimum of P and Q changes
	consumeA := func(nextTime int64) {
		totalAmountPeriodsA = totalAmountPeriodsA.Add(periodsA[idxPeriodsA].Amount...)
		min := totalAmountPeriodsA.Min(totalAmountPeriodsB)
		if amount.IsAllLTE(min) {
			diff := min.Sub(amount...)
			if !diff.IsZero() {
				emit(nextTime, diff)
			}
		}
		timePeriodsA = nextTime
		idxPeriodsA++
	}

	// consumeB processes the next event in Q and emits an event
	// if the minimum of P and Q changes
	consumeB := func(nextTime int64) {
		totalAmountPeriodsB = totalAmountPeriodsB.Add(periodsB[idxPeriodsB].Amount...)
		min := totalAmountPeriodsA.Min(totalAmountPeriodsB)
		if amount.IsAllLTE(min) {
			diff := min.Sub(amount...)
			if !diff.IsZero() {
				emit(nextTime, diff)
			}
		}
		timePeriodsB = nextTime
		idxPeriodsB++
	}

	// consumeBoth processes simultaneous events in P and Q and emits an
	// event if the minimum of P and Q changes
	consumeBoth := func(nextTime int64) {
		totalAmountPeriodsA = totalAmountPeriodsA.Add(periodsA[idxPeriodsA].Amount...)
		totalAmountPeriodsB = totalAmountPeriodsB.Add(periodsB[idxPeriodsB].Amount...)
		min := totalAmountPeriodsA.Min(totalAmountPeriodsB)
		if amount.IsAllLTE(min) {
			diff := min.Sub(amount...)
			if !diff.IsZero() {
				emit(nextTime, diff)
			}
		}
		timePeriodsA = nextTime
		timePeriodsB = nextTime
		idxPeriodsA++
		idxPeriodsB++
	}

	// while there are events left in both schedules, process the next one
	for idxPeriodsA < lenPeriodsA && idxPeriodsB < lenPeriodsB {
		nextPeriodA := timePeriodsA + periodsA[idxPeriodsA].Length // next periods A event in absolute time
		nextPeriodB := timePeriodsB + periodsB[idxPeriodsB].Length // next periods B event in absolute time
		switch {
		case nextPeriodA < nextPeriodB:
			consumeA(nextPeriodA)
		case nextPeriodA > nextPeriodB:
			consumeB(nextPeriodB)
		default:
			consumeBoth(nextPeriodA)
		}
	}

	// consume remaining events in schedule P
	for idxPeriodsA < lenPeriodsA {
		nextPeriodA := timePeriodsA + periodsA[idxPeriodsA].Length
		consumeA(nextPeriodA)
	}

	// consume remaining events in schedule Q
	for idxPeriodsB < lenPeriodsB {
		nextPeriodB := timePeriodsB + periodsB[idxPeriodsB].Length
		consumeB(nextPeriodB)
	}

	endTime = time
	return startTime, endTime, conjunctionPeriods
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
