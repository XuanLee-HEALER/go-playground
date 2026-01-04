#!/usr/bin/env fish
# Generate background traffic for the demo service.
set base_url "http://localhost:8080"
set min_sleep_ms 100
set max_sleep_ms 400
set routes "/ping" "/work" "/slow" "/error"

while true
    set idx (random 1 (count $routes))
    set route $routes[$idx]
    curl -sSf "$base_url$route" > /dev/null
    set sleep_ms (random $min_sleep_ms $max_sleep_ms)
    sleep (math "$sleep_ms / 1000")
end
