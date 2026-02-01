$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/users/login" -Method Post -ContentType "application/json" -Body '{"email": "test@example.com", "password": "password123"}'
Write-Host "Login Result Code: $($loginResponse.code)"
$accessToken = $loginResponse.data.access_token

if (-not $accessToken) {
    Write-Error "Failed to get access token"
    exit 1
}

# 1. Access Logout
$logoutResponse = Invoke-RestMethod -Uri "http://localhost:8080/users/logout" -Method Post -ContentType "application/json" -Headers @{Authorization = "Bearer $accessToken" }
Write-Host "Logout Result Code: $($logoutResponse.code)"

# 2. Verify Access Token Invalid (Expect 401)
try {
    Invoke-RestMethod -Uri "http://localhost:8080/users/logout" -Method Post -Headers @{Authorization = "Bearer $accessToken" }
    Write-Host "Access Token Verification Failed: Should be 401"
}
catch {
    $code = $_.Exception.Response.StatusCode.value__
    if ($code -eq 401) {
        Write-Host "Access Token Verification PASSED: Token is blacklisted (401)"
    }
    else {
        Write-Host "Access Token Verification Failed: Got $code"
    }
}
