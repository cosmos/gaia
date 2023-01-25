# ICF

Run `go run main.go` to get a summary of the allocations from the ICF's 10%
and output the full set of addresses and amounts in `contributors.json`.

## Early

Addresses were submitted via email and collected in a spreadsheet.

Some addresses were submitted as hex, long before bech32 was used. 
They have been converted to their bech32 version.

## GoS

Addresses were collected through the GoS KYC process and collected in a
spreadsheet with all information and rewards. The address and amount columns were 
copied into a file, any empty lines were removed, and the lines were formatted
as csv.

## Multisig

Foundation atoms are split between two multisigs - a 2-of-3 with 30% of the ICF atoms
and a 3-of-5 with 70%. 
