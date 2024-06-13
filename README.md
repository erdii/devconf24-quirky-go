# quirky go

slides for my devconf 2024 talk



command | description
--- | ---
`go run ./cmd/server` | run dev server with live reloading
`RENDER=true go run ./cmd/server` | render presentation to index.html
`rsync --delete-before -r -v index.html static /path/to/where/you/want/this/on/a/webpage && rm index.html` | sync all presentation files to output folder
