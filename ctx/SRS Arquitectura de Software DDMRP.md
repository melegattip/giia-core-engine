Arquitectura de Software DDMRP: Gestión de Buffer y Flujo  
El siguiente documento representa la **Especificación de Requisitos de Software (SRS)** para el programa de Gestión de Inventario IA (GIIA), enfocado en la metodología **Demand Driven Material Requirements Planning (DDMRP)** \[1\]. Asume un rol único de **Usuario (Demand Planner)**, quien es responsable tanto de las operaciones diarias como de la administración estratégica del modelo.  
\--------------------------------------------------------------------------------  
Especificación de Requisitos de Software (SRS) \- GIIA V2.0  
1\. Introducción  
El sistema GIIA se basa en los **cinco componentes de DDMRP** \[1, 2\] y tiene como objetivo optimizar la gestión de inventario, buscando alcanzar un **alto servicio con bajo inventario** y **menos urgencias** \[3\]. El principio fundamental del modelo es el **Flujo** \[3, 4\], utilizando **buffers dinámicos** para absorber la variabilidad y comprimir los *lead times* \[4-6\].  
El objetivo final es maximizar el **Retorno sobre el Capital Invertido (ROCE)** \[3\].  
2\. Historias de Usuario (Rol: Usuario / Demand Planner)  
El Usuario tiene control total sobre la configuración estratégica (Administración del modelo) y la ejecución operativa diaria \[7\].

| ID | Historia de Usuario | Categoría de Requisito | DDMRP Concepto Clave | Fuentes |
| :---- | :---- | :---- | :---- | :---- |
| **HU-U-1** | Como **Usuario**, quiero configurar, dar de alta y modificar (**ABM**) las **Propiedades de Planeación de la Parte** (Tipo de artículo, Lead Time, MOQ, CPD) para definir la base de mis cálculos de buffer \[8-10\]. | Configuración Estratégica | Atributos individuales de la parte \[8, 9\] | \[8-10\] |
| **HU-U-2** | Como **Usuario**, necesito gestionar completamente (**ABM**) y asignar los **Perfiles de Buffer** que combinan la Categoría de Lead Time (L, M, C) y la Categoría de Variabilidad (A, M, B), para reflejar la realidad operativa de mis SKUs. | Configuración Estratégica | Perfiles y niveles de buffer \[1, 11\] | \[1, 11-14\] |
| **HU-U-3** | Como **Usuario**, quiero visualizar los niveles de mis **Buffers (Rojo, Amarillo, Verde)** inmediatamente para evaluar la salud de mi inventario y la **Prioridad de planeación** de forma relativa respecto a otros artículos \[15, 16\]. | Ejecución Visible | Estado del buffer \[3\] | \[15-18\] |
| **HU-U-4** | Como **Usuario**, necesito que el sistema calcule diariamente la **Ecuación de Flujo Neto (EFN)** \[19\] y genere una **Sugerencia de Reposición** (OC/OP) si la posición se encuentra en la **Zona Roja o Amarilla** \[15, 20\]. | Planeación Operativa | Ecuación de Flujo Neto \[19, 21, 22\] | \[19-22\] |
| **HU-U-5** | Como **Usuario**, quiero ser alertado proactivamente sobre órdenes en tránsito que muestren **Suministro tardío, Inicio antes del compromiso o Suministro insuficiente** (Alertas de sincronización) \[23\] para gestionar la **ejecución visible y colaborativa** \[24\]. | Ejecución Proactiva | Alerta de Lead Time \[25\] | \[23, 25\] |
| **HU-U-6** | Como **Usuario**, requiero aplicar **Factores de Ajuste Planeado (FAP)** (ajustes dinámicos) para manipular temporalmente la ecuación del buffer \[26\] y poder nivelar la carga de recursos ante una demanda futura o gestionar estacionalidades \[27\]. | Ajustes Dinámicos | FAP \[26, 28, 29\] | \[26, 27, 29, 30\] |
| **HU-U-7** | Como **Usuario**, necesito ver **Alertas de Agotado Proyectado** \[31\] basadas en el inventario físico actual, el CPD \[32\] y el calendario de órdenes de suministro, para detectar posibles problemas de ejecución a corto plazo \[31\]. | Inteligencia del Modelo | Alerta de Inventario Físico \[31\] | \[31, 32\] |
| **HU-U-8** | Como **Usuario**, quiero que el sistema analice el desempeño histórico (Análisis de Varianza) para proponer **cambios estratégicos de política** (ej. reducción de Lead Time o MOQ) y **proyecciones del modelo** (espacio, capacidad de carga) \[33-36\]. | Análisis DDS\&OP | Controlar, medir, adaptar y proyectar \[33\] | \[33-36\] |

\--------------------------------------------------------------------------------  
3\. Requisitos Funcionales (RF)  
Los requisitos funcionales describen lo que el sistema GIIA debe hacer, enfocándose en la planeación (generación de órdenes de reposición) y la ejecución DDMRP \[15, 20\].  
3.1. RF: Gestión de Datos Maestros y ABM (Administración Total por el Usuario)  
\*\*RF 3.1.1. Gestión Completa de Entidades Maestras (ABM):\*\*El sistema debe proporcionar una interfaz unificada para que el usuario pueda realizar el **Alta, Baja y Modificación** de todas las entidades críticas para el modelo \[7\]: a. **Productos/Partes (SKU):** Incluyendo Tipo de artículo, Lead Time, MOQ y Costo unitario \[8-10\]. b. **Perfiles de Buffer:** Configuración de grupo basada en Lead Time y Variabilidad \[11, 12\]. c. **Proveedores:** Requeridos para la gestión de Lead Time de Compras \[25, 37\]. d. **Nodos/Centros de Distribución:** El sistema debe gestionar que **cada nodo tenga un buffer único por cada artículo** distribuido \[38\].  
\*\*RF 3.1.2. Gestión del Consumo Promedio Diario (CPD):\*\*El usuario debe poder configurar la fórmula para el cálculo del CPD, seleccionando entre datos del **pasado, futuro o un enfoque mixto** \[39-42\], asegurando que el CPD no esté drásticamente subestimado o sobreestimado \[39, 43\].  
\*\*RF 3.1.3. Gestión de Demanda Calificada:\*\*El sistema debe incluir en la **Demanda Calificada** \[21, 44\] los **pedidos de venta vencidos**, los que se deben entregar **hoy**, o los **futuros picos calificados** \[21, 44, 45\]. Para eventos (promociones cortas), el usuario debe poder tratarlos como una orden de venta interna calificada que se cancela en la fecha de entrega, evitando afectar el CPD \[46\].  
3.2. RF: Lógica del Modelo DDMRP  
\*\*RF 3.2.1. Cálculo Dinámico de Buffer:\*\*El sistema debe calcular las **Zonas (Verde, Amarillo, Roja)** del buffer para cada parte, utilizando la combinación de la **Configuración de Grupo (Perfil de Buffer)** con las **Propiedades Individuales de las partes** \[8, 9\].  
• La **Zona Roja** (seguridad) debe calcularse mediante tres ecuaciones secuenciales (Rojo Base, Rojo de Seguridad, y Zona Roja Total) \[47\].  
\*\*RF 3.2.2. Aplicación de la Ecuación de Flujo Neto (EFN):\*\*La **EFN** debe calcularse **diariamente para todos los artículos en buffer** \[20, 22\] y se define como:

*FlujoNeto*\=*Inventariof*ı*sico*\+*Inventarioentransito*−*Demandacalificada*

\[19, 21, 44\].  
\*\*RF 3.2.3. Generación de Órdenes de Reposición:\*\*El sistema solo debe generar órdenes de reposición para los artículos cuya EFN esté en la **Zona Roja o Amarilla** \[20\]. El sistema debe utilizar el buffer como el **mecanismo principal de planeación** \[16\].  
\*\*RF 3.2.4. Explosión Desacoplada (Producción):\*\*Cuando se genera una orden de reposición, el sistema debe realizar una **explosión desacoplada** de materiales \[48\], que se detiene en cada uno de los buffers, independientemente de la cadena de suministro \[48\].  
\*\*RF 3.2.5. Optimización de la Reposición:\*\*El sistema debe incluir la funcionalidad de **Asignación Priorizada** \[49\] para optimizar la generación de órdenes, respetando restricciones como el lote mínimo (MOQ) o la carga del transporte \[50, 51\], basándose en la EFN relativa al resto del grupo \[15\]. Los usos principales incluyen **optimización de cobertura, optimización de descuento y optimización de carga** \[49\].  
3.3. RF: Mantenimiento y Análisis Estratégico  
\*\*RF 3.3.1. Aplicación de Ajustes Planeados (FAP):\*\*El usuario debe poder ingresar y aplicar Factores de Ajuste Planeado (FAP) \[28\] que son **manipulaciones a la ecuación del buffer** para aumentar o disminuir los niveles de inventario en un momento específico \[26, 29\].  
\*\*RF 3.3.2. Proyección de Recursos (DDS\&OP):\*\*El sistema debe proyectar el impacto de los cambios de política DDMRP en los recursos, incluyendo la **conversión del Inventario Físico Objetivo en Requisitos de Espacio (posiciones por estiba)** \[34, 35\] y la **Proyección de Capacidad de Carga** para recursos críticos \[36\].  
\*\*RF 3.3.3. Reportes de Varianza y Desempeño:\*\*El sistema debe generar análisis enfocados en la **Integridad de la señal** (puntualidad y exactitud), la **Velocidad del modelo**, y la **Integridad de los puntos de desacoplamiento** \[52, 53\]. Debe generar informes de **eventos atípicos** (escasez o exceso) y de **incumplimiento de proveedores** \[54\].  
\--------------------------------------------------------------------------------  
4\. Requisitos No Funcionales (RNF)

| Categoría | ID | Requisito Detallado | Fuentes |
| :---- | :---- | :---- | :---- |
| **Usabilidad (UX) / Visibilidad** | RNF 4.1.1 | **Prioridad Relativa (Execution Dashboard):** La interfaz debe mostrar la **prioridad de planeación** (color del buffer) y la **posición del flujo neto** como un porcentaje relativo \[15, 16\], facilitando la toma de decisiones \[55\]. | \[15, 16, 18, 55\] |
| RNF 4.1.2 | **Diseño de Soporte a la Decisión:** El sistema debe tener un diseño visual e intuitivo, ayudando activamente a la toma de decisiones \[18\]. | \[18\] |  |
| **Rendimiento** | RNF 4.2.1 | **Cálculo Diario de EFN:** La **Ecuación de Flujo Neto** debe calcularse diariamente para todos los artículos que tienen un buffer \[20\]. | \[20\] |
| RNF 4.2.2 | **Retroalimentación en Tiempo Real:** El sistema debe actualizarse de forma continua con información sobre las entregas para mantener la precisión \[56\]. | \[56\] |  |
| **Seguridad y Confianza** | RNF 4.3.1 | **Validez de Planeación:** La implementación de DDMRP debe resultar en una planeación más realista, con horizontes de planeación más cortos y una menor variabilidad que se propague a través del sistema \[6\]. | \[6, 57\] |
| **Modelo de Negocio** | RNF 4.4.1 | **Suscripción:** El modelo de comercialización del software se planteará basado en la **suscripción (mensual, anual o semestral)** \[56\]. | \[56\] |
| **Motivación y Adopción** | RNF 4.5.1 | **Gamificación:** El sistema debe incluir **desafíos** para incentivar al usuario a mejorar su desempeño en la gestión, lo que permitirá **desbloquear funcionalidades que añadan valor al sistema** \[56\]. | \[56\] |

\--------------------------------------------------------------------------------  
5\. Requisitos Técnicos (RT)  
Los requisitos técnicos se centran en la infraestructura y las capacidades internas necesarias para soportar los complejos cálculos de DDMRP y las funciones de administración total del usuario.

| ID | Requisito Técnico | Detalle | Fuentes |
| :---- | :---- | :---- | :---- |
| **RT 5.1** | **Motor de Cálculo DDMRP Optimizado:** | Debe haber un motor de procesamiento capaz de calcular la EFN y los complejos niveles de buffer (Rojo Base, Rojo de Seguridad, etc.) eficientemente y diariamente \[20, 47\]. Este motor es clave, ya que los MRP tradicionales no pueden calcular la EFN para generar la explosión de acuerdo con las tácticas de DDMRP \[58\]. | \[20, 47, 58\] |
| **RT 5.2** | **Arquitectura de Integración (APIs):** | El sistema debe diseñarse para la integración con sistemas ERP (o sistemas de gestión de inventario/ventas) para recibir automáticamente las entradas de **Inventario físico, Órdenes en tránsito y Demanda calificada** \[7, 21\]. | \[7, 21\] |
| **RT 5.3** | **Soporte para Múltiples Nodos/Ubicaciones:** | La arquitectura de datos debe soportar la gestión simultánea de múltiples **Nodos** o centros de distribución, donde cada artículo en cada nodo requiere un buffer único \[38\]. | \[38\] |
| **RT 5.4** | **Plataforma de Interfaz de Gestión Única:** | Se requiere una interfaz (dashboard) que facilite la **Administración del modelo y parametrización de las partes (Configuración maestra)**, permitiendo al usuario configurar perfiles y gestionar todos los inputs \[7\]. | \[7\] |
| **RT 5.5** | **Algoritmos de Proyección/AI:** | Debe implementar capacidades de simulación y proyección para mostrar el rendimiento futuro (Ej. **proyección de capacidad de carga** en minutos o **requisitos de espacio** en estibas) basado en el CPD proyectado y los cambios del modelo \[35, 36\]. | \[35, 36\] |

\--------------------------------------------------------------------------------  
Resumen del Modelo de Planeación DDMRP  
El sistema GIIA se centra en el **DDMRP**, que representa un cambio fundamental respecto a la planeación tradicional \[3\]. Mientras que el *Stock de seguridad* y el *Punto de reorden* se basan en el pronóstico y a menudo fallan en desacoplar el *lead time* o generar órdenes independientes \[59, 60\], el **Buffer DDMRP** sí está diseñado para **desacoplar el *lead time*** \[5\], tiene **independencia de órdenes** (no está reservado a una orden específica) \[61\] y es el **mecanismo principal de planeación** \[16\].  
