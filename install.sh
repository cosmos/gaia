#!/bin/bash
set -e

# Galaxy ACII graphic with newline for better UI
echo -e "
                                                       +
                                     +
                      +                       +
                              +
                                 +        +
                  +
                          +     +                +
                       +          +
    Galaxy                  +
                +                       +
                      +
        +                +         +
                              +
                   +
"

DEPENDENCIES="curl"

ARCH="$(uname -s)_$(uname -m)"
INSTALL_DIR="$HOME/galaxy"
SEEDS=$(curl -s https://raw.githubusercontent.com/galaxypi/galaxy/develop/seeds)


function getMatchingAssets {
    # the following sed/grep block of magic transforms the api response from GitHub into a list of
    # the assets of the latest release, the list consists of the asset name and then its url,
    # all separated by newlines; this got make `jq` obsolete as dependency
    assets=$(curl -s https://api.github.com/repos/galaxypi/galaxy/releases/latest | \
      sed -n '/"assets": \[/,/\]/p' | \
      sed -n '/"\(name\|browser_download_url\)": "/p' | \
      sed 's/\s*"name": "//g' | \
      sed 's/^\s*"browser_download_url": "//g' | \
      sed 's/",$//g' | \
      sed 's/"$//g')
    # meta: the following comments document the commands from above line by line (because of the `\` it's not possible to document them above)
    # downloads JSON for the last release
    # greps the relevant asset block
    # greps all relevant lines (asset names and asset urls)
    # removes the `"name": "` at the beginning of a line
    # removes the `"browser_download_url": "` at the beginning of a line
    # removes the `",` at the end of a line
    # removes the `"` at the end of a line


    urls=""
    index=0
    take_next=false

    # go through the list
    while read -r line; do
        if [ $(( $index % 2 )) -eq 0 ]; then
            # `line` is a asset name

            # set `take_next` to true to add it to `urls` in the next iteration
            if [ "$line" == "galaxycli_$ARCH" ] || [ "$line" == "galaxyd_$ARCH" ]; then
                take_next=true
            fi
        else
            # `line` is a asset url

            # add the asset url to `urls` and reset `take_next`
            if [ "$take_next" = true ]; then
                urls+="$line "
            fi
            take_next=false
        fi

        let index=index+1
    done <<< $assets

    echo "$urls"
}


function downloadAssets {
    urls="$@"
    echo -e "\e[92m "
    for url in $urls; do
        curl -LO#f "$url"
    done
    echo -e "\e[0m "
}

# download and parse latest release information from GitHub
urls="$(getMatchingAssets)"

# fail if there wasn't a matching architecture in the release assets
if [ -z "$urls" ]; then
    echo "Could not find a matching release of galaxycli and galaxyd for your architecture ($ARCH)."
    echo "If you know what you're doing and think it should work on your architecture, you can set your architecture manually at the beginning of this script and then run it again."
    exit 1
fi


# check for needed dependencies
for dependency in $DEPENDENCIES; do
    if ! command -v "$dependency" &>/dev/null; then
        echo "It seems that \"$dependency\" isn't installed but I really need it :/"
        echo "Please install it and re-run this script."
        exit 1
    fi
done


# ask for consent
# Ask for consent with line breaks for better readability
echo -e "This script will remove previously installed directories:
- galaxycli
- galaxyd"

read -p "Are you ok with that? (y/N): " choice
case "$choice" in
    y|Y) echo -e "Continuing with install... This could take a moment.\n";;
    *) echo "Aborting."; exit 1;;
esac


# create install dir (if necessary) and change into it
[[ ! -d "$INSTALL_DIR" ]] && mkdir "$INSTALL_DIR"
cd "$INSTALL_DIR"


# clear old files
echo "Removing previously installed galaxycli and galaxyd."
[[ -f "./galaxycli" ]] && rm "./galaxycli"
[[ -f "./galaxyd" ]] && rm "./galaxyd"
[[ -d "$HOME/.galaxycli" ]] && rm -r "$HOME/.galaxycli"
[[ -d "$HOME/.galaxyd" ]] && rm -r "$HOME/.galaxyd"


# download the (previously) matched release assets
echo -e "\nDownloading and installing..... galaxycli and galaxyd."
downloadAssets $urls

# move the binaries to not include the arch and make them executable
mv "galaxycli_$ARCH" "galaxycli"
mv "galaxyd_$ARCH" "galaxyd"
chmod +x "galaxycli"
chmod +x "galaxyd"


# intialize galaxyd
echo -e "\nInitializing galaxyd...."
./galaxyd init &>/dev/null


# add seeds
echo -e "\nAdding seeds to config...."
original_string="seeds = \"\""
replace_string="seeds = \"$SEEDS\""
sed -i -e "s/$original_string/$replace_string/g" "$HOME/.galaxyd/config/config.toml"

# get moniker
echo -e "\nGalaxy needs to distinguish individual nodes from one another. This is \naccomplished by having users choose a Galaxy node name. \n\nRecommended name: 'galaxy-node'\n"
read -p "Name your galaxy node: " name
moniker_original="moniker = \"\""
moniker_actual="moniker = \"$name\""
sed -i -e "s/$moniker_original/$moniker_actual/g" "$HOME/.galaxyd/config/config.toml"

# fetch the genesis block
echo -e "\nFetching genesis block...."
curl -Os "https://raw.githubusercontent.com/galaxypi/galaxy/master/genesis.json"
mv "genesis.json" "$HOME/.galaxyd/config/genesis.json"


# summary
echo -e "\033[1;35m\n\nWelcome to the Galaxy network \xF0\x9F\x8E\x89 \xF0\x9F\x8C\x8C ..................... \033[0m\n
Galaxy blockchain is now installed and ready to sync......

...............................................................................

Navigate into the galaxy directory by typing the following;
cd ~/galaxy

..............................................................................."

echo -e "\nThen open a new terminal window and sync your Galaxy Node by typing....
\"$INSTALL_DIR\"
./galaxyd start

Note: Syncing your Galaxy Node can take a while."
