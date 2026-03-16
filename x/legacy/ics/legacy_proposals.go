package ics

// stubProposal satisfies both proto.Message and the govv1beta1.Content
// interface with stub/no-op implementations.
type stubProposal struct{ stubMsg }

func (s *stubProposal) GetTitle() string       { return "Legacy ICS Proposal" }
func (s *stubProposal) GetDescription() string { return "" }
func (s *stubProposal) ProposalRoute() string  { return "provider" }
func (s *stubProposal) ProposalType() string   { return "LegacyICS" }
func (s *stubProposal) ValidateBasic() error   { return nil }
func (s *stubProposal) String() string         { return "{}" }

// ICS provider governance proposal stubs.

type (
	ConsumerAdditionProposal     struct{ stubProposal }
	ConsumerRemovalProposal      struct{ stubProposal }
	ConsumerModificationProposal struct{ stubProposal }
	ChangeRewardDenomsProposal   struct{ stubProposal }
	EquivocationProposal         struct{ stubProposal }
)

func (m *ConsumerAdditionProposal) ProtoMessage()            {}
func (m *ConsumerAdditionProposal) Reset()                   {}
func (m *ConsumerAdditionProposal) String() string           { return "{}" }
func (m *ConsumerAdditionProposal) ProposalType() string     { return "ConsumerAddition" }
func (m *ConsumerRemovalProposal) ProtoMessage()             {}
func (m *ConsumerRemovalProposal) Reset()                    {}
func (m *ConsumerRemovalProposal) String() string            { return "{}" }
func (m *ConsumerRemovalProposal) ProposalType() string      { return "ConsumerRemoval" }
func (m *ConsumerModificationProposal) ProtoMessage()        {}
func (m *ConsumerModificationProposal) Reset()               {}
func (m *ConsumerModificationProposal) String() string       { return "{}" }
func (m *ConsumerModificationProposal) ProposalType() string { return "ConsumerModification" }
func (m *ChangeRewardDenomsProposal) ProtoMessage()          {}
func (m *ChangeRewardDenomsProposal) Reset()                 {}
func (m *ChangeRewardDenomsProposal) String() string         { return "{}" }
func (m *ChangeRewardDenomsProposal) ProposalType() string   { return "ChangeRewardDenoms" }
func (m *EquivocationProposal) ProtoMessage()                {}
func (m *EquivocationProposal) Reset()                       {}
func (m *EquivocationProposal) String() string               { return "{}" }
func (m *EquivocationProposal) ProposalType() string         { return "Equivocation" }

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
