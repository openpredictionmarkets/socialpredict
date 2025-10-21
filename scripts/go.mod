module socialpredict-scripts

go 1.23.1

require github.com/joho/godotenv v1.5.1

// Use local socialpredict backend module
replace socialpredict => ../backend

require (
	gorm.io/gorm v1.25.12
	socialpredict v0.0.0-00010101000000-000000000000
)

require (
	github.com/brianvoe/gofakeit v3.18.0+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/postgres v1.5.9 // indirect
)
