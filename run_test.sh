printf "Running: 'file_sorter -h'\n----\n"
go run main.go -h

printf "\n----\nRunning: 'file_sorter -v'\n----\n"
go run main.go -v

printf "\n----\nRunning: 'file_sorter -i test -o test-dest -dr -m'\n----\n"
go run main.go -i test -o test-dest -dr -m

printf "\n----\nRunning: 'file_sorter -i test -o test-dest'\n----\n"
go run main.go -i test -o test-dest
