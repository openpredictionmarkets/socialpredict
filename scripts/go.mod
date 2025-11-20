module socialpredict-scripts

go 1.25

require github.com/joho/godotenv v1.5.1

// Use local socialpredict backend module
replace socialpredict => ../backend

require (
	gorm.io/gorm v1.25.12
	socialpredict v0.0.0-00010101000000-000000000000
)

require (
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/brianvoe/gofakeit v3.18.0+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/microcosm-cc/bluemonday v1.0.27 // indirect
	github.com/yuin/goldmark v1.7.13 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/postgres v1.5.9 // indirect
)
