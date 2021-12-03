# Módulo `gov`
El módulo `gov` es responsable de las propuestas de gobierno en cadena y la funcionalidad de la votación. Nótese que [este módulo requiere una forma única de cambiar sus parámetros](https://github.com/cosmos/cosmos-sdk/issues/5800). `gov` está activo en Cosmos Hub 3 y actualmente tiene tres parámetros con seis subkeys que pueden ser modificados por una propuesta de gobernanza:
1. [`depositparams`](#1-depositparams)
   - [`mindeposit`](#mindeposit) - `512000000` `uatom` (micro-ATOMs)
   - [`maxdepositperiod`](#maxdepositperiod) - `1209600000000000` (nanosegundos)

2. [`votingparams`](#2-votingparams)
   - [`votingperiod`](#votingperiod) - `1209600000000000` (nanosegundos)

3. [`tallyparams`](#3-tallyparams)
   - [`quorum`](#quorum) - `0.400000000000000000` (proporción de la red)
   - [`threshold`](#threshold) - `0.500000000000000000` (proporción del poder de voto)
   - [`veto`](#veto) - `0.334000000000000000` (proporción del poder de voto)

Los valores de lanzamiento de cada subkey de los parámetros están indicados arriba, pero puede [verificarlos usted mismo](./Governance.md#verify-parameter-values).

Se están considerando [algunas funciones adicionales](./Governance.md#future) para el desarrollo del módulo de gobernanza.

Si estás técnicamente preparado, [estas son las especificaciones técnicas](./Governance.md#technical-specifications). Si quieres crear una propuesta para cambiar uno o más de estos parámetros, [mira esta sección para el formato](../submitting.md#formatting-the-json-file-for-the-governance-proposal).

## 1. `depositparams`
## `mindeposit`
### El depósito mínimo requerido para que una propuesta entre en el [período de votación](params-change/Governance.md#votingperiod), en micro-ATOMs
#### `cosmoshub-3` por defecto: `512000000` `uatom`

Antes de que una propuesta de gobierno entre en el [período de votación](./Governance.md#votingperiod) (es decir, para que la propuesta sea votada), debe haber al menos un número mínimo de ATOMs depositados. Cualquiera puede contribuir a este depósito. Los depósitos de las propuestas aprobadas y fallidas se devuelven a los contribuyentes. Los depósitos se queman cuando las propuestas 1) [expiran](./Governance.md#maxdepositperiod), 2) no alcanzan el [quórum](./Governance.md#quorum), o 3) son [vetadas](./Governance.md#veto). El valor de subkey de este parámetro representa el depósito mínimo requerido para que una propuesta entre en el [período de votación](./Governance.md#votingperiod) en micro-ATOMs, donde `512000000uatom` equivalen a 512 ATOM.

### Posibles consecuencias
#### Disminución del valor `mindeposit`
La disminución del valor de subkey `mindeposit` permitirá que las propuestas de gobernanza entren en el [período de votación](./Governance.md#votingperiod) con menos ATOMs en juego. Es probable que esto aumente el volumen de nuevas propuestas de gobernanza.

#### Aumentar el valor `mindeposit`
Para aumentar el valor de subkey `mindeposit` será necesario arriesgar un mayor número de ATOMs antes de que las propuestas de gobierno puedan entrar en el [período de votación](./Governance.md#votingperiod). Es probable que esto disminuya el volumen de nuevas propuestas de gobierno.

## `maxdepositperiod`
### La cantidad máxima de tiempo que una propuesta puede aceptar contribuciones de depósito antes de expirar, en nanosegundos.
#### `cosmoshub-3` por defecto: `1209600000000000`

Antes de que una propuesta de gobierno entre en el [período de votación](./Governance.md#votingperiod), debe haber al menos un número mínimo de ATOMs depositados. El valor de subkey de este parámetro representa la cantidad máxima de tiempo que la propuesta tiene para alcanzar la cantidad mínima de depósito antes de expirar. La cantidad máxima de tiempo que una propuesta puede aceptar contribuciones de depósito antes de expirar es actualmente de 1209600000000000 nanosegundos o 14 días. Si la propuesta expira, cualquier cantidad de depósito será quemada.

### Posibles consecuencias
#### Disminución del valor `maxdepositperiod`
La disminución del valor de subkey `maxdepositperiod` reducirá el tiempo de depósito de las contribuciones a las propuestas de gobernanza. Es probable que esto disminuya el tiempo que algunas propuestas permanecen visibles y que disminuya la probabilidad de que entren en el período de votación. Esto puede aumentar la probabilidad de que las propuestas caduquen y se quemen sus depósitos.

#### Aumentar el valor `maxdepositperiod`
El aumento del valor de subkey `maxdepositperiod` ampliará el plazo para las contribuciones de depósito a las propuestas de gobernanza. Es probable que esto aumente el tiempo en que algunas propuestas siguen siendo visibles y aumente potencialmente la probabilidad de que entren en el [período de votación](./Governance.md#votingperiod). Esto puede disminuir la probabilidad de que las propuestas caduquen y se quemen sus depósitos.

#### Observaciones
Actualmente, la mayoría de los exploradores de la red (por ejemplo, Hubble, Big Dipper, Mintscan) dan la misma visibilidad a las propuestas en el período de depósito que a las del [período de votación](./Governance.md#votingperiod). Esto significa que una propuesta con un pequeño depósito (por ejemplo, 0.001 ATOM) tendrá la misma visibilidad que aquellas con un depósito completo de 512 ATOM en el período de votación.

## 2. `votingparams`
## `votingperiod`
### La cantidad máxima de tiempo que una propuesta puede aceptar votos antes de que concluya el período de votación, en nanosegundos.
#### `cosmoshub-3` por defecto: `1209600000000000`

Una vez que una propuesta de gobierno entra en el período de votación, hay un período máximo de tiempo que puede transcurrir antes de que concluya el período de votación. El valor de subkey de este parámetro representa la cantidad máxima de tiempo que la propuesta tiene para aceptar los votos, que actualmente es de `1209600000000000` nanosegundos o 14 días. Si la votación de la propuesta no alcanza el quórum (es decir, el 40% del poder de voto de la red participa) antes de este tiempo, se quemarán las cantidades depositadas y el resultado de la propuesta no se considerará válido. Los votantes pueden cambiar su voto tantas veces como quieran antes de que termine el período de votación. Este período de votación es actualmente el mismo para cualquier tipo de propuesta de gobierno.

### Posibles consecuencias
#### Disminución del valor `votingperiod`
La disminución del valor de subkey `votingperiod` reducirá el tiempo de votación de las propuestas de gobernanza. Esto podría significar:
1. disminuir la proporción de la red que participa en la votación, y
2. disminución de la probabilidad de que se alcance el quórum.

#### Aumentar el valor `votingperiod`
El aumento del valor de subkey `votingperiod` aumentará el tiempo de votación de las propuestas de gobernanza. Esto puede:
1. aumentar la proporción de la red que participa en la votación, y
2. aumentar la probabilidad de que se alcance el quórum.

#### Observaciones
Históricamente, los debates y el compromiso fuera de la cadena parecen haber sido mayores durante el período de votación de una propuesta de gobernanza que cuando la propuesta se publica fuera de la cadena como un boceto. En la segunda semana del período de votación se ha votado una cantidad no trivial del poder de voto. Las propuestas 23, 19 y 13 tuvieron cada una aproximadamente un 80% de participación en la red o más.

## 2. `tallyparams`
## `quorum`
### La proporción mínima de poder de voto de la red que se requiere para que el resultado de una propuesta de gobierno se considere válido.
#### `cosmoshub-3` por defecto: `0.400000000000000000`

Se requiere quórum para que el resultado de la votación de una propuesta de gobierno se considere válido y para que los contribuyentes de depósitos recuperen sus cantidades depositadas, y el valor de subkey de este parámetro representa el valor mínimo para el quórum. El poder de voto, ya sea que respalde un voto de 'yes', 'abstain', 'no', or 'no-with-veto', cuenta para el quórum. Si la votación de la propuesta no alcanza el quórum (es decir, el 40% del poder de voto de la red participa) antes de este momento, se quemará cualquier cantidad depositada y el resultado de la propuesta no se considerará válido.

### Posibles consecuencias
#### Disminución del valor `quorum`
La disminución del valor de subkey `quorum` permitirá que una proporción menor de la red legitime el resultado de una propuesta. Esto aumenta el riesgo de que se tome una decisión con una proporción menor de los participantes con ATOMs, al tiempo que disminuye el riesgo de que una propuesta se considere inválida. Esto probablemente disminuirá el riesgo de que el depósito de una propuesta se queme.

#### Aumentar el valor `quorum`
El aumento del valor de subkey `quorum` requerirá una mayor proporción de la red para legitimar el resultado de una propuesta. Esto disminuye el riesgo de que se tome una decisión con una proporción menor de los participantes con ATOMs, al tiempo que aumenta el riesgo de que una propuesta se considere inválida. Es probable que esto aumente el riesgo de que se queme el depósito de una propuesta.

## `threshold`
### La proporción mínima del poder de voto necesario para que se apruebe una propuesta de gobierno.
#### `cosmoshub-3` por defecto: `0.500000000000000000`

Se requiere una mayoría simple de votos a favor (es decir, el 50% del poder de voto participativo) para que se apruebe una propuesta de gobierno. Aunque es necesario, un voto de mayoría simple 'yes' puede no ser suficiente para aprobar una propuesta en dos escenarios:
1. No se alcanza un [quórum](./Governance.md#quorum) del 40% de la capacidad de la red o
2. Un voto de 'no-with-veto' del 33,4% del poder de voto o mayor.

Si se aprueba una propuesta de gobernanza, las cantidades depositadas se devuelven a los contribuyentes. Si se aprueba una propuesta basada en texto, nada se promulga automáticamente, pero existe una expectativa social de que los participantes se coordinen para promulgar los compromisos señalados en la propuesta. Si se aprueba una propuesta de cambio de parámetros, el parámetro de protocolo cambiará inmediatamente después de que termine el [período de votación](./Governance.md#votingperiod), y sin necesidad de ejecutar un nuevo software. Si se aprueba una propuesta de gasto comunitario, el saldo de la Reserva Comunitaria disminuirá en el número de ATOMs indicados en la propuesta y la dirección del destinatario aumentará en ese mismo número de ATOMs inmediatamente después de que termine el período de votación.

### Posibles consecuencias
#### Disminución del valor `threshold`
La disminución del valor de subkey `threshold` disminuirá la proporción del poder de voto necesario para aprobar una propuesta. Esto puede:
1. aumentará la probabilidad de que una propuesta sea aprobada, y
2. aumentará la probabilidad de que un grupo minoritario realice cambios en la red.

#### Aumentar el valor `threshold`
Aumentar el valor de subkey `threshold` aumentará la proporción de poder de voto necesario para aprobar una propuesta. Esto puede:
1. disminuir la probabilidad de que una propuesta sea aprobada, y
2. disminuir la probabilidad de que un grupo minoritario realice cambios en la red.

## `veto`
### La proporción mínima de poder de voto de los participantes para vetar (es decir, rechazar) una propuesta de gobierno.
#### `cosmoshub-3` por defecto: `0.334000000000000000`

Aunque se requiere un voto de 'yes' por mayoría simple (es decir, el 50% del poder de voto participante) para que se apruebe una propuesta de gobierno, un voto de 'no-with-veto' del 33,4% del poder de voto participante o superior puede anular este resultado y hacer que la propuesta fracase. Esto permite que un grupo minoritario que represente más de 1/3 del poder de voto pueda hacer fracasar una propuesta que de otro modo sería aprobada.

### Posibles consecuencias
#### Disminución del valor `veto`
Disminuir el valor de subkey `veto` disminuirá la proporción de poder de voto de los participantes requerida para vetar. Esto puede:
1. permiten a un grupo minoritario más pequeño evitar que las propuestas sean aprobadas, y
2. disminuyen la probabilidad de que se aprueben propuestas controvertidas.

#### Aumentar el valor `veto`
Aumentar el valor de subkey `veto` aumentará la proporción del poder de voto requerido para vetar. Esto requerirá un grupo minoritario más grande para evitar que las propuestas sean aprobadas, y probablemente aumentará la probabilidad de que se aprueben las propuestas controvertidas.

# Verificar los valores de los parámetros
## Parámetros de Génesis (aka lanzamiento)
Esto es útil si no tienes `gaiad` instalado y no tienes una razón para creer que el parámetro ha cambiado desde que se lanzó la cadena.

Cada parámetro puede ser verificado en el archivo génesis de la cadena, que encuentra [aquí](https://raw.githubusercontent.com/cosmos/launch/master/genesis.json). Estos son los parámetros con los que la última cadena del Hub de Cosmos se lanzó, y seguirá haciéndolo, a menos que una propuesta de gobierno los cambie. He resumido esos valores originales en la sección [Especificaciones Técnicas](./Governance.md#technical-specifications).

El archivo génesis contiene texto y es grande. El esquema de nombres de los parámetros de génesis no es idéntico a los de la lista anterior, así que cuando busco, pongo un guión bajo entre los caracteres en mayúsculas y minúsculas, y luego convierto todos los caracteres a minúsculas.

Por ejemplo, si quiero buscar `depositparams`, buscaré en el [génesis](https://raw.githubusercontent.com/cosmos/launch/master/genesis.json) `deposit_params`.

## Parámetros actuales
Puede verificar los valores actuales de los parámetros (en caso de que hayan sido modificados mediante la propuesta de gobierno posterior al lanzamiento) con la aplicación de [línea de comandos gaiad](params-change/gaiad). Aquí están los comandos para cada uno:
1. `depositparams` - `gaiad q ..` --> **to do** <--

## Futuro

La documentación actual sólo describe el producto mínimo viable para el módulo de gobernanza. Las mejoras futuras pueden incluir:

* **`BountyProposals`:** Si es aceptada, un `BountyProposal` crea una recompensa abierta. El `BountyProposal` especifica cuántos átomos se entregarán al finalizar. Estos átomos serán tomados del `reserve pool`. Después de que un `BountyProposal` es aceptado por el gobierno, cualquiera puede presentar un `SoftwareUpgradeProposal` con el código para reclamar la recompensa. Tenga en cuenta que una vez que el `BountyProposal` es aceptado, los fondos correspondientes en la `reserve pool` se bloquean para que el pago siempre pueda ser cumplido. Para vincular un `SoftwareUpgradeProposal` con una recompensa abierta, el remitente del `SoftwareUpgradeProposal` utilizará el atributo `Proposal.LinkedProposal`. Si un `SoftwareUpgradeProposal` vinculado a una recompensa abierta es aceptado por la administración, los fondos reservados se transfieren automáticamente al autor de la propuesta.
* **Complex delegation:** Los delegadores podrán elegir otros representantes además de sus validadores. En última instancia, la cadena de representantes siempre terminaría en un validador, pero los delegadores podrían heredar el voto de su representante elegido antes de heredar el voto de su validador. En otras palabras, sólo heredarían el voto de su validador si su otro representante designado no votara.
* **Mejor proceso de revisión de propuestas:** La propuesta consta de dos partes de `proposal.Deposit`, uno para la lucha contra el correo basura (igual que en el MVP) y otro para recompensar a los auditores de terceros.

  [origen](https://github.com/cosmos/cosmos-sdk/blob/master/x/gov/spec/05_future_improvements.md)

# Especificaciones técnicas

El módulo `gov` es responsable del sistema de gobierno de la cadena. En este sistema, los titulares del token nativo de la cadena pueden votar sobre las propuestas en una base de 1-token, 1-voto. A continuación hay una lista de las características que el módulo apoya actualmente:

- **Entrega de propuestas**: Los usuarios pueden presentar propuestas con un depósito. Una vez que se alcanza el depósito mínimo, la propuesta entra en el período de votación.
- **Voto**: Los participantes pueden votar sobre las propuestas que llegaron a `MinDeposit`.
- **Herencia y sanciones**: Los delegadores heredan su voto de validación si no votan ellos mismos.
- **Reclamación del depósito**: Los usuarios que depositaron en las propuestas pueden recuperar sus depósitos si la propuesta fue aceptada O si la propuesta nunca entró en el período de votación.

El módulo `gov` contiene los siguientes parámetros:

| Key           | Type   | cosmoshub-3 genesis setting                                                                     |
|---------------|--------|:----------------------------------------------------------------------------------------------------|
| depositparams | object | {"min_deposit":[{"denom":"uatom","amount":"512000000"}],"max_deposit_period":"1209600000000000"}     |
| **Subkeys** |
| min_deposit        | array (coins)    | [{"denom":"uatom","amount":"512000000"}] |
| max_deposit_period | string (time ns) | "1209600000000000"                       |

| Key           | Type   | cosmoshub-3 genesis setting                                                                     |
|---------------|--------|:----------------------------------------------------------------------------------------------------|
| votingparams  | object | {"voting_period":"1209600000000000"}     |
| **Subkey** |
| voting_period      | string (time ns) | "1209600000000000" |

| Key           | Type   | cosmoshub-3 genesis setting                                                                     |
|---------------|--------|:----------------------------------------------------------------------------------------------------|
| depositparams | object | {"min_deposit":[{"denom":"uatom","amount":"512000000"}],"max_deposit_period":"1209600000000000"}     |
| **Subkeys** |
| quorum             | string (dec)     | "0.400000000000000000" |
| threshold          | string (dec)     | "0.500000000000000000"                       |
| veto               | string (dec)     | "0.334000000000000000" |

__Observación__: El módulo de gobierno contiene parámetros que son objetos que no son como los demás módulos. Si sólo se desea modificar un subconjunto de parámetros, sólo hay que incluirlos y no toda la estructura de objetos de parámetros.
