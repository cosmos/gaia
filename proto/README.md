# Maintaining Cosmoshub Proto Files
All of the Cosmoshub proto files are defined here.

Updating the dependencies of third_party proto:
```bash
# update  version of ibc and tendermint when update deps in makefile, then
make proto-update-deps
# update the deps hash in buf.yaml, then
make proto-gen
```
