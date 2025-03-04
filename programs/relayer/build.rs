fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::configure()
        .build_server(true)
        .compile_protos(&["proto/relayer/relayer.proto"], &["proto/relayer"])?;
    Ok(())
}
