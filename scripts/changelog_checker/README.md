# Changelog Checker

This utility checks if the contents of the changelog fit the desired formatting and spelling requirements.
Run the checker by executing `make check-changelog` in the root directory of the repository.

```bash
make check-changelog
```

It is also possible to have the script fix some easily fixable issues automatically.

```bash
make fix-changelog
```

## Configuration

It is possible to adjust the configuration of the changelog checker
by adjusting the contents of `config.py`.

Things that can be adjusted include:

- the allowed change types with a release
- the allowed description categories (i.e. the `(...)` portion at the beginning of an entry)
- PRs that are allowed to occur twice in the changelog (e.g. backports of bug fixes)
- a set of known patterns in PR descriptions and their preferred way of spelling
- known exceptions that do not need to follow the formatting rules
- the legacy version at which to stop the checking
