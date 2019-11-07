#!/bin/bash

from=1QJUkhLJqTgWnpgGhh64dYgPiig7QgwxfK
to=1Ap3rbMR8DCW56BvRFXp1XAana7G8wWN7F
miner=1Cci258kuNs9r9A7GzpP44RHCPFuoMgnRE

new=19dpiTubN8ty2Ji5JTrTSpbUYMhnuuq888
new2=1M5FCg1U3Wf1m3Fmedpf2W3R4Y6nVsqbpP


./blockchain send  $from $to 10 $miner "hello world"

echo "++++++++++++++"

./blockchain getBalance $from ##12.5 - 10 + 12.5 = 15
./blockchain getBalance $to #10

echo "**********************"
./blockchain send $from $new 14 $new2 "hello world"
./blockchain getBalance $from #1
./blockchain getBalance $new #1

