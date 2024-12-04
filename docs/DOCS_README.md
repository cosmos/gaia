<!--
parent:
  order: false
-->

# Updating the docs

If you want to open a PR on Gaia to update the documentation, please follow the guidelines in the [`CONTRIBUTING.md`](https://github.com/cosmos/gaia/tree/main/CONTRIBUTING.md)

## Docs Build Workflow

The documentation for Gaia is hosted at:

- <https://hub.cosmos.network/>

built from the files in this (`/docs`) directory for [main](https://github.com/cosmos/gaia/tree/main/docs).

### How It Works

There is a [Github Action](https://github.com/cosmos/gaia/blob/main/.github/workflows/deploy-docs.yml)
listening for changes in the `/docs` directory, on the `main` branch.
Any updates to files in this directory on that branch will automatically
trigger a website deployment. Under the hood, `make build-docs` is run from the
[Makefile](https://github.com/cosmos/gaia/blob/main/Makefile) in this repo.

## README

The [README.md](./README.md) is also the landing page for the documentation
on the website. During the Jenkins build, the current commit is added to the bottom
of the README.

## Links

**NOTE:** Strongly consider the existing links - both within this directory
and to the website docs - when moving or deleting files.

Relative links should be used nearly everywhere, having discovered and weighed the following:

### Relative

Where is the other file, relative to the current one?

- works both on GitHub and for the VuePress build
- confusing / annoying to have things like: `../../../../myfile.md`
- requires more updates when files are re-shuffled

### Absolute

Where is the other file, given the root of the repo?

- works on GitHub, doesn't work for the VuePress build
- this is much nicer: `/docs/hereitis/myfile.md`
- if you move that file around, the links inside it are preserved (but not to it, of course)

### Full

The full GitHub URL to a file or directory. Used occasionally when it makes sense
to send users to the GitHub.

## Building Locally

To build and serve the documentation locally, make sure you're in the `docs` directory and run the following:

Clear `node_modules` for a clean install. This is not necessary every time.

```bash
rm -rf node_modules
```

Install project dependencies

```bash
npm install
```

Pull the versions referenced in `versions.json` and `docusaurus.config.js`:

```shell
cd ../ # back to project root
make build-docs
git checkout <your_working_branch>
cd docs
```

Start the app

```bash
npm run start
```

then navigate to `localhost:3000` in your browser.

To build documentation as a static website run `npm run build`. You will find the website in `build` directory.
After building the docs website you can start it with `npm run serve`.


## Update and Build the RPC docs

1. Execute the following command at the root directory to install the swagger-ui generate tool.

   ```bash
   make tools
   ```

2. Edit API docs
   1. Directly Edit API docs manually: `cmd/gaiad/swagger-ui/swagger.yaml`.
   2. Edit API docs within the [Swagger Editor](https://editor.swagger.io/). Please refer to this [document](https://swagger.io/docs/specification/2-0/basic-structure/) for the correct structure in `.yaml`.
3. Download `swagger.yaml` and replace the old `swagger.yaml` under folder `cmd/gaiad/swagger-ui`.
4. Compile gaiad

   ```bash
   make install
   ```
