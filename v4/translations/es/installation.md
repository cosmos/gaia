<!--
order: 2
-->

# Instalación de Gaia

Esta guía le explicará como instalar los puntos de entrada `gaiad` y `gaiad` en su sistema. Con esto instalado en su servidor, puede participar en la red principal como un [Full Node](./join-mainnet.md) o como un [Validador](../validators/validator-setup.md).

## Instalación de Go

Instale `Go` siguiendo la [documentación oficial](https://golang.org/doc/install).
Recuerde establecer su variable de entorno en el `$PATH` por ejemplo:

```bash
mkdir -p $HOME/go/bin
echo "export PATH=$PATH:$(go env GOPATH)/bin" >> ~/.bash_profile
source ~/.bash_profile
```

:::consejo
**Go 1.14+** es necesario para el SDK de Cosmos.
:::

## Instalación de los binarios

Siguiente, instalemos la última versión de Gaia. Asegúrese de hacer `git checkout` a la [versión publicada](https://github.com/cosmos/gaia/releases) correcta.

```bash
git clone -b <latest-release-tag> https://github.com/cosmos/gaia
cd gaia && make install
```

Si este comando falla a causa del siguiente mensaje de error, es posible que ya haya establecido `LDFLAGS` antes de ejecutar este paso.

```
# github.com/cosmos/gaia/cmd/gaiad
flag provided but not defined: -L
usage: link [options] main.o
...
make: *** [install] Error 2
```

Elimine esta variable de entorno e inténtelo de nuevo.

```
LDFLAGS="" make install
```

> _NOTA_: Si aún tiene problemas en este paso, por favor compruebe que tiene instalada la última versión estable de GO.

Esto debería instalar los binarios de `gaiad`y `gaiad`. Verifique que todo esta OK:

```bash
$ gaiad version --long
$ gaiad version --long
```

`gaiad` por su parte, debería dar como resultado algo similar a:

```shell
name: gaia
server_name: gaiad
client_name: gaiad
version: 2.0.3
commit: 2f6783e298f25ff4e12cb84549777053ab88749a
build_tags: netgo,ledger
go: go version go1.12.5 darwin/amd64
```

### Tags para la construcción

Las etiquetas (_tags_) para la construcción indican opciones especiales que deben ser activadas en el binario.

| Etiquetas Construcción | Descripción                                     |
| --------- | ----------------------------------------------- |
| netgo     | La resolución del nombre usará código puro de Go |
| ledger    | Añade compatibilidad de dispositivos hardware (wallets físicas) |

### Instalación de los binarios via snap (Linux solamente)

**No use _snap_ en este momento para instalar los binarios para producción hasta que tengamos un sistema binario reproducible.**

## Workflow para el desarrollador

Para probar cualquier cambio hecho en el SDK o Tendermint, se debe agregar una cláusula de `replace` en `go.mod` proporcionando la ruta de entrada correcta.

- Realice los cambios apropiados
- Añada `replace github.com/cosmos/cosmos-sdk => /ruta/a/clon/cosmos-sdk` en `go.mod`
- Ejecute `make clean install` o `make clean build`
- Compruebe sus cambios

## Siguiente

Ahora puede unirse a la [red principal](./join-mainnet.md), [testnet](./join-testnet.md) o crear [su propia testnet pública](./deploy-testnet.md)
