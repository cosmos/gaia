# Cosmos Hub 3 y la Community Pool
La iniciativa Cosmos Hub 3 fue lanzada por parte de la comunidad el 11 de Diciembre de 2019, liberando así la posibilidad de que las personas con tokens puedan votar la aprobación de gastos desde la Community Pool.**Esta documentación es un desarrollo en curso, por favor, de momento no te bases en esta información** [Puedes debatir este desarrollo aquí](https://forum.cosmos.network/t/gwg-community-spend-best-practices/3240).

## ¿Por qué crear una propuesta para utilizar fondos de la Community Pool?
Hay otras opciones de financiación, principalmente el programa de concesión de la Interchain Foundation. ¿Por qué crear una propuesta para gastos?

**Como estrategia: puedes hacer ambas.** Puedes enviar tu propuesta a la Interchain Foundation, pero también tienes la posibilidad sobre la cadena de forma pública. Si el Hub vota a favor, puedes retirar tu solicitud a la Interchain Foundation.

**Como estrategia: la financiación es rápida.** Aparte del tiempo que necesitas para enviar tu propuesta a la cadena, el único factor limitante es un periodo de votación fijo de 14 días. Tan pronto como la propuesta se apruebe, la cantidad total solicitada en tu propuesta será abonada en tu cuenta.

**Para crear buenas relaciones.** Involucrarse públicamente con la comunidad es la oportunidad para establecer relaciones con otras personas interesadas y mostrarles la importancia de tu trabajo. Podrían surgir colaboraciones inesperadas y la comunidad al completo podría valorar más tu trabajo si están involucrados como stakeholders.

**Para ser más independiente.** La Interchain Foundation (ICF) puede no ser capaz de financiar trabajo siempre. Tener una fuente de financiación más constante y con un canal directo a los stakeholders, significa que puedes usar tus buenas relaciones para tener la confianza de ser capaz de encontrar financiación segura sin depender únicamente de la ICF.

## Creación de una propuesta de gastos a la comunidad
Crear y enviar una propuesta es un proceso que lleva tiempo, atención y conlleva riesgo. El objetivo de esta documentación es hacer este proceso más fácil, preparando a los participantes para aquello en lo que deben prestar atención, la información que debe ser incluida en la propuesta, y cómo reducir el riesgo de perder depósitos. Idealmente, una propuesta que no sigue adelante debería ser solamente porque los votantes 1) son conscientes y están involucrados y 2) son capaces de tomar una decisión e informarla votando en la propuesta. 


Si estás considerando realizar una propuesta, deberías conocer:
1. [Sobre la Community Pool](#sobre-la-community-pool)
2. [Cómo funciona el mecanismo de voto y gobernanza](../overview.md#_2-voting-period)
3. [Dónde y cómo involucrar a la comunidad de Cosmos acerca de tu idea](../best_practices.md)
4. [Lo que la comunidad querrá saber sobre tu propuesta](./best_practices.md#elements-of-a-community-spend-proposal)
5. [Cómo preparar tu borrador de propuesta final para ser enviada](../submitting.md)
6. [Cómo enviar tu propuesta al Cosmos Hub testnet & mainnet](../submitting.md)


## Sobre la Community Pool

### ¿Cómo está financiada la Community Pool?
El 2% de todas las delegaciones de fondos generadas (vía recompensa de bloques y tasas de transacción) es continuamente transferido y acumulado en la Community Pool. Por ejemplo, desde el 19 de Diciembre de 2019 hasta el 20 de Enero de 2020 (32 días), se generaron y añadieron a la pool un total de 28,726 ATOM.

### ¿Cómo puede cambiar la financiación de la Community Pool?
Aunque la tasa de financiación está actualmente fijada en el 2% de las delegaciones, la tasa más efectiva es dependiente de los fondos pertenecientes al Cosmos Hub, que puede cambiar con la inflacción y tiempos de bloque. 

La tasa actual del 2% de financiación podría ser modificada con una propuesta de gobernanza y aprobaba de forma inmediata en cuanto la propuesta sea aprobada.

Actualmente, no se pueden enviar fondos a la Community Pool, pero esperamos que esto cambie con la siguiente actualización. Lee más sobre esta nueva funcionalidad [aqui](https://github.com/cosmos/cosmos-sdk/pull/5249). ¿Qué hace que esta funcionalidad sea importante?
1. Los proyectos financiados que finalmente no se ejecuten deben devolver los fondos a la Community Pool;
2. Las entidades podrían ayudar a aportar fondos a la Community Pool mediante aportación directa a la cuenta.

### ¿Cuál es el saldo de la Community Pool?
Puedes solicitar directamente al Cosmos Hub 3 el saldo de la Community Pool:

```gaiad q distribution community-pool --chain-id cosmoshub-3 --node cosmos-node-1.figment.network:26657```

De forma alternativa, los navegadores de Cosmos más populares como [Big Dipper](https://cosmos.bigdipper.live) y [Hubble](https://hubble.figment.network/cosmos/chains/cosmoshub-3) muestran la evolución del saldo de la Community Pool.

### ¿Cómo se pueden gastar los fondos de la Community Pool?
Los fondos de la Cosmos Community Pool pueden ser gastados a través de propuestas de gobernanza aprobadas.

### ¿Cómo se deberían gastar los fondos de la Community Pool?
No lo sabemos 🤷

La suposición principal es que los fondos deberían ser gastados de forma que aporte valor al Cosmos Hub. Sin embargo, hay debate entorno a cómo hacer el fondo sostenible. También hay algún debate acerca de cómo se debería recibir financiación. Por ejemplo, parte de la comunidad cree que los fondos solamente deberían ser utilizados por aquellas personas que más los necesiten. Otros temas de discusión son:  
- concesiones retroactivas
- negociación de precio
- desembolso de fondos (por ejemplo, pagos por fases; pagos fijos para reducir volatilidad)
- revisión drástica de cómo los mecanimos de gastos de comunidad funcionan

Esperamos que todo esto tome forma a medida que las propuestas sean debatidas, aceptadas, y rechazadas por parte de la comunidad Cosmos Hub.

### ¿Cómo se desembolsan los fondos una vez que una prouesta de gastos de comunidad es aprobada?
Si una propuesta de gastos de comunidad es aprobada, el número de ATOM inluidos en la propuesta serán transferidos desde la community pool a la dirección especificada en la propuesta, y esto ocurrirá justo inmediatamente después de que el periodo de votación termine.
