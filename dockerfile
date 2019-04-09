FROM golang:1.11

RUN go get github.com/jmoiron/sqlx && go get github.com/mattn/go-sqlite3 && go get github.com/mozillazg/go-pinyin && go get github.com/rs/xid