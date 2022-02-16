# `vestcalc`: A vesting schedule calculator

A periodic vesting account has its vesting schedule configured as a sequence
of vesting events, spaced by the relative time between them, in seconds.
Most vesting agreements, however, are specified in terms of a number of
monthly events from a given start time, possibly subject to one or more
vesting "cliffs" which delay vesting until at or after the cliff.

This tool can generate a vesting schedule given the parameters above,
and can translate a vesting schedule into readable timestamps.

This tool correctly handles:

- clipping event dates to the end of the month (e.g. vesting on the 31st of
  the month happens on the 30th in June);
- daylight saving time;
- leap years;
- gigantic amounts (up to 255-bit);
- multiple denominations.

Times are interpreted in the local timezone unless explicitly overridden,
since the desired vesting schedule is commonly specified in local time.
To use another timezone, set your `TZ` environment variable before
running the command.

## Build and install

Run `go install` in this directory, which will create or update the
`vestcalc` binary in (by default) your `~/go/bin` directory. See the
[documentation](https://pkg.go.dev/cmd/go) for the `go` command-line
tool for other options.

## Writing a schedule

When the `--write` flag is set, the tool will write a schedule in JSON to
stdout. The following flags control the output:

- `--coins:` The coins to vest, e.g. `100ubld,50urun`.
- `--months`: The number of months to vest over.
- `--time`: The time of day of the vesting event, in 24-hour HH:MM format.
  Defaults to midnight.
- `--start`: The vesting start time: i.e. the first event happens in the
  next month. Specified in the format `YYYY-MM-DD` or `YYYY-MM-DDThh:mm`,
  e.g. `2006-01-02T15:04` for 3:04pm on January 2, 2006.
- `--cliffs`: One or more vesting cliffs in `YYYY-MM-DD` or `YYYY-MM-DDThh:mm`
  format. Only the latest one will have any effect, but it is useful to let
  the computer do that calculation to avoid mistakes. Multiple cliff dates
  can be separated by commas or given as multiple arguments.

## Reading a schedule

When the `--read` flag is set, the tool will read a schedule in JSON from
stdin and write the vesting events in absolute time to stdout.

## Examples

```
$ vestcalc --write --start=2021-01-01 --coins=1000000000ubld \
> --months=24 --time=09:00 --cliffs=2022-01-15T00:00 | \
> vestcalc --read
[
    2022-01-15T00:00: 500000000ubld
    2022-02-01T09:00: 41666666ubld
    2022-03-01T09:00: 41666667ubld
    2022-04-01T09:00: 41666667ubld
    2022-05-01T09:00: 41666666ubld
    2022-06-01T09:00: 41666667ubld
    2022-07-01T09:00: 41666667ubld
    2022-08-01T09:00: 41666666ubld
    2022-09-01T09:00: 41666667ubld
    2022-10-01T09:00: 41666667ubld
    2022-11-01T09:00: 41666666ubld
    2022-12-01T09:00: 41666667ubld
    2023-01-01T09:00: 41666667ubld
]
$ vestcalc --write --start=2021-01-01 --coins=1000000000ubld \
> --months=24 --time=09:00 --cliffs=2022-01-15T00:00
{
  "start_time": 1609488000,
  "periods": [
    {
      "coins": "500000000ubld",
      "length_seconds": 32745600
    },
    {
      "coins": "41666666ubld",
      "length_seconds": 1501200
    },
    {
      "coins": "41666667ubld",
      "length_seconds": 2419200
    },
    {
      "coins": "41666667ubld",
      "length_seconds": 2674800
    },
    {
      "coins": "41666666ubld",
      "length_seconds": 2592000
    },
    {
      "coins": "41666667ubld",
      "length_seconds": 2678400
    },
    {
      "coins": "41666667ubld",
      "length_seconds": 2592000
    },
    {
      "coins": "41666666ubld",
      "length_seconds": 2678400
    },
    {
      "coins": "41666667ubld",
      "length_seconds": 2678400
    },
    {
      "coins": "41666667ubld",
      "length_seconds": 2592000
    },
    {
      "coins": "41666666ubld",
      "length_seconds": 2678400
    },
    {
      "coins": "41666667ubld",
      "length_seconds": 2595600
    },
    {
      "coins": "41666667ubld",
      "length_seconds": 2678400
    }
  ]
}
```
