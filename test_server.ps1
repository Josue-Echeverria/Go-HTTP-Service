# Script para probar el servidor HTTP
Write-Host "Probando servidor HTTP personalizado..." -ForegroundColor Green

# Probar /ping
Write-Host "`nProbando GET /ping..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8090/ping" -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Body: $($response.Content)" -ForegroundColor Cyan
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

# Probar /status
Write-Host "`nProbando GET /status..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8090/status" -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Body: $($response.Content)" -ForegroundColor Cyan
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

# Probar /time
Write-Host "`nProbando GET /time..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8090/time" -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Body: $($response.Content)" -ForegroundColor Cyan
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

# Probar /echo con POST
Write-Host "`nProbando POST /echo..." -ForegroundColor Yellow
try {
    $body = '{"message":"Hello from test"}'
    $response = Invoke-WebRequest -Uri "http://localhost:8090/echo" -Method POST -Body $body -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Headers devueltos en la respuesta:" -ForegroundColor Cyan
    Write-Host $response.Headers
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "`n✅ Pruebas completadas!" -ForegroundColor Green
