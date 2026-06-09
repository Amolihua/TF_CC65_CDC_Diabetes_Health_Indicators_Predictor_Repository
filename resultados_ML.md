[RANDOM FOREST] Consolidacion de 50 arboles concluida exitosamente.
C:\Users\USER\Documents\2026-1\Programacion Concurrente y Distribuida\ProyectoFinal\TF_CC65_CDC_Diabetes_Health_Indicators_Predictor_Repository\api-coordinador>go run main.go
[API-COORDINADOR] Dataset de 2536800 registros enviado a localhost:9000 en: 1m15.0399774s

========================================================
Matriz de Confusión [Fila=Real, Col=Predicha]:
1138920 361050 637060
8430 9620 28260
36740 50100 266620

## Clase | Precision | Recall | F1-Score

0 | 0.9619 | 0.5329 | 0.6859
1 | 0.0229 | 0.2077 | 0.0412
2 | 0.2861 | 0.7543 | 0.4148
========================================================

[SOFTMAX] Entrenamiento iterativo concluido exitosamente.

C:\Users\USER\Documents\2026-1\Programacion Concurrente y Distribuida\ProyectoFinal\TF_CC65_CDC_Diabetes_Health_Indicators_Predictor_Repository\api-coordinador>go run main.go
[API-COORDINADOR] Dataset de 2536800 registros enviado a localhost:9000 en: 56.099476s
========================================================
Matriz de Confusión [Fila=Real, Col=Predicha]:
1295609 0 841421
24110 0 22200
182160 0 171300

## Clase | Precision | Recall | F1-Score

0 | 0.8627 | 0.6063 | 0.7121
1 | 0.0000 | 0.0000 | 0.0000
2 | 0.1655 | 0.4846 | 0.2468
========================================================

[NAIVE BAYES] Map-Reduce estadistico concluido exitosamente.
C:\Users\USER\Documents\2026-1\Programacion Concurrente y Distribuida\ProyectoFinal\TF_CC65_CDC_Diabetes_Health_Indicators_Predictor_Repository\api-coordinador>go run main.go
[API-COORDINADOR] Dataset de 2536800 registros enviado a localhost:9000 en: 17.8576009s
========================================================
Matriz de Confusión [Fila=Real, Col=Predicha]:
1540980 251810 344240
18810 9180 18320
99060 57780 196620

## Clase | Precision | Recall | F1-Score

0 | 0.9289 | 0.7211 | 0.8119
1 | 0.0288 | 0.1982 | 0.0503
2 | 0.3516 | 0.5563 | 0.4309
========================================================
