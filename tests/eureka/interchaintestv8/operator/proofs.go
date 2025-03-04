package operator

// SupportedProofType is an enum for supported proof types.
type SupportedProofType int

const (
	// ProofTypeGroth16 represents the Groth16 SP1 proof type.
	ProofTypeGroth16 SupportedProofType = iota
	// ProofTypePlonk represents the Plonk SP1 proof type.
	ProofTypePlonk
)

// String returns the string representation of the proof type.
func (pt SupportedProofType) String() string {
	return [...]string{"groth16", "plonk"}[pt]
}

// ToOperatorArgs returns the proof type as arguments for the operator command.
func (pt SupportedProofType) ToOperatorArgs() []string {
	return []string{"-p", pt.String()}
}
