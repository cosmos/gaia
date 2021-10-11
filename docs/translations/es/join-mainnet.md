<!--
order: 3
-->

# Únase a la red principal del Cosmos Hub

::: tip
Vea el [repositorio para el lanzamiento](https://github.com/cosmos/launch) para la información de la red principal, incluyendo la versión correcta para el SDK de Cosmos que usar y detalles acerca del archivo génesis.
:::

::: aviso
**Necesitará [instalar gaia](./installation.md) antes de avanzar más**
:::

## Configurando un nuevo nodo

Estas instrucciones son para establecer un nuevo nodo completo desde cero.

Primero, inicie el nodo y cree los archivos de configuración necesarios:

```bash
gaiad init <your_custom_moniker>
```

:::Warning
El moniker solo debe contener carácteres ASCII.  El uso de caracteres Unicode hará que tu nodo sea irreconocible.
:::

Puede editar el apodo (`moniker`) después, en el archivo `~/.gaia/config/config.toml`:

```toml
# A custom human readable name for this node
moniker = "<tu nombre personalizado>"
```
Puede editar el archivo `~/.gaia/config/app.toml` para activar el mecanismo antispam y rechazar las transacciones entrantes con valores inferiores a los precios mínimos para el _gas_:

```
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 10uatom).

minimum-gas-prices = ""
```

¡Su nodo completo ha sido iniciado!

## Génesis y semillas

### Copie el archivo génesis

Busque el archivo `genesis.json` de la red principal en el directorio de configuración de `gaiad`.

```bash
mkdir -p $HOME/.gaia/config
curl https://raw.githubusercontent.com/cosmos/launch/master/genesis.json > $HOME/.gaia/config/genesis.json
```

Observe que usamos el directorio `latest` en el [repositorio de lanzamiento](https://github.com/cosmos/launch) que contiene detalles para la red principal como la última versión y el archivo de génesis.

:::consejo
Si en cambio quiere conectarse a la red de pruebas pública, haga clic [aquí](./join-testnet.md)
:::

Para verificar la validez de la configuración:

```bash
gaiad start
```

### Añada los nodos semilla

Su nodo necesita saber cómo encontrar pares (_peers_). Necesita añadir nodos semilla en buen estado en `$HOME/.gaia/config/config.toml`. El repositorio para el [`lanzamiento`](https://github.com/cosmos/launch) contiene enlaces a algunos nodos semilla.

Si estas semillas no funcionan, puedes encontrar más _seeds_ y _peers_ persistentes en un explorador de Cosmos Hub (puede encontrar una lista en la [página del lanzamiento](https://cosmos.network/launch))

También puedes preguntar por _peers_ en el [canal de Validadores de Riot](https://riot.im/app/#/room/#cosmos-validators:matrix.org)

Para más información acerca de seeds y peers, puede leer [este enlace](https://docs.tendermint.com/master/spec/p2p/peer.html)

## Nota sobre el Fee y el Gas

::: Aviso
En el Hub de Cosmos, la denominación aceptada es `uatom`, donde `1atom = 1.000.000uatom`
:::

Las transacciones en la red del Hub de Cosmos deben incluir una tarifa de transacción para poder ser procesadas. Esta tarifa paga el gas necesario para llevar a cabo la transacción. La fórmula es la siguiente:

```
tarifa = techo(gas * precioPorGas)
```

El `gas` depende de la transacción. Diferentes transacciones requieren diferentes cantidades de `gas`. La cantidad de `gas` para una transacción se calcula mientras se procesa, pero hay una forma previa de estimarla usando el valor `auto` para el indicador de `gas`. Por supuesto, esto sólo da una estimación. Puede ajustar esta estimación con el identificador `--gas-adjustment` (por defecto `1.0`) si quiere estar seguro de que proporciona suficiente `gas` para la transacción. 

El `gasPrice` (i.e `precioPorGas`) es el precio de cada unidad de `gas`. Cada validador establece un valor de `min-gas-price`, y sólo incluirá transacciones que tengan un `gasPrice` mayor que su `min-gas-price`.

Los `fees` de la transacción son el producto del `gas` y del `gasPrice`. Como usuario, tiene que introducir 2 de 3. Cuanto más alto sea el `gasPrice`/`fees`, mayor será la posibilidad de que su transacción se incluya en un bloque.

::: consejo
Para la red principal, el `gas-prices` recomendado es `0.0025uatom`.
:::

## Establezca `minimum-gas-prices`

Su nodo completo mantiene las transacciones no confirmadas en la _mempool_. Para protegerlo de ataques de spam, es mejor establecer un `minimum-gas-prices` que la transacción debe cumplir para ser aceptada en la _mempool_ de su nodo. Este parámetro puede ser establecido en el siguiente archivo `~/.gaia/config/app.toml`.

El valor inicial recomendado para `min-gas-prices` es `0.0025uatom`, pero puede querer cambiarlo más tarde.

## Reducción del Estado

Hay tres estrategias para reducir el estado, por favor tenga en cuenta que esto es sólo para el estado y no para el almacenamiento de bloques:

1. `PruneEverything`: Esto significa que todos los estados salvados serán reducidos aparte del actual.
2. `PruneNothing`: Esto significa que todo el estado se guardará y nada se borrará.
3. `PruneSyncable`: Esto significa que sólo se salvará el estado de los últimos 100 y cada 10.000 bloques.

Por defecto cada nodo está en modo `PruneSyncable`. Si desea cambiar su estrategia de reducción en su nodo, debe hacerlo cuando el nodo se ha iniciado. Por ejemplo, si desea cambiar su nodo al modo `PruneEverything` entonces puede pasar la opción `---pruning everything` cuando llame a `gaiad start`.

> Nota: Cuando esté en estado de reducción no podrá consultar las partes que no estén en su base de datos.

## Ejecute un nodo completo

Inicie el nodo completo con este comando:

```bash
gaiad start
```

Comprueba que todo funciona bien:

```bash
gaiad status
```

Vea el estado de la red con el [Explorador de Cosmos](https://cosmos.network/launch)

## Exportar el estado

Gaia puede volcar todo el estado de la aplicación a un archivo JSON, que podría ser útil para el análisis manual y también puede ser usado como el archivo génesis para una nueva red.

Exporte el estado con:

```bash
gaiad export > [filename].json
```

También puede exportar el estado desde una altura en especial (al final del procesamiento del bloque en esa altura):

```bash
gaiad export --height [height] > [filename].json
```

Si desea empezar una nueva red desde el estado exportado, expórtelo con la opción `--for-zero-height`:

```bash
gaiad export --height [height] --for-zero-height > [filename].json
```

## Verifica la red principal

Ayude a prevenir problemas críticos ejecutando invariantes en cada bloque de su nodo. En esencia, al ejecutar invariantes se asegura que el estado de la red principal es el estado esperado correcto. Una comprobación de la invariante vital es que ningún átomo está siendo creado o destruido fuera del protocolo esperado, sin embargo hay muchas otras invariantes, comprueben cada una de ellas de forma única para su respectivo módulo. Porque la invariante es costosa desde el punto de vista computacional, no están habilitados por defecto. Para ejecutar un nodo con  estas comprobaciones inicie su nodo con la opción assert-invariants-blockly:

```bash
gaiad start --assert-invariants-blockly
```

Si se rompe una invariante en su nodo, su nodo entrará en pánico (`panic` de Golang) y le pedirá que envíe una transacción que detenga la red principal. Por ejemplo, el mensaje proporcionado puede parecerse a:

```bash
invariant broken:
    loose token invariance:
        pool.NotBondedTokens: 100
        sum of account tokens: 101
    CRITICAL please submit the following transaction:
        gaiad tx crisis invariant-broken staking supply

```

Cuando se presenta una transacción inválida, no se deducen los tokens de honorarios de la transacción ya que la cadena de bloques se detendrá (también conocido como transacción gratuita).

## Actualice a un nodo validador

Ahora tienes un nodo completo activo. ¿Cuál es el siguiente paso? Puedes actualizar tu nodo completo para convertirte en un Validador del Cosmos. Los 120 mejores validadores tienen la capacidad de proponer nuevos bloques en el Hub de Cosmos. Continúe en la [Configuración del Validador](../validators/validator-setup.md)
