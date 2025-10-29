# Script de teste para API

Write-Host "🧪 Testando API Cacadores de Corrida" -ForegroundColor Cyan
Write-Host ""

# Teste 1: Health Check
Write-Host "✅ Teste 1: Health Check" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:18020/api/healthz" -Method GET
    Write-Host "Resultado: " -NoNewline
    $health | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "❌ ERRO: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Teste 2: Classificar Intenção - Limite
Write-Host "✅ Teste 2: Classificar 'quero mais limite'" -ForegroundColor Yellow
try {
    $body = @{
        intent = "quero mais limite"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "http://localhost:18020/api/find-service" -Method POST -Body $body -ContentType "application/json"
    Write-Host "Resultado: " -NoNewline
    $response | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "❌ ERRO: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Teste 3: Classificar Intenção - Perda de Cartão
Write-Host "✅ Teste 3: Classificar 'perdi meu cartão'" -ForegroundColor Yellow
try {
    $body = @{
        intent = "perdi meu cartão"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "http://localhost:18020/api/find-service" -Method POST -Body $body -ContentType "application/json"
    Write-Host "Resultado: " -NoNewline
    $response | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "❌ ERRO: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Teste 4: Classificar Intenção - Segunda Via Fatura
Write-Host "✅ Teste 4: Classificar 'quero meu boleto'" -ForegroundColor Yellow
try {
    $body = @{
        intent = "quero meu boleto"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "http://localhost:18020/api/find-service" -Method POST -Body $body -ContentType "application/json"
    Write-Host "Resultado: " -NoNewline
    $response | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "❌ ERRO: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Teste 5: Intenção Ambígua
Write-Host "✅ Teste 5: Classificar intenção ambígua 'oi'" -ForegroundColor Yellow
try {
    $body = @{
        intent = "oi"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "http://localhost:18020/api/find-service" -Method POST -Body $body -ContentType "application/json"
    Write-Host "Resultado: " -NoNewline
    $response | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "❌ ERRO: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

Write-Host "🏁 Testes finalizados!" -ForegroundColor Green
