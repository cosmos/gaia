# Maintaining Cosmoshub Proto Files
All Cosmoshub proto files are defined here.

Updating the dependencies of third_party proto:
```bash
# update the deps hash in buf.yaml, then
cd proto
buf mod update
cd ..
make proto-gen
```
