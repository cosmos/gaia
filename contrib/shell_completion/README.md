# gaiad

This [ohmyzsh](https://github.com/ohmyzsh/ohmyzsh) plugin adds completion for the cosmos hub blockchain [`Cosmos Hub`](https://github.com/cosmos/gaia), also known as `gaia`.

To use it, copy the contents of the folder into `$ZSH/custom/gaia` and add `gaiad` to the plugins array in your zshrc file:

```zsh
plugins=(... gaiad)
```

To regenerate the script please run `make gen-completion`
