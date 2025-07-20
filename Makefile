COVERAGE_HTML=coverage.html
COVERPROFILE_TMP=coverprofile.tmp

test:
	go test -json ./... -coverprofile coverprofile_.tmp -coverpkg=./... ; \
    grep -v -e 'mocks.go' -e 'mock.go' -e 'docs.go' -e '_easyjson.go' -e 'gen_sql.go' coverprofile_.tmp > coverprofile.tmp ; \
    rm coverprofile_.tmp ; \
	go tool cover -html ${COVERPROFILE_TMP} -o  $(COVERAGE_HTML); \
    go tool cover -func ${COVERPROFILE_TMP}

view-coverage:
	open $(COVERAGE_HTML)

clean:
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML) ${COVERPROFILE_TMP} 