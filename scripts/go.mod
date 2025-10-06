module socialpredict-scripts

go 1.21

require (
	github.com/joho/godotenv v1.5.1
)

// Use local socialpredict backend module
replace socialpredict => ../backend

require socialpredict v0.0.0-00010101000000-000000000000