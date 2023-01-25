# Pre Funders

The "seed" and "strategic/early" contributors were detailed shortly after the fundraiser,
at the bottom of a working copy of the recomended genesis file [available on
github](https://github.com/cosmos/fundraiser-lib/blob/32e01ca0a0d2c0fdf388e497a5dd0c4e8b1bf8a6/src/atom_query/data/fundraiser_atoms.json),
and was queryable at the [fundraiser website](https://fundraiser.cosmos.network).

Copy the bottom two sections of that [JSON file](https://github.com/cosmos/fundraiser-lib/blob/32e01ca0a0d2c0fdf388e497a5dd0c4e8b1bf8a6/src/atom_query/data/fundraiser_atoms.json),
into `early.json` and `seed.json`, respectively, and make them lists (to
detect duplicates).
One address, which is present in both files (`df403fa10845bd5e238827bd0d937a8c52f3a64d`),
is modified to `40d19f92685334c8ace23f9dc8c0c85977483e26`.

Run `go run main.go` to output the `contributors.json` by consolidating all the duplicate entries.


