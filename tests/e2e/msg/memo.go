package msg

import (
	"fmt"
	"time"
)

func BuildCallbacksMemo(entrypointAddress, recipientDenom, adapterAddress, recipientAddress string) string {
	ibcHooksData := fmt.Sprintf(`"wasm": {
						"contract": "%s",
						"msg": {
						  "action": {
							"sent_asset": {
							  "native": {
								"denom":"%s",
								"amount":"1"
							  }
							},
							"exact_out": false,
							"timeout_timestamp": %d,
							"action": {
							  "transfer":{
								"to_address": "%s"
							  }
							}
						  }
						}
					  }`, entrypointAddress, recipientDenom, time.Now().Add(time.Minute).UnixNano(), recipientAddress)
	destCallbackData := fmt.Sprintf(`"dest_callback": {
					"address": "%s",
					"gas_limit": "%d"
				  }`, adapterAddress, 10_000_000)
	memo := fmt.Sprintf("{%s,%s}", destCallbackData, ibcHooksData)
	return memo
}
