$scriptStart = Get-Date
$outputFile = "resultados_escalabilidad_avanzado.txt"
Clear-Content $outputFile -ErrorAction SilentlyContinue

function Log($msg) {
    Write-Host $msg -ForegroundColor Cyan
    Add-Content $outputFile $msg
}

Log "=========================================================="
Log " BATERIA DE PRUEBAS DE ESCALABILIDAD AUTOMATICA (AVANZADO)"
Log "=========================================================="

# Crear un payload.json físico para evitar problemas de escape de comillas al pasarlo a Docker o curl
$jsonPayload = '{"HighBP": 1, "HighChol": 1, "CholCheck": 1, "BMI": 30, "Smoker": 1, "Stroke": 0, "HeartDiseaseorAttack": 1, "PhysActivity": 0, "Fruits": 0, "Veggies": 0, "HvyAlcoholConsump": 0, "AnyHealthcare": 1, "NoDocbcCost": 0, "GenHlth": 5, "MentHlth": 15, "PhysHlth": 15, "DiffWalk": 1, "Sex": 1, "Age": 10, "Education": 4, "Income": 2}'
Set-Content -Path ".\payload.json" -Value $jsonPayload -Encoding UTF8

# Funcion para realizar las 10 iteraciones y calcular la media recortada
function Run-Benchmark {
    param([string]$TestName)
    
    Log ">> Ejecutando 10 iteraciones para: $TestName"
    $rpsList = @()
    $cpuList = @()
    $ramList = @()

    for ($k = 1; $k -le 10; $k++) {
        # 1. Ejecutar hey montando el archivo payload.json (evita errores de parseo JSON)
        $heyOut = docker run --rm -v "${PWD}:/app" williamyeh/hey -n 2000 -c 100 -m POST -T "application/json" -D /app/payload.json http://host.docker.internal:8080/api/predict | Select-String "Requests/sec:" | Out-String
        $rpsStr = $heyOut -replace "[^\d\.]", ""
        $rps = 0
        if ([double]::TryParse($rpsStr, [ref]$rps)) {}

        # 2. Capturar CPU y RAM del contenedor api-coordinador
        $stats = docker stats --no-stream --format "{{.CPUPerc}}|{{.MemUsage}}" tf_cc65_cdc_diabetes_health_indicators_predictor_repository-api-coordinador-1
        $cpuStr = ($stats -split "\|")[0] -replace "%", ""
        $ramStr = (($stats -split "\|")[1] -split "/")[0].Trim()
        
        $cpu = 0
        if ([double]::TryParse($cpuStr, [ref]$cpu)) {}
        
        $ramVal = 0
        if ($ramStr -match "MiB") {
            $ramVal = [double]($ramStr -replace "MiB", "")
        }
        elseif ($ramStr -match "GiB") {
            $ramVal = [double]($ramStr -replace "GiB", "") * 1024
        }
        elseif ($ramStr -match "KiB") {
            $ramVal = [double]($ramStr -replace "KiB", "") / 1024
        }
        elseif ($ramStr -match "B") {
            $ramVal = [double]($ramStr -replace "B", "") / (1024 * 1024)
        }
        
        $rpsList += $rps
        $cpuList += $cpu
        $ramList += $ramVal
    }

    # Ordenar y aplicar Media Recortada (quitando el 1ero y el ultimo)
    $rpsList = $rpsList | Sort-Object
    $cpuList = $cpuList | Sort-Object
    $ramList = $ramList | Sort-Object

    $rpsTrim = 0; $cpuTrim = 0; $ramTrim = 0
    for ($x = 1; $x -le 8; $x++) {
        $rpsTrim += $rpsList[$x]
        $cpuTrim += $cpuList[$x]
        $ramTrim += $ramList[$x]
    }
    $rpsAvg = $rpsTrim / 8
    $cpuAvg = $cpuTrim / 8
    $ramAvg = $ramTrim / 8

    Log "   [RESULTADO FINAL] $TestName -> RPS: $([math]::Round($rpsAvg, 2)) req/s | CPU: $([math]::Round($cpuAvg, 2))% | RAM: $([math]::Round($ramAvg, 2)) MB"
}

foreach ($i in 1, 2, 3, 4, 5, 10) {
    Log ""
    Log "----------------------------------------------------------"
    Log " PRUEBA ESCALABILIDAD CON $i NODOS ML"
    Log "----------------------------------------------------------"
    
    # 1. Configurar .env dinamicamente
    $nodos = @()
    for ($j = 1; $j -le $i; $j++) {
        $nodos += "nodo-ml-$($j):9000"
    }
    $nodosStr = $nodos -join ","
    
    (Get-Content .env) -replace "^NODOS_ML_ADDRS=.*", "NODOS_ML_ADDRS=$nodosStr" | Set-Content .env
    
    Log ">> Reiniciando cluster con $i nodos en Docker..."
    docker-compose down 2>$null
    docker-compose up -d --build 2>$null
    
    Log ">> Esperando 15 segundos para que los nodos levanten y se estabilicen..."
    Start-Sleep -Seconds 15
    
    # 2. Login (Usando Invoke-RestMethod nativo de PowerShell para evitar que curl.exe rompa las comillas del JSON)
    Log ">> Obteniendo token JWT..."
    $loginRes = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/login" -Method Post -ContentType "application/json" -Body '{"username": "admin", "password": "admin123"}'
    $token = $loginRes.token
    
    # 3. Entrenar
    Log ">> Entrenando modelo distribuido en $i nodos (10 iteraciones para media recortada)..."
    $trainTimes = @()
    for ($k = 1; $k -le 10; $k++) {
        $trainRes = curl.exe -s -X POST http://127.0.0.1:8080/api/train -H "Authorization: Bearer $token" -F "dataset=@.\datos_raw\diabetes_1M_extended.csv"
        $trainTimeStr = ($trainRes | ConvertFrom-Json).time_elapsed
        $tStr = $trainTimeStr -replace "[^\d\.]", ""
        $t = 0
        if ([double]::TryParse($tStr, [System.Globalization.NumberStyles]::Any, [System.Globalization.CultureInfo]::InvariantCulture, [ref]$t)) {
            $trainTimes += $t
        }
    }
    
    $trainTimes = $trainTimes | Sort-Object
    $trainTrim = 0
    for ($x = 1; $x -le 8; $x++) {
        $trainTrim += $trainTimes[$x]
    }
    $trainAvg = $trainTrim / 8
    Log "   [RESULTADO FINAL] TIEMPO ENTRENAMIENTO -> $([math]::Round($trainAvg, 3)) s"
    
    # 4. Prueba CON Redis
    Run-Benchmark -TestName "CON REDIS (Cache Hit)"
    
    # 5. Apagar Redis
    Log ">> Apagando contenedor Redis (Simulando Caida del Cache)..."
    docker stop tf_cc65_cdc_diabetes_health_indicators_predictor_repository-redis-1 | Out-Null
    
    # 6. Prueba SIN Redis
    Run-Benchmark -TestName "SIN REDIS (RAM + Singleflight)"
}

$scriptEnd = Get-Date
$totalTime = $scriptEnd - $scriptStart

Log "=========================================================="
Log " PRUEBAS FINALIZADAS."
Log " TIEMPO TOTAL DE EJECUCION: $($totalTime.Hours)h $($totalTime.Minutes)m $($totalTime.Seconds)s"
Log " REVISAR ARCHIVO: $outputFile"
Log "=========================================================="
