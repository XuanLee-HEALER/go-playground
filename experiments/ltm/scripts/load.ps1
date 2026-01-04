# Generate background traffic for the demo service.
param(
  [string]$BaseUrl = "http://localhost:8080",
  [int]$MinSleepMs = 100,
  [int]$MaxSleepMs = 400
)

$routes = @("/ping", "/work", "/slow", "/error")

while ($true) {
  $route = Get-Random -InputObject $routes
  try {
    Invoke-WebRequest -Uri ($BaseUrl + $route) -UseBasicParsing | Out-Null
  } catch {
    # Ignore transient errors while the stack is starting.
  }
  Start-Sleep -Milliseconds (Get-Random -Minimum $MinSleepMs -Maximum $MaxSleepMs)
}
