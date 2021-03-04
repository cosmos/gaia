#!/bin/bash
set -e

# Banner 'cause why not right?'
echo -e "
  _____          _____          
 / ____|   /\   |_   _|   /\    
| |  __   /  \    | |    /  \   
| | |_ | / /\ \   | |   / /\ \  
| |__| |/ ____ \ _| |_ / ____ \ 
 \_____/_/    \_\_____/_/    \_\
                                
"

DEPENDENCIES="curl"
CPUTYPE=$(uname -m)
if [ $CPUTYPE == "x86_64" ]; then
CPUTYPE="amd64"
fi
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH="$OS-$CPUTYPE"
SEEDS="bf8328b66dceb4987e5cd94430af66045e59899f@public-seed.cosmos.vitwit.com:26656,cfd785a4224c7940e9a10f6c1ab24c343e923bec@164.68.107.188:26656,d72b3011ed46d783e369fdf8ae2055b99a1e5074@173.249.50.25:26656,ba3bacc714817218562f743178228f23678b2873@public-seed-node.cosmoshub.certus.one:26656,3c7cad4154967a294b3ba1cc752e40e8779640ad@84.201.128.115:26656"


function getMatchingAssets {
    # the following sed/grep block of magic transforms the api response from GitHub into a list of
    # the assets of the latest release, the list consists of the asset name and then its url,
    # all separated by newlines; this got make `jq` obsolete as dependency
    assets=$(curl -s https://api.github.com/repos/cosmos/gaia/releases/latest | \
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
            if [ "$line" == "gaiad-$ARCH" ]; then
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
    echo "Could not find a release of Gaia for your architecture ($ARCH)."
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
- ~/.gaiad"

read -p "Are you ok with that? (y/N): " choice
case "$choice" in
    y|Y) echo -e "Continuing with install... This could take a moment.\n";;
    *) echo "Aborting."; exit 1;;
esac


# create install dir (if necessary) and change into it
[[ ! -d "$INSTALL_DIR" ]] && mkdir "$INSTALL_DIR"
cd "$INSTALL_DIR"


# clear old files
echo "Removing previously installed gaiad."
[[ -f "./gaiad" ]] && rm "./gaiad"
[[ -d "$HOME/.gaiad" ]] && rm -r "$HOME/.gaiad"


# download the (previously) matched release assets
echo -e "\nDownloading and installing gaiad."
downloadAssets $urls

# move the binaries to not include the arch and make them executable
mv "gaiad_$ARCH" "gaiad"
chmod +x "gaiad"


# intialize gaiad
echo -e "\nInitializing gaiad...."
./gaiad init &>/dev/null


# add seeds
echo -e "\nAdding seeds to config...."
original_string="seeds = \"\""
replace_string="seeds = \"$SEEDS\""
sed -i -e "s/$original_string/$replace_string/g" "$HOME/.gaiad/config/config.toml"

# get moniker
echo -e "\nNow you can give your name a moniker, human-readable name that lets people identify your node.  It does not need to be your human name.\n"
read -p "Name your gaia node: " name
moniker_original="moniker = \"\""
moniker_actual="moniker = \"$name\""
sed -i -e "s/$moniker_original/$moniker_actual/g" "$HOME/.gaiad/config/config.toml"

# fetch genesis.json
echo -e "\nFetching genesis...."
curl -Os "https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz"
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json ~/.gaia/config/genesis.json
mv "genesis.json" "$HOME/.gaiad/config/genesis.json"


# summary
echo -e "\033[1;35m\n\nWelcome to the Cosmos Hub! \xF0\x9F\x8E\x89 \xF0\x9F\x8C\x8C ..................... \033[0m\n
Gaia now installed and ready to sync......
...............................................................................
Navigate into the Gaia directory by typing the following;
cd ~/gaia
..............................................................................."

echo -e "\nThen open a new terminal window and sync your Gaia Node by typing....
\"$INSTALL_DIR\"
./gaiad start
Note: Syncing your Gaia Node can take a while."


