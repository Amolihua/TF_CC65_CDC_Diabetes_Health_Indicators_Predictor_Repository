C:\Users\USER\Documents\2026-1\Programacion Concurrente y Distribuida\ProyectoFinal\TF_CC65_CDC_Diabetes_Health_Indicators_Predictor_Repository\api-coordinador>go run benchmark.go

> > Iniciando Benchmark
> > Dataset: ../datos_raw/diabetes_1M_extended.csv  
> > Nodo ML: localhost:9000

======================================================
EVALUANDO ALGORITMO: softmax
======================================================

-> Prueba con 2 Goroutines (10 iteraciones)...

- Iter 1: Total=4.8419516s | Envio=4.6362093s | Nodo=205.7423ms - Iter 2: Total=4.2969644s | Envio=4.0898285s | Nodo=207.1359ms - Iter 3: Total=4.3219897s | Envio=4.114957s | Nodo=207.0327ms - Iter 4: Total=4.2968013s | Envio=4.0836771s | Nodo=213.1242ms - Iter 5: Total=4.3171942s | Envio=4.1111072s | Nodo=206.087ms - Iter 6: Total=4.3062686s | Envio=4.1043698s | Nodo=201.8988ms - Iter 7: Total=4.3130539s | Envio=4.1013609s | Nodo=211.693ms - Iter 8: Total=4.405701s | Envio=4.1967083s | Nodo=208.9927ms - Iter 9: Total=4.4869416s | Envio=4.2767342s | Nodo=210.2074ms - Iter 10: Total=4.4462858s | Envio=4.2422268s | Nodo=204.059ms

  METRICA -> Algoritmo: softmax | Workers_Coord: 2 | Workers_Nodo: 2 | Tiempo: 4.3617999s | Envio_Prom: 4.19571791s | Nodo_Prom: 207.5973ms | RAM_Heap: 98.76 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 5 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 4 Goroutines (10 iteraciones)... - Iter 1: Total=4.7571852s | Envio=4.6524501s | Nodo=104.7351ms - Iter 2: Total=4.7030415s | Envio=4.5945047s | Nodo=108.5368ms - Iter 3: Total=4.6726863s | Envio=4.5705999s | Nodo=102.0864ms - Iter 4: Total=4.6634579s | Envio=4.5614494s | Nodo=102.0085ms - Iter 5: Total=4.7087256s | Envio=4.6077861s | Nodo=100.9395ms - Iter 6: Total=4.7433224s | Envio=4.6411004s | Nodo=102.222ms - Iter 7: Total=4.6792838s | Envio=4.5627976s | Nodo=116.4862ms - Iter 8: Total=4.721029s | Envio=4.6090084s | Nodo=112.0206ms - Iter 9: Total=4.6894446s | Envio=4.5848821s | Nodo=104.5625ms - Iter 10: Total=4.6994027s | Envio=4.5860482s | Nodo=113.3545ms

METRICA -> Algoritmo: softmax | Workers_Coord: 4 | Workers_Nodo: 4 | Tiempo: 4.702116987s | Envio_Prom: 4.59706269s | Nodo_Prom: 106.69521ms | RAM_Heap: 99.38 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 7 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 8 Goroutines (10 iteraciones)... - Iter 1: Total=4.7045953s | Envio=4.6373584s | Nodo=67.2369ms - Iter 2: Total=4.732004s | Envio=4.666844s | Nodo=65.16ms - Iter 3: Total=4.7577889s | Envio=4.6902624s | Nodo=67.5265ms - Iter 4: Total=4.7292461s | Envio=4.6630112s | Nodo=66.2349ms - Iter 5: Total=4.7254267s | Envio=4.6587581s | Nodo=66.6686ms - Iter 6: Total=4.7387246s | Envio=4.6667437s | Nodo=71.9809ms - Iter 7: Total=4.7564359s | Envio=4.6875758s | Nodo=68.8601ms - Iter 8: Total=4.7359483s | Envio=4.66472s | Nodo=71.2283ms - Iter 9: Total=4.7868008s | Envio=4.7135913s | Nodo=73.2095ms - Iter 10: Total=4.8659977s | Envio=4.7967361s | Nodo=69.2616ms

METRICA -> Algoritmo: softmax | Workers_Coord: 8 | Workers_Nodo: 8 | Tiempo: 4.745296912s | Envio_Prom: 4.6845601s | Nodo_Prom: 68.73673ms | RAM_Heap: 99.56 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 11 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 12 Goroutines (10 iteraciones)... - Iter 1: Total=4.8681946s | Envio=4.8118315s | Nodo=56.3631ms - Iter 2: Total=4.7718233s | Envio=4.7138789s | Nodo=57.9444ms - Iter 3: Total=4.8040037s | Envio=4.7456423s | Nodo=58.3614ms - Iter 4: Total=4.7952832s | Envio=4.7370312s | Nodo=58.252ms - Iter 5: Total=4.7913168s | Envio=4.7297659s | Nodo=61.5509ms - Iter 6: Total=4.8997774s | Envio=4.8321812s | Nodo=67.5962ms - Iter 7: Total=4.8635943s | Envio=4.8047219s | Nodo=58.8724ms - Iter 8: Total=4.7552538s | Envio=4.6947661s | Nodo=60.4877ms - Iter 9: Total=4.7951282s | Envio=4.7379404s | Nodo=57.1878ms - Iter 10: Total=4.7774372s | Envio=4.7203561s | Nodo=57.0811ms

METRICA -> Algoritmo: softmax | Workers_Coord: 12 | Workers_Nodo: 12 | Tiempo: 4.808347662s | Envio_Prom: 4.75281155s | Nodo_Prom: 59.3697ms | RAM_Heap: 99.03 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 15 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 16 Goroutines (10 iteraciones)... - Iter 1: Total=4.7855518s | Envio=4.7375767s | Nodo=47.9751ms - Iter 2: Total=4.8018112s | Envio=4.7381362s | Nodo=63.675ms - Iter 3: Total=4.9570324s | Envio=4.8849044s | Nodo=72.128ms - Iter 4: Total=4.969138s | Envio=4.916921s | Nodo=52.217ms - Iter 5: Total=4.7820143s | Envio=4.7332457s | Nodo=48.7686ms - Iter 6: Total=4.7975417s | Envio=4.7473279s | Nodo=50.2138ms - Iter 7: Total=4.7902355s | Envio=4.7375314s | Nodo=52.7041ms - Iter 8: Total=4.8410612s | Envio=4.7923204s | Nodo=48.7408ms - Iter 9: Total=4.8108076s | Envio=4.7559637s | Nodo=54.8439ms - Iter 10: Total=4.8763403s | Envio=4.822962s | Nodo=53.3783ms

METRICA -> Algoritmo: softmax | Workers_Coord: 16 | Workers_Nodo: 16 | Tiempo: 4.832547712s | Envio_Prom: 4.78668894s | Nodo_Prom: 54.46446ms | RAM_Heap: 99.04 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 19 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

======================================================
EVALUANDO ALGORITMO: random_forest
======================================================

-> Prueba con 2 Goroutines (10 iteraciones)... - Iter 1: Total=7.4781636s | Envio=4.1688369s | Nodo=3.3093267s - Iter 2: Total=7.3335752s | Envio=4.1809013s | Nodo=3.1526739s - Iter 3: Total=7.1806587s | Envio=4.0676226s | Nodo=3.1130361s - Iter 4: Total=7.380064s | Envio=4.1400668s | Nodo=3.2399972s - Iter 5: Total=7.6492694s | Envio=4.4232257s | Nodo=3.2260437s - Iter 6: Total=7.495573s | Envio=4.1874216s | Nodo=3.3081514s - Iter 7: Total=7.3737567s | Envio=4.1841611s | Nodo=3.1895956s - Iter 8: Total=7.5165713s | Envio=4.1333291s | Nodo=3.3832422s - Iter 9: Total=7.2745746s | Envio=4.1114461s | Nodo=3.1631285s - Iter 10: Total=7.2830133s | Envio=4.1119286s | Nodo=3.1710847s

METRICA -> Algoritmo: random_forest | Workers_Coord: 2 | Workers_Nodo: 2 | Tiempo: 7.391911462s | Envio_Prom: 4.17089398s | Nodo_Prom: 3.225628s | RAM_Heap: 97.92 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 5 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 4 Goroutines (10 iteraciones)... - Iter 1: Total=7.9409541s | Envio=4.5797342s | Nodo=3.3612199s - Iter 2: Total=8.1184922s | Envio=4.6715633s | Nodo=3.4469289s - Iter 3: Total=7.8477821s | Envio=4.5562159s | Nodo=3.2915662s - Iter 4: Total=8.0935408s | Envio=4.7518513s | Nodo=3.3416242s - Iter 5: Total=7.8088566s | Envio=4.5753775s | Nodo=3.2334791s - Iter 6: Total=7.7903463s | Envio=4.5884425s | Nodo=3.2019038s - Iter 7: Total=7.837374s | Envio=4.5507535s | Nodo=3.2866205s - Iter 8: Total=7.6899339s | Envio=4.5600831s | Nodo=3.1298508s - Iter 9: Total=7.9426827s | Envio=4.5572545s | Nodo=3.3854282s - Iter 10: Total=7.7182083s | Envio=4.5813707s | Nodo=3.1368376s

METRICA -> Algoritmo: random_forest | Workers_Coord: 4 | Workers_Nodo: 4 | Tiempo: 7.872468112s | Envio_Prom: 4.59726465s | Nodo_Prom: 3.28154592s | RAM_Heap: 98.29 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 7 |
Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 8 Goroutines (10 iteraciones)... - Iter 1: Total=7.8141968s | Envio=4.7553138s | Nodo=3.058883s - Iter 2: Total=8.227575s | Envio=4.7230958s | Nodo=3.5044792s - Iter 3: Total=7.846326s | Envio=4.6704375s | Nodo=3.1758885s - Iter 4: Total=7.7752111s | Envio=4.6393222s | Nodo=3.1358889s - Iter 5: Total=7.9003326s | Envio=4.6291407s | Nodo=3.2711919s - Iter 6: Total=7.6911352s | Envio=4.6103852s | Nodo=3.08075s - Iter 7: Total=7.7751239s | Envio=4.6018681s | Nodo=3.1732558s - Iter 8: Total=7.7209782s | Envio=4.637501s | Nodo=3.0834772s - Iter 9: Total=8.1531135s | Envio=4.6165336s | Nodo=3.5365799s - Iter 10: Total=8.0916578s | Envio=4.6705194s | Nodo=3.4211384s

METRICA -> Algoritmo: random_forest | Workers_Coord: 8 | Workers_Nodo: 8 | Tiempo: 7.884617487s | Envio_Prom: 4.65541173s | Nodo_Prom: 3.24415328s | RAM_Heap: 98.52 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 11 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 12 Goroutines (10 iteraciones)... - Iter 1: Total=7.997463s | Envio=4.813372s | Nodo=3.184091s - Iter 2: Total=7.9889921s | Envio=4.678077s | Nodo=3.3109151s - Iter 3: Total=7.9057824s | Envio=4.718018s | Nodo=3.1877644s - Iter 4: Total=7.8740916s | Envio=4.7160705s | Nodo=3.1580211s - Iter 5: Total=7.7628046s | Envio=4.6591715s | Nodo=3.1036331s - Iter 6: Total=8.1022741s | Envio=4.7218906s | Nodo=3.3803835s - Iter 7: Total=7.6857973s | Envio=4.7410378s | Nodo=2.9447595s - Iter 8: Total=7.8023966s | Envio=4.6575752s | Nodo=3.1448214s - Iter 9: Total=8.258221s | Envio=4.8928809s | Nodo=3.3653401s - Iter 10: Total=7.8247532s | Envio=4.7021922s | Nodo=3.122561s

METRICA -> Algoritmo: random_forest | Workers_Coord: 12 | Workers_Nodo: 12 | Tiempo: 7.9073197s | Envio_Prom: 4.73002857s | Nodo_Prom: 3.19022902s | RAM_Heap: 98.34 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 15 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 16 Goroutines (10 iteraciones)... - Iter 1: Total=7.8782171s | Envio=4.7438011s | Nodo=3.134416s - Iter 2: Total=8.107482s | Envio=4.7256107s | Nodo=3.3818713s - Iter 3: Total=7.8629258s | Envio=4.7080948s | Nodo=3.154831s - Iter 4: Total=8.3296581s | Envio=4.7481884s | Nodo=3.5814697s - Iter 5: Total=8.0103617s | Envio=4.7964623s | Nodo=3.2138994s - Iter 6: Total=8.1328983s | Envio=4.7865501s | Nodo=3.3463482s - Iter 7: Total=8.4401552s | Envio=4.8699477s | Nodo=3.5702075s - Iter 8: Total=7.8669997s | Envio=4.7266798s | Nodo=3.1403199s - Iter 9: Total=7.7231998s | Envio=4.7393352s | Nodo=2.9838646s - Iter 10: Total=8.2190574s | Envio=4.8764659s | Nodo=3.3425915s

METRICA -> Algoritmo: random_forest | Workers_Coord: 16 | Workers_Nodo: 16 | Tiempo: 8.050950012s | Envio_Prom: 4.7721136s | Nodo_Prom: 3.28498191s | RAM_Heap: 98.44 MB | RAM_Alloc_Delta: 954.81 MB | Pico_Goroutines: 19
| Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

======================================================
EVALUANDO ALGORITMO: naive_bayes
======================================================

-> Prueba con 2 Goroutines (10 iteraciones)... - Iter 1: Total=4.204684s | Envio=4.114173s | Nodo=90.511ms - Iter 2: Total=4.2723256s | Envio=4.1438935s | Nodo=128.4321ms - Iter 3: Total=4.2270869s | Envio=4.1310647s | Nodo=96.0222ms - Iter 4: Total=4.2126376s | Envio=4.1225125s | Nodo=90.1251ms - Iter 5: Total=4.2731952s | Envio=4.176623s | Nodo=96.5722ms - Iter 6: Total=4.2467355s | Envio=4.1532174s | Nodo=93.5181ms - Iter 7: Total=4.2100166s | Envio=4.1151951s | Nodo=94.8215ms - Iter 8: Total=4.1845382s | Envio=4.0940974s | Nodo=90.4408ms - Iter 9: Total=4.3277169s | Envio=4.2352069s | Nodo=92.51ms - Iter 10: Total=4.3726555s | Envio=4.2754487s | Nodo=97.2068ms

METRICA -> Algoritmo: naive_bayes | Workers_Coord: 2 | Workers_Nodo: 2 | Tiempo: 4.246799787s | Envio_Prom: 4.15614322s | Nodo_Prom: 97.01598ms | RAM_Heap: 96.30 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 5 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 4 Goroutines (10 iteraciones)... - Iter 1: Total=4.6798142s | Envio=4.6338147s | Nodo=45.9995ms - Iter 2: Total=4.6885904s | Envio=4.6418174s | Nodo=46.773ms - Iter 3: Total=4.635208s | Envio=4.5892088s | Nodo=45.9992ms - Iter 4: Total=4.6796803s | Envio=4.6315368s | Nodo=48.1435ms - Iter 5: Total=4.6941908s | Envio=4.6444179s | Nodo=49.7729ms - Iter 6: Total=4.6489417s | Envio=4.5888196s | Nodo=60.1221ms - Iter 7: Total=4.699252s | Envio=4.652999s | Nodo=46.253ms - Iter 8: Total=4.6042832s | Envio=4.5592857s | Nodo=44.9975ms - Iter 9: Total=4.7234216s | Envio=4.6724811s | Nodo=50.9405ms - Iter 10: Total=4.6739668s | Envio=4.6245167s | Nodo=49.4501ms

METRICA -> Algoritmo: naive_bayes | Workers_Coord: 4 | Workers_Nodo: 4 | Tiempo: 4.674955525s | Envio_Prom: 4.62388977s | Nodo_Prom: 48.84513ms | RAM_Heap: 97.14 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 7 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 8 Goroutines (10 iteraciones)... - Iter 1: Total=4.7214974s | Envio=4.689498s | Nodo=31.9994ms - Iter 2: Total=4.7329523s | Envio=4.6977281s | Nodo=35.2242ms - Iter 3: Total=4.7686335s | Envio=4.7211363s | Nodo=47.4972ms - Iter 4: Total=4.8561047s | Envio=4.8218122s | Nodo=34.2925ms - Iter 5: Total=4.7463906s | Envio=4.7099434s | Nodo=36.4472ms - Iter 6: Total=4.7362282s | Envio=4.6958083s | Nodo=40.4199ms - Iter 7: Total=4.6927403s | Envio=4.6571871s | Nodo=35.5532ms - Iter 8: Total=4.816574s | Envio=4.7757378s | Nodo=40.8362ms - Iter 9: Total=4.7849959s | Envio=4.7512703s | Nodo=33.7256ms - Iter 10: Total=4.71299s | Envio=4.6801763s | Nodo=32.8137ms

METRICA -> Algoritmo: naive_bayes | Workers_Coord: 8 | Workers_Nodo: 8 | Tiempo: 4.752532737s | Envio_Prom: 4.72002978s | Nodo_Prom: 36.88091ms | RAM_Heap: 98.06 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 11 | Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 12 Goroutines (10 iteraciones)... - Iter 1: Total=4.7662648s | Envio=4.7372667s | Nodo=28.9981ms - Iter 2: Total=4.7932831s | Envio=4.7622817s | Nodo=31.0014ms - Iter 3: Total=4.7690102s | Envio=4.7408s | Nodo=28.2102ms - Iter 4: Total=4.7719875s | Envio=4.7431417s | Nodo=28.8458ms - Iter 5: Total=4.7664185s | Envio=4.73733s | Nodo=29.0885ms - Iter 6: Total=4.7873519s | Envio=4.758131s | Nodo=29.2209ms - Iter 7: Total=4.8548804s | Envio=4.8217925s | Nodo=33.0879ms - Iter 8: Total=4.9020353s | Envio=4.8739181s | Nodo=28.1172ms - Iter 9: Total=4.7749521s | Envio=4.7453723s | Nodo=29.5798ms - Iter 10: Total=4.7703005s | Envio=4.7393016s | Nodo=30.9989ms

METRICA -> Algoritmo: naive_bayes | Workers_Coord: 12 | Workers_Nodo: 12 | Tiempo: 4.786023025s | Envio_Prom: 4.76593356s | Nodo_Prom: 29.71487ms | RAM_Heap: 98.03 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 15 |
Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

-> Prueba con 16 Goroutines (10 iteraciones)... - Iter 1: Total=4.8278124s | Envio=4.793326s | Nodo=34.4864ms - Iter 2: Total=4.8187485s | Envio=4.7895868s | Nodo=29.1617ms - Iter 3: Total=4.8182756s | Envio=4.7845031s | Nodo=33.7725ms - Iter 4: Total=4.7972099s | Envio=4.7684209s | Nodo=28.789ms - Iter 5: Total=4.8256464s | Envio=4.7961917s | Nodo=29.4547ms - Iter 6: Total=4.7622733s | Envio=4.737403s | Nodo=24.8703ms - Iter 7: Total=4.7925099s | Envio=4.7674804s | Nodo=25.0295ms - Iter 8: Total=4.8140715s | Envio=4.7840596s | Nodo=30.0119ms - Iter 9: Total=4.7896148s | Envio=4.7658044s | Nodo=23.8104ms - Iter 10: Total=4.8133907s | Envio=4.7829007s | Nodo=30.49ms

METRICA -> Algoritmo: naive_bayes | Workers_Coord: 16 | Workers_Nodo: 16 | Tiempo: 4.808683412s | Envio_Prom: 4.77696766s | Nodo_Prom: 28.98764ms | RAM_Heap: 97.65 MB | RAM_Alloc_Delta: 954.80 MB | Pico_Goroutines: 19 |
Ciclos_GC_Delta: 15 | CPUs: 16 | Total_Registros: 2536800

> > Benchmark finalizado en 14m19.5590734s.
