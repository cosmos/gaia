# Gaia documentation archive

This branch contains all Gaia (CosmosHub) documentation files starting from v1.

## Adding to this branch

To add a documentation version to this archive you can copy the `docs` folder of the gaia repo and rename it to match the name of the version you copied:

```shell
# move to legacy-docs branch
git checkout legacy-docs
git checkout v15.2.0 -- docs # copy docs folder
mv docs v15 # create the archive
```

You can optionally remove docs platform files (docusaurus of vuepress) from the archive to leave only the `.md` files.

**Note**
This is an orphaned branch. It was created using:
```shell
git checkout --orphan legacy-docs
rm -rf <all directories> 
```

## Platform tooling

* Versions `v3` to `v14` use vuepress as build tool.
* Versions `>= v15` use docusaurus as build tool.
