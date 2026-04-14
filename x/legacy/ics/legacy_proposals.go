package ics

// stubProposal satisfies both proto.Message and the govv1beta1.Content
// interface with stub/no-op implementations.
type stubProposal struct{ stubMsg }

func (s *stubProposal) GetTitle() string       { return "Legacy ICS Proposal" }
func (s *stubProposal) GetDescription() string { return "" }
func (s *stubProposal) ProposalRoute() string  { return "provider" }
func (s *stubProposal) ProposalType() string   { return "LegacyICS" }
func (s *stubProposal) ValidateBasic() error   { return nil }

// ICS provider governance proposal stubs.

type (
	ConsumerAdditionProposal     struct{ stubProposal }
	ConsumerRemovalProposal      struct{ stubProposal }
	ConsumerModificationProposal struct{ stubProposal }
	ChangeRewardDenomsProposal   struct{ stubProposal }
	EquivocationProposal         struct{ stubProposal }
)

func (m *ConsumerAdditionProposal) ProposalType() string     { return "ConsumerAddition" }
func (m *ConsumerRemovalProposal) ProposalType() string      { return "ConsumerRemoval" }
func (m *ConsumerModificationProposal) ProposalType() string { return "ConsumerModification" }
func (m *ChangeRewardDenomsProposal) ProposalType() string   { return "ChangeRewardDenoms" }
func (m *EquivocationProposal) ProposalType() string         { return "Equivocation" }

// Descriptor satisfies the descriptorIface required by the Cosmos SDK's
// unknownproto package for field validation when a historical governance
// proposal tx (submit-proposal) is retrieved by hash. Each proposal's fields
// are declared with wire-type-compatible descriptors matching the original ICS
// provider proto so that the unknownproto field checker does not reject stored
// historical data.

func (m *ConsumerAdditionProposal) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxConsumerAdditionProposal}
}

func (m *ConsumerRemovalProposal) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxConsumerRemovalProposal}
}

func (m *ConsumerModificationProposal) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxConsumerModificationProposal}
}

func (m *ChangeRewardDenomsProposal) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxChangeRewardDenomsProposal}
}

func (m *EquivocationProposal) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxEquivocationProposal}
}
