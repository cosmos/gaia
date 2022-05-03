# GitHub Actions

| Action                       | File            | Schedule        | Branch         | Tag       | Release | Env              | Description                                |
| ---------------------------- | --------------- | --------------- | -------------- | --------- | ------- | ---------------- | ------------------------------------------ |
| Docker - ghcr Build and Push | docker-push.yml | "0 10 \* \* \*" | "release/\*\*" | "v*.*.\*" |         | interchain_comms | Builds and publishes docker images to ghcr |
