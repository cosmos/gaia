package ethereum

import "strconv"

func (h BeaconBlockHeader) GetSlot() uint64 {
	slot, err := strconv.ParseUint(h.Slot, 0, 0)
	if err != nil {
		panic(err)
	}

	return slot
}

func (u LightClientUpdate) GetSignatureSlot() uint64 {
	slot, err := strconv.ParseUint(u.SignatureSlot, 0, 0)
	if err != nil {
		panic(err)
	}

	return slot
}
