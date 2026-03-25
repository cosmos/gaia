package ics

// stubProposal satisfies both proto.Message and the govv1beta1.Content
// interface with stub/no-op implementations.
type stubProposal struct{ stubMsg }

func (s *stubProposal) GetTitle() string       { return "Legacy ICS Proposal" }
func (s *stubProposal) GetDescription() string { return "" }
func (s *stubProposal) ProposalRoute() string  { return "provider" }
func (s *stubProposal) ProposalType() string   { return "LegacyICS" }

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
