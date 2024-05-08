# Cosmos Hub Inflation variable modification: Blocks Per Year

**Quick Summary of issue**

There are 6 main variables that control the maximum, minimum, & change of the
atom inflation rate for the cosmos hub. Description of these variables can be
found under the [mint module in the parameters
wiki](https://github.com/cosmos/governance/blob/master/params-change/Mint.md).

In this proposal we will be looking at adjusting the blocks per year parameter.

Currently the variable named “blocks per year” is set at 4,855,015. This works
out to one block every 6.5 second roughly, which as many Atom holders know, is
not a very good approximation. This leads to the stated inflation rate of the
cosmos hub to not match reality.

**How to fix the issues**

The goal is to select a value that is as close as possible to the future block
throughput for the cosmos hub. To do that I will look at current (past couple
days) and historical time frames to try and get as close of an approximation as
possible.

**Past Blocks per Year Data**

Using Big Dipper, CosmosScan, or any one of the popular cosmos hub explorers,
the time stamp for each block can be found. The typical cosmos hub block comes
in between 7-8 seconds, with the majority being closer to 7. If you look over
the past day (written on 10/14/2020) you can see an average block time coming in
around 7.29 seconds. Looking on an hourly & minute level, 7.25-7.3 seconds per
block can be seen fairly consistently. Big Dipper has also conveniently provided
the all time (for cosmos hub-3) block time data, which is coming in around 7.18
second. Considering the slight discrepancy, I figured shooting right in the
middle would be an appropriate starting point, which could later be adjusted for
finer accuracy if need be. Now to find how many seconds are in a year, which
equals 365.25 (days / year) X 24 (Hours / Day) x 60 (Minutes / Hour) X 60
(Seconds / minute) = 31.5576 million seconds per year. A quick google search can
confirm the math. So finally, taking 31.5576 Million / 7.24 we get a value of
4.358 Million blocks per year, which can be rounded up to **4.36 Million blocks
per year**.

**Possible Risks / Benefits**

I will split this up into two sections, doing nothing & doing the proposed
changes.

1a) Doing nothing Risks / Benefits: There are no structural risks per say doing
nothing, but the stated inflation rate of the hub will continue to not match
reality. There are very little benefits of doing nothing; besides the fact its
working just fine now as long as you don’t care how close stated inflation is vs
real inflation.

1b) Changing to recommended value Risks / Benefits: Again, I don’t think there
are structural / game theory risks to making the blocks per year more closely
match reality. On the other hand, changing this variable to more closely match
reality is beneficial to all participants when doing any sort of economic
calculus. Currently the true inflation rate is actually lower than the stated
rate by a factor of 10ish % (4.36 Mil blocks per year / 4.85 Mil blocks per
year). So while the target rate is 7%, the actual current rate is more like
~6.29%.

**Conclusion**

I know there isn’t a right answer for blocks per year. I fully expect this value
to be fine tuned over the coming years / decades. This seems like a very good
starting place and a greatly beneficial change before we enter the post star
gate world ☺
