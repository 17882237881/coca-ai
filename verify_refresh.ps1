$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/users/login" -Method Post -ContentType "application/json" -Body '{"email": "test@example.com", "password": "password123"}'
Write-Host "Login Result Code: $($loginResponse.code)"

$refreshToken = $loginResponse.data.refresh_token
if (-not $refreshToken) {
    Write-Error "Failed to get refresh token"
    exit 1
}
Write-Host "Got Refresh Token: $refreshToken"

$refreshBody = @{
    refresh_token = $refreshToken
} | ConvertTo-Json

$refreshResponse = Invoke-RestMethod -Uri "http://localhost:8080/users/refresh_token" -Method Post -ContentType "application/json" -Body $refreshBody
Write-Host "Refresh Result Code: $($refreshResponse.code)"
Write-Host "New Access Token: $($refreshResponse.data.access_token)"
Write-Host "New Refresh Token: $($refreshResponse.data.refresh_token)"

if ($refreshResponse.code -eq 200) {
    Write-Host "Verification PASSED"
} else {
    Write-Host "Verification FAILED"
}
