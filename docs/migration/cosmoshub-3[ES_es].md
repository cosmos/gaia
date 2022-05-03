# Instrucciones de actualización del Cosmos Hub 3

El siguiente documento describe los pasos necesarios que deben seguir los operadores de un full node para actualizar de `cosmoshub-3` a `cosmoshub-4`. El equipo de Tendermint publicará un archivo génesis oficial actualizado, pero se recomienda que los validadores ejecuten las siguientes instrucciones para verificar el archivo génesis resultante.

Existe un amplio consenso social en torno a la `propuesta de actualización del Cosmos Hub 4` sobre el `cosmoshub-3`. Siguiendo las propuestas #[27](https://www.mintscan.io/cosmos/proposals/27), #[35](https://www.mintscan.io/cosmos/proposals/35) y #[36](https://www.mintscan.io/cosmos/proposals/36). Se indica que el procedimiento de actualización debe realizarse el `18 de febrero de 2021 a las 06:00 UTC`.

- [Migraciones](#migraciones)
- [Preliminares](#preliminares)
- [Principales actualizaciones](#principales-actualizaciones)
- [Riesgos](#riesgos)
- [Recuperación](#recuperación)
- [Procedimiento de actualización](#procedimiento-de-actualización)
- [Notas para los proveedores de servicios](#notas-para-los-proveedores-de-servicios)

# Migraciones

Estos capítulos contienen todas las guías de migración para actualizar tu aplicación y módulos a Cosmos v0.40 Stargate.

Si tienes un explorador de bloques, un monedero, un exchange, un validador o cualquier otro servicio (por ejemplo, un proveedor de custodia) que dependa del Cosmos Hub o del ecosistema Cosmos, deberás prestar atención, porque esta actualización implicará cambios sustanciales.

1. [Migración de aplicaciones y módulos](ttps://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/app_and_modules.md)
1. [Guía de actualización de la cadena a v0.41](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/chain-upgrade-guide-040.md)
1. [Migración de endpoints REST](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/rest.md)
1. [Recopilación de modificaciones de ruptura de los registros de cambios](https://github.com/cosmos/gaia/blob/main/docs/migration/breaking_changes.md)
1. [Comunicación entre cadenas de bloques (IBC) - transacciones entre cadenas](https://figment.io/resources/cosmos-stargate-upgrade-overview/#ibc)
1. [Migración de Protobuf - rendimiento de la cadena de bloques y aceleración del desarrollo](https://figment.network/resources/cosmos-stargate-upgrade-overview/#proto)
1. [Sincronización de estados - minutos para sincronizar nuevos nodos](https://figment.network/resources/cosmos-stargate-upgrade-overview/#sync)
1. [Clientes ligeros con todas las funciones](https://figment.network/resources/cosmos-stargate-upgrade-overview/#light)
1. [Módulo de actualización de la cadena - automatización de la actualización](https://figment.network/resources/cosmos-stargate-upgrade-overview/#upgrade)

Si quieres probar el procedimiento antes de que se produzca la actualización el 18 de febrero, consulta este [post](https://github.com/cosmos/gaia/issues/569#issuecomment-767910963) en relación a ello.

## Preliminares

Se han producido muchos cambios en el SDK de Cosmos y en la aplicación Gaia desde la última gran actualización (`cosmoshub-3`). Estos cambios consisten principalmente en muchas nuevas características, cambios de protocolo y cambios estructurales de la aplicación que favorecen la ergonomía del desarrollador y el desarrollo de la aplicación.

En primer lugar, se habilitará [IBC](https://docs.cosmos.network/master/ibc/overview.html) siguiendo los [estándares de Interchain](https://github.com/cosmos/ics#ibc-quick-references). Esta actualización viene con varias mejoras en la eficiencia, la sincronización de nodos y las siguientes actualizaciones de la cadena de bloques. Más detalles en el [sitio web de Stargate](https://stargate.cosmos.network/).

__La aplicación [Gaia](https://github.com/cosmos/gaia) v4.0.0 es lo que los operadores de nodos actualizarán y ejecutarán en esta próxima gran actualización__. Tras la versión v0.41.0 del SDK de Cosmos y la v0.34.3 de Tendermint.

## Principales actualizaciones

Hay muchas características y cambios notables en la próxima versión del SDK. Muchos de ellos se discuten a de forma general [aquí](https://github.com/cosmos/stargate).

Algunos de los principales cambios que hay que tener en cuenta a la hora de actualizar como desarrollador o cliente son los siguientes:

- **Protocol Buffers**: Inicialmente el SDK de Cosmos utilizaba _codecs_ de Amino para casi toda la codificación y decodificación. En esta versión se ha integrado una importante actualización de los Protocol Buffers. Se espera que con los Protocol Buffers las aplicaciones ganen en velocidad, legibilidad, conveniencia e interoperabilidad con muchos lenguajes de programación. Para más información consulta [aquí](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/app_and_modules.md#protocol-buffers).
- **CLI**: El CLI y el commando de full node para la cadena de bloques estaban separados en las versiones anteriores del SDK de Cosmos. Esto dio lugar a dos binarios, `gaiad` y `gaiacli`, que estaban separados y podían utilizarse para diferentes interacciones con la cadena de bloques. Ambos se han fusionado en un solo comando `gaiad` que ahora soporta los comandos que antes soportaba el `gaiacli`.
- **Configuración del nodo**: Anteriormente los datos de la cadena de bloques y la configuración de los nodos se almacenaban en `~/.gaia/`, ahora residirán en `~/.gaia/`, si utilizas scripts que hacen uso de la configuración o de los datos de la cadena de bloques, asegúrate de actualizar la ruta.

## Riesgos

Como validador, realizar el procedimiento de actualización en sus nodos de consenso conlleva un mayor riesgo de de doble firma y de ser penalizado. La parte más importante de este procedimiento es verificar su versión del software y el hash del archivo génesis antes de iniciar el validador y firmar.

Lo más arriesgado que puede hacer un validador es descubrir que ha cometido un error y volver a repetir el procedimiento de actualización durante el arranque de la red. Si descubre un error en el proceso, lo mejor es esperar a que la red se inicie antes de corregirlo. Si la red se detiene y ha comenzado con un archivo de génesis diferente al esperado, busque el asesoramiento de un desarrollador de Tendermint antes de reiniciar su validador.

## Recuperación

Antes de exportar el estado de `cosmoshub-3`, se recomienda a los validadores que tomen una instantánea completa de los datos a la altura de la exportación antes de proceder. La toma de snapshots depende en gran medida de la infraestructura, pero en general se puede hacer una copia de seguridad del directorio `.gaia`.

Es muy importante hacer una copia de seguridad del archivo `.gaia/data/priv_validator_state.json` después de detener el proceso de gaiad. Este archivo se actualiza en cada bloque cuando tu validador participa en las rondas de consenso. Es un archivo crítico necesario para evitar la doble firma, en caso de que la actualización falle y sea necesario reiniciar la cadena anterior.

En el caso de que la actualización no tenga éxito, los validadores y operadores deben volver a actualizar a
gaia v2.0.15 con v0.37.15 del _Cosmos SDK_ y restaurar a su último snapshot antes de reiniciar sus nodos.

## Procedimiento de actualización

__Nota__: Se asume que actualmente está operando un nodo completo ejecutando gaia v2.0.15 con v0.37.15 del _Cosmos SDK_.

El hash de la versión/commit de Gaia v2.0.15: `89cf7e6fc166eaabf47ad2755c443d455feda02e`

1. Compruebe que está ejecutando la versión correcta (v2.0.15) de _gaiad_:

    ```bash
    $ gaiad version --long
    name: gaia
    server_name: gaiad
    client_name: gaiacli
    version: 2.0.15
    commit: 89cf7e6fc166eaabf47ad2755c443d455feda02e
    build_tags: netgo,ledger
    go: go version go1.15 darwin/amd64
   ```

1. Asegúrese de que su cadena se detiene en la fecha y hora correctas:
    18 de febrero de 2021 a las 06:00 UTC es en segundos UNIX: `1613628000`

    ```bash
    perl -i -pe 's/^halt-time =.*/halt-time = 1613628000/' ~/.gaia/config/app.toml
    ```
1. Después de que la cadena se haya detenido, haz una copia de seguridad de tu directorio `.gaia`.

    ```bash
    mv ~/.gaia ./gaiad_backup
    ```
    **NOTA**: Se recomienda a los validadores y operadores que tomen una instantánea completa de los datos a la altura de la exportación antes de proceder en caso de que la actualización no vaya según lo previsto o si no se pone en línea suficiente poder de voto en un tiempo determinado y acordado. En tal caso, la cadena volverá a funcionar con `cosmoshub-3`. Consulte [Recuperación](#recuperación) para saber cómo proceder.

1. Exportar el estado existente de `cosmoshub-3`:

    Antes de exportar el estado a través del siguiente comando, el binario `gaiad` debe estar detenido. Como validador, puedes ver la última altura del bloque creado en el `~/.gaia/config/data/priv_validator_state.json` -o que ahora reside en `gaiad_backup` cuando hiciste una copia de seguridad como en el último paso- y obtenerla con

    ```bash
    cat ~/.gaia/config/data/priv_validator_state.json | jq '.height'
    ```
   
   ```bash
   $ gaiad export --for-zero-height --height=<height> > cosmoshub_3_genesis_export.json
   ```
   _esto puede llevar un tiempo, puede esperar una hora para este paso_

1. Verifique el SHA256 del archivo génesis exportado (ordenado):

    Compara este valor con otros validadores / operadores de nodos completos de la red.
    En el futuro será importante que todas las partes puedan crear la misma exportación de archivos génesis.

   ```bash
   $ jq -S -c -M '' cosmoshub_3_genesis_export.json | shasum -a 256
   [SHA256_VALUE]  cosmoshub_3_genesis_export.json
   ```

1. En este punto, ya tiene un estado de génesis exportado válido. Todos los pasos posteriores requieren ahora v4.0.0 de [Gaia](https://github.com/cosmos/gaia).
Compruebe el hash de su génesis con otros compañeros (otros validadores) en las salas de chat.

   **NOTA**: Go [1.15+](https://golang.org/dl/) es necesario!

   ```bash
   $ git clone https://github.com/cosmos/gaia.git && cd gaia && git checkout v4.0.0; make install
   ```

1. Verifique que está ejecutando la versión correcta (v4.0.0) de _Gaia_:

    ```bash
    $ gaiad version --long
    name: gaia
    server_name: gaiad
    version: 4.0.0
    commit: 2bb04266266586468271c4ab322367acbf41188f
    build_tags: netgo,ledger
    go: go version go1.15 darwin/amd64
    build_deps:
    ...
    ```
    El hash y versión/commit de Gaia v4.0.0: `2bb04266266586468271c4ab322367acbf41188f`

1. Migrar el estado exportado de la versión actual v2.0.15 a la nueva versión v4.0.0:

    ```bash
    $ gaiad migrate cosmoshub_3_genesis_export.json --chain-id=cosmoshub-4 --initial-height [last_cosmoshub-3_block+1] > genesis.json
    ```

    Esto migrará nuestro estado exportado del archivo `genesis.json` requerido para iniciar el cosmoshub-4.

1. Verifique el SHA256 del JSON final del génesis:

    ```bash
    $ jq -S -c -M '' genesis.json | shasum -a 256
    [SHA256_VALUE]  genesis.json
    ```

    Compare este valor con otros validadores / operadores de nodos de la red. 
    Es importante que cada parte pueda reproducir el mismo archivo genesis.json de los pasos correspondientes.

1. Reinicio del estado:

    **NOTA**: Asegúrese de tener una copia de seguridad completa de su nodo antes de proceder con este paso.
   Consulte [Recuperación](#recuperación) para obtener detalles sobre cómo proceder.

    ```bash
    $ gaiad unsafe-reset-all
    ```

1. Mueve el nuevo `genesis.json` a tu directorio `.gaia/config/`.

    ```bash
    cp genesis.json ~/.gaia/config/
    ```

1. Inicie su blockchain

    ```bash
    gaiad start
    ```

    Las auditorías automatizadas del estado de génesis pueden durar entre 30 y 120 minutos utilizando el módulo de crisis. Esto se puede desactivar mediante 
    `gaiad start --x-crisis-skip-assert-invariants`.

## Notas para los proveedores de servicios

# Servidor REST

En caso de que hayas estado ejecutando el servidor REST con el comando `gaiacli rest-server` previamente, ejecutar este comando ya no será necesario. El servidor API está ahora en proceso con el demonio y puede ser activado/desactivado por la configuración de la API en su `.gaia/config/app.toml`:

```
[api]
# Enable define si la API del servidor debe estar habilitada.
enable = false
# Swagger define si la documentación swagger debe registrarse automáticamente.
swagger = false
```

El ajuste `swagger` se refiere a la activación/desactivación de la API de documentos swagger, es decir, el punto final de la API /swagger/.

# Configuración gRPC

Configuración gRPC en tu `.gaia/config/app.toml`

```yaml
[grpc]
# Enable define si el servidor gRPC debe estar habilitado.
enable = true
# Address define la dirección del servidor gRPC a la que se vincula.
address = "0.0.0.0:9090"
```

# State Sync

Configuración de State Sync en tu `.gaia/config/app.toml`

```yaml
# State sync o las instantáneas de sincronización de estado permiten que otros nodos 
# se unan rápidamente a la red sin reproducir los bloques históricos, descargando y 
# aplicando en su lugar una instantánea del estado de la aplicación a una altura determinada.
[state-sync]
# snapshot-interval especifica el intervalo de bloques en el que se toman instantáneas 
# de sincronización de estado local (0 para deshabilitar). Debe ser un múltiplo de 
# pruning-keep-every.
snapshot-interval = 0
# snapshot-keep-recent especifica el número de instantáneas recientes a conservar y servir 
# (0 para conservar todas).
snapshot-keep-recent = 2
```
