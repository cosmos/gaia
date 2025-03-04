//! Programs for `sp1-ics07-tendermint`.

use sp1_sdk::{Prover, ProverClient, SP1VerifyingKey};

/// Trait for SP1 ICS07 Tendermint programs.
pub trait SP1Program {
    /// The ELF file for the program.
    const ELF: &'static [u8];

    /// Get the verifying key for the program using [`MockProver`].
    #[must_use]
    fn get_vkey() -> SP1VerifyingKey {
        let mock_prover = ProverClient::builder().mock().build();
        let (_, vkey) = mock_prover.setup(Self::ELF);
        vkey
    }
}

/// SP1 ICS07 Tendermint update client program.
pub struct UpdateClientProgram;

/// SP1 ICS07 Tendermint verify (non)membership program.
pub struct MembershipProgram;

/// SP1 ICS07 Tendermint update client and verify (non)membership program.
pub struct UpdateClientAndMembershipProgram;

/// SP1 ICS07 Tendermint misbehaviour program.
pub struct MisbehaviourProgram;

impl SP1Program for UpdateClientProgram {
    const ELF: &'static [u8] =
        include_bytes!("../../../target/elf-compilation/riscv32im-succinct-zkvm-elf/release/sp1-ics07-tendermint-update-client");
}

impl SP1Program for MembershipProgram {
    const ELF: &'static [u8] =
        include_bytes!("../../../target/elf-compilation/riscv32im-succinct-zkvm-elf/release/sp1-ics07-tendermint-membership");
}

impl SP1Program for UpdateClientAndMembershipProgram {
    const ELF: &'static [u8] =
        include_bytes!("../../../target/elf-compilation/riscv32im-succinct-zkvm-elf/release/sp1-ics07-tendermint-uc-and-membership");
}

impl SP1Program for MisbehaviourProgram {
    const ELF: &'static [u8] =
        include_bytes!("../../../target/elf-compilation/riscv32im-succinct-zkvm-elf/release/sp1-ics07-tendermint-misbehaviour");
}
